package teamFunctionsUtils

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/fatih/color"
	"github.com/olekukonko/tablewriter"
	"goli-cli/utils"
	"goli-cli/utils/outputUtils"
	"io"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
)

type SaaStatus struct {
	JobID         string `json:"job_id"`
	TenantID      string `json:"tenant_id"`
	Status        string `json:"status"`
	TotalEntities int    `json:"total_entities"`
	InstanceID    string `json:"instance_id"`
	ProductName   string `json:"product_name"`
}
type ByStatusAndTenantID []SaaStatus

func (a ByStatusAndTenantID) Len() int      { return len(a) }
func (a ByStatusAndTenantID) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a ByStatusAndTenantID) Less(i, j int) bool {
	// Define the order of statuses
	statusOrder := map[string]int{
		"NOT_PROCESSED": 1,
		"SUCCESS":       2,
		"FAILED":        3,
		"RUNNING":       4,
	}

	if statusOrder[a[i].Status] != statusOrder[a[j].Status] {
		return statusOrder[a[i].Status] < statusOrder[a[j].Status]
	}
	return a[i].TenantID < a[j].TenantID
}

func GetPayloadFile(path string) (map[string][]string, error) {
	file, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var payload map[string][]string
	err = json.Unmarshal(file, &payload)
	return payload, err
}

func GetPayloadUser(chunks int) (payload map[string][]string) {
	payload = map[string][]string{
		"tenants":      {},
		"contexts":     {},
		"entityTypes":  {},
		"productNames": {},
	}
	if chunks == 0 {
		payload["tenants"] = getArrayFromUser("enter the tenants: (in the format: 'tenant1,tenant2,tenant3', default: 'All')")
	}
	payload["entityTypes"] = getArrayFromUser("enter the entityTypes: (in the format: 'role,provider,page', default: 'All'")
	payload["productNames"] = getArrayFromUser("enter the productNames (in the format: 'LAUNCHPAD,WORKZONE', default: 'LAUNCHPAD'):")
	return payload
}
func getArrayFromUser(prompt string) []string {
	value := utils.StringPrompt(prompt)
	if value == "" {
		return []string{}
	}
	return strings.Split(value, ",")
}

func GetJobStatus(token, portalUrl, jobId string) ([]byte, error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", portalUrl+"/cdm_store_service/events/replay/"+jobId, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("x-attribute-tenant-id", "test")
	req.Header.Set("x-attribute-instance-id", "test")

	resp, err := client.Do(req)

	if err != nil {
		fmt.Println("HTTP call error:", err)
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, errors.New(fmt.Sprintf("HTTP error: %d - %s\n", resp.StatusCode, resp.Status))
	}
	resRaw, err := io.ReadAll(resp.Body)
	return resRaw, err
}

func PrintJobStatus(status [][]string) {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Index", "Tenant Id", "Status", "Number of entities", "Product Name"})

	for _, v := range status {
		table.Append(v)
	}
	table.SetCenterSeparator("")
	table.SetColumnSeparator("")
	table.SetAlignment(tablewriter.ALIGN_CENTER)
	table.SetColumnColor(tablewriter.Colors{},
		tablewriter.Colors{tablewriter.Bold, tablewriter.FgHiYellowColor},
		tablewriter.Colors{tablewriter.Bold},
		tablewriter.Colors{tablewriter.Bold, tablewriter.FgHiGreenColor},
		tablewriter.Colors{tablewriter.Bold})

	table.Render()
}

func FormatJobStatus(resRaw []byte) ([][]string, error) {
	var jobStatus []SaaStatus
	var parsedStatus [][]string
	err := json.Unmarshal(resRaw, &jobStatus)
	if err != nil {
		return nil, err
	}
	sort.Sort(ByStatusAndTenantID(jobStatus))

	for index, value := range jobStatus {
		var stat string
		if value.Status == "SUCCESS" {
			stat = color.HiCyanString(value.Status)
		} else if value.Status == "FAILED" {
			stat = color.HiRedString(value.Status)
		} else if value.Status == "RUNNING" {
			stat = color.HiGreenString(value.Status)
		} else {
			stat = color.HiBlackString(value.Status)
		}
		parsedStatus = append(parsedStatus, []string{strconv.Itoa(index + 1), value.TenantID, stat, fmt.Sprint(value.TotalEntities), value.ProductName})
	}
	return parsedStatus, err
}

func ValidateQuery(query string) bool {
	tempQuery := strings.ToLower(query)
	if !strings.HasPrefix(tempQuery, "select") {
		outputUtils.PrintErrorMessage("query must begin with 'select'")
		return false
	}
	if !strings.Contains(tempQuery, "from") {
		outputUtils.PrintErrorMessage("query must have with 'from'")
		return false
	}
	// validate there is no update/ delete / insert
	if strings.Contains(tempQuery, "update ") || strings.Contains(tempQuery, "delete ") || strings.Contains(tempQuery, "insert ") {
		outputUtils.PrintErrorMessage("query must not have 'update '/'delete '/'insert '")
		return false
	}
	return true
}

func GetAndPrintJobStatus(token string, portalUrl string, jobIds []string) error {
	var parsedStatuses [][]string
	for _, jobId := range jobIds {

		resRaw, err := GetJobStatus(token, portalUrl, jobId)
		if err != nil {
			return err
		}
		parsedStatus, err := FormatJobStatus(resRaw)
		parsedStatuses = append(parsedStatuses, parsedStatus...)
		if err != nil {
			return err
		}
	}
	PrintJobStatus(parsedStatuses)

	return nil
}
