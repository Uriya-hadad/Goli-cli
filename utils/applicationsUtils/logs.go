package applicationsUtils

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/cloudfoundry/go-cfclient/v3/client"
	"github.com/fatih/color"
	"goli-cli/utils"
	"goli-cli/utils/outputUtils"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type log struct {
	timestamp     string
	method        string
	path          string
	status        string
	correlationID string
}

type logPayload struct {
	Timestamp string `json:"timestamp"`
	Log       struct {
		Payload string `json:"payload"`
		Type    string `json:"type"`
	} `json:"log"`
}

type logReqPayload struct {
	Envelopes struct {
		Batch []logPayload `json:"batch"`
	} `json:"envelopes"`
}

var domain string
var localTime = time.Now()

// Get the location (time zone) of the system
var location = localTime.Location()

func GetCurrentLogs(cf *client.Client, appGUID, appName string, level string) error {
	domain = utils.ExtractDomain(cf.Config.ApiURL(""))
	now := time.Now()
	timestamp := now.UnixNano()
	stop := false

	var err error
	fmt.Printf("Getting logs for app %s click on '%s' to stop\n", color.HiCyanString(appName), color.HiRedString("Enter"))
	go func() {
		for !stop {
			timestamp, err = getAndPrintIncomeLogs(cf, appGUID, timestamp, "", level)
			if err != nil {
				fmt.Println("error getting logs ", err)
				break
			}
		}
	}()
	if err != nil {
		return err
	}
	utils.StopUntilEnter()
	stop = true
	return nil
}

func GetRecentLogs(cf *client.Client, appGUID string, appName string, numOfMin int, level, correlationId string) error {
	var err error
	var timestamp int64

	domain = utils.ExtractDomain(cf.Config.ApiURL(""))
	numberOfMinutes := numOfMin
	if numberOfMinutes == 0 {
		timestamp = -6795364578871345152
		fmt.Printf("Getting earliest logs for app %s\n", color.HiCyanString(appName))
	} else {
		now := time.Now()
		timestamp = now.Add(-time.Duration(numberOfMinutes) * time.Minute).UnixNano()
		fmt.Printf("Getting logs for app %s for the last '%s' minutes\n", color.HiCyanString(appName), color.HiCyanString(strconv.Itoa(numberOfMinutes)))
	}
	current := int64(0)
	for timestamp != current {
		current = timestamp
		timestamp, err = getAndPrintIncomeLogs(cf, appGUID, current, correlationId, level)
		if err != nil {
			return err
		}
	}
	if numberOfMinutes == 0 {
		fmt.Println("This is all of the logs the app has for now")
	} else {
		fmt.Println("No more logs for the last", color.HiCyanString(strconv.Itoa(numberOfMinutes)), "minutes")
	}
	return nil
}

func getAndPrintIncomeLogs(cf *client.Client, appGUID string, currentTimestamp int64, correlationId, level string) (nextTimestamp int64, err error) {
	logs, err := getLogsFromService(cf, appGUID, currentTimestamp)
	if err != nil {
		return 0, errors.New("error getting logs from service " + err.Error())
	}
	if len(*logs) > 0 {
		nextTimestamp, err = getNextTimestamp(logs)
		if err != nil {
			return 0, errors.New("error getting next timestamp " + err.Error())
		}
		err = printLogs(logs, correlationId, level)
		if err != nil {
			return 0, errors.New("error printing logs " + err.Error())
		}

		return nextTimestamp, nil
	}
	return currentTimestamp, nil
}

func getNextTimestamp(logs *[]logPayload) (int64, error) {
	lastLog := (*logs)[len(*logs)-1]
	timestamp, err := strconv.ParseInt(lastLog.Timestamp, 10, 64)
	if err != nil {
		return 0, errors.New("error parsing timestamp " + err.Error())
	}
	return timestamp + 1, nil
}

func getLogsFromService(cf *client.Client, appGUID string, timestamp int64) (logs *[]logPayload, err error) {
	logsReq := &logReqPayload{}
	resp, err := http.NewRequest("GET", fmt.Sprintf("https://log-cache.cf.%s/api/v1/read/%s?envelope_types=LOG&limit=1000&start_time=%d", domain, appGUID, timestamp), nil)
	if err != nil {
		return nil, errors.New("error creating request " + err.Error())
	}
	response, err := cf.ExecuteAuthRequest(resp)
	if err != nil {
		return nil, errors.New("error executing request " + err.Error())
	}
	defer response.Body.Close()
	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, errors.New("error reading response body " + err.Error())
	}
	err = json.Unmarshal(responseBody, logsReq)
	if err != nil {
		return nil, errors.New("error unmarshalling logs " + err.Error())
	}

	return &logsReq.Envelopes.Batch, nil
}

func printLogs(logs *[]logPayload, correlationId, level string) error {
	for _, logData := range *logs {
		payloadBytes, err := base64.StdEncoding.WithPadding(base64.StdPadding).DecodeString(logData.Log.Payload)
		if err != nil {
			return errors.New("error decoding payload " + err.Error())
		}
		payloadAsString := string(payloadBytes)
		if strings.HasPrefix(payloadAsString, "{") && strings.HasSuffix(payloadAsString, "}") {
			//json output
			var logJson struct {
				CorrelationID string `json:"correlation_id"`
				Request       string `json:"request"`
				Method        string `json:"method"`
				Msg           string `json:"msg"`
				Location      string `json:"location"`
				Level         string `json:"level"`
				Timestamp     string `json:"timestamp"`
			}
			err = json.Unmarshal(payloadBytes, &logJson)
			logJson.Timestamp = getTimeOfLog(logData.Timestamp)
			if err != nil {
				fmt.Println(string(payloadBytes))
				continue
			}

			if level != "" && !strings.EqualFold(logJson.Level, level) {
				continue
			}
			if correlationId != "" && logJson.CorrelationID != correlationId {
				continue
			}
			outputUtils.PrintInterface(logJson)
		} else if strings.Contains(payloadAsString, "HTTP/") {
			//request

			// Parse the JSON payload
			logSplited := strings.Split(strings.ReplaceAll(payloadAsString, "\"", ""), " ")

			logCorrelationId := ""
			for i, v := range logSplited {
				if strings.Contains(v, "correlation") {
					logCorrelationId = strings.Split(logSplited[i], ":")[1]
				}
			}

			logFormmated := log{
				timestamp:     getTimeOfLog(logData.Timestamp),
				method:        logSplited[3],
				path:          logSplited[4],
				status:        logSplited[6],
				correlationID: logCorrelationId,
			}
			if level != "" {
				continue
			}
			if correlationId != "" && logFormmated.correlationID != correlationId {
				continue
			}
			outputUtils.PrintInterface(logFormmated)
		} else if logData.Log.Type == "ERR" {
			//error output
			if correlationId != "" {
				continue
			}
			if level != "" {
				continue
			}
			fmt.Println("  " + color.HiRedString(payloadAsString))
		} else {
			//application output
			if correlationId != "" {
				continue
			}
			if level != "" {
				continue
			}
			fmt.Println("  " + payloadAsString)
		}
	}
	return nil
}

func getTimeOfLog(timestampRaw string) (timestamp string) {
	// Parse the time string into a time.Time object
	ns, err := strconv.ParseInt(timestampRaw, 10, 64)
	if err != nil {
		fmt.Println("Error parsing nanosecond timestamp:", err)
		return
	}
	t := time.Unix(0, ns)
	t = t.In(location)

	// Format the time as desired
	timestampRaw = t.Format("2006-01-02 15:04:05")

	return timestampRaw
}
