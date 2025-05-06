package ops

import (
	"context"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"github.com/cloudfoundry/go-cfclient/v3/client"
	"github.com/cloudfoundry/go-cfclient/v3/config"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"goli-cli/db"
	"goli-cli/entities"
	"goli-cli/helpers"
	. "goli-cli/types"
	"goli-cli/utils"
	"goli-cli/utils/outputUtils"
	"goli-cli/utils/teamFunctionsUtils"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

func NewRunSaaCmd(cf *client.Client, apps **map[string]AppData, instances **map[string][]*entities.Instance) *cobra.Command {
	var chunks, payloadPath string

	cmd := &cobra.Command{
		Use:   "run-saa",
		Short: "Trigger the re-sending of SubAccount Activation events through the message queue.",
		Long: `This command triggers the SubAccount Activation (SAA) flow, which re-sends all the activation events for the specified tenants, contexts, and entity types.
It creates a job to send the events again to the message queue.
You can specify different parameters in the payload to control the scope of the activation process, including tenants, contexts, entity types, and product names.

If the 'tenants' list is empty, all tenants will be activated. The 'chunks' flag allows you to split the activation into chunks for more manageable processing.

Usage:
  goli team-features run-saa [OPTIONS]

Options:
  -h, --help  
      Display this help message and exit.

  -p, --payloadPath <path>  
      A path to the JSON file containing the payload for the activation job.
      The payload can include 'tenants', 'contexts', 'entityTypes', and 'productNames'.
      See the detailed description for more information on the required structure.

  -c, --chunks <number>
      The number of chunks to split the activation process into if the 'tenants' list is empty (full SAA).
      This helps distribute the load across multiple activations.

Examples:
  goli team-features run-saa -p "/path/to/payload.json" -c 5
      Trigger the re-sending of activation events for the tenants and contexts defined in the JSON file, splitting the load into 5 chunks.

  goli team-features run-saa
      Open interactive mode to input the payload and trigger the re-sending of activation events.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return RunSAAFun(cf, *instances, *apps, payloadPath, chunks)
		},
	}

	cmd.Flags().StringVarP(&chunks, "chunks", "c", "", "number of chunks the SAA will be split by")
	cmd.Flags().StringVarP(&payloadPath, "payloadPath", "p", "", "path to the payload of the SAA")

	cmd.SetHelpTemplate(cmd.Long)

	return cmd
}
func NewSaaStatusCmd(cf *client.Client, instances **map[string][]*entities.Instance) *cobra.Command {
	var id, interval string

	cmd := &cobra.Command{
		Use:   "status-saa",
		Short: "Display the status of SubAccount Activation events for each tenant.",
		Long: `This command retrieves the status of SubAccount Activation events and displays a table showing which tenants have had their events sent successfully.
If a job ID is specified, the status will be tracked for that particular job.
If no job ID is provided, an interactive mode will open for the user to enter the job ID.
If the job ID is still not specified, the command will attempt to retrieve it from the 'job-id.txt' file (saved by the 'run-saa' command).

If the 'interval' flag is not provided, the status will be fetched once. If 'interval' is provided, the status will be updated periodically based on the specified interval.

Usage:
  goli team-features status-saa [OPTIONS]

Options:
  -h, --help
      Display this help message and exit.

  -i, --id <job-id>
      The job ID for the SubAccount Activation process. If specified, the command will display the status for that particular job.

  -t, --interval <seconds>
      The interval (in seconds) to wait between status checks. If provided, the command will periodically fetch the status at the specified interval. If not provided, the status will be fetched once.

Examples:
  goli team-features status-saa -id "12345"
      Retrieve the status of the SubAccount Activation events for job ID "12345".

  goli team-features status-saa -id "12345" -interval "10"
      Retrieve the status of the SubAccount Activation events for job ID "12345" and update the status every 10 seconds.

  goli team-features status-saa
      Start interactive mode to enter a job ID for the status retrieval.

  goli team-features status-saa -interval "5"
      Retrieve the status of SubAccount Activation events and update the status every 5 seconds.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return GetStatusOfSAAJobFun(cf, *instances, id, interval)
		},
	}

	cmd.Flags().StringVarP(&id, "id", "i", "", "job Id")
	cmd.Flags().StringVarP(&interval, "interval", "t", "", "interval for pulling the status")

	cmd.SetHelpTemplate(cmd.Long)

	return cmd
}

func NewRunQueryAllCmd(instances **map[string][]*entities.Instance) *cobra.Command {
	var query, dbName string

	cmd := &cobra.Command{
		Use:   "run-query-all",
		Short: "Run a query on all data centers using credentials from config.json.",
		Long: `This command runs the specified SQL query on every data center (DC), using the credentials stored in 'config.json' for a technical user.
It allows for running the same query across multiple DCs simultaneously without needing to specify individual database connections for each one.

Usage:
  goli team-features run-query-all [OPTIONS]

Options:
  -q, --query <query>  
      The SQL query you wish to execute across all data centers.
      This is a required flag and must be specified.

  -d, --db <database>  
      The name of the database where the query will be executed.
      If not specified, interactive mode will open to choose the database.

  -h, --help  
      Display this help message and exit.

Examples:
  goli team-features run-query-all -q "SELECT * FROM users" -d "my_database"  
      Execute the query "SELECT * FROM users" across all data centers for the "my_database" database.

  goli team-features run-query-all  
      Open interactive mode to input the query and select the database to run the query across all data centers.`,
		Aliases: []string{"query-all"},
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return RunQueryAllFun(*instances, query, dbName)
		},
	}

	cmd.Flags().StringVarP(&query, "query", "q", "", "query to run on the DB")
	cmd.Flags().StringVarP(&dbName, "db", "d", "", "the DB on which the query will be run on")

	cmd.SetHelpTemplate(cmd.Long)

	return cmd
}

func NewCheckCertCmd(cf *client.Client, apps **map[string]AppData, instances **map[string][]*entities.Instance) *cobra.Command {
	var appName, InstanceName string

	cmd := &cobra.Command{
		Use:   "show-cert-details",
		Short: "Show certificate details for a bound instance.",
		Long: `Retrieve and display certificate details for a specific bound instance of an application.
This includes the certificate issuer, subject, validity dates, and expiration status.
Use this command to verify certificates associated with specific app-instance bindings.

Usage:
  goli team-features show-cert-details [OPTIONS]

Aliases:
  show-cert-details, scd

Options:
  -a, --app <app-name>  
      The name of the application.
      If not specified, an interactive mode will start to choose an application.

  -i, --instance <instance-name>  
      The name of the instance. 
      If not specified, an interactive mode will start to choose an instance.

  -h, --help
     Display this help message and exit.

Examples:
  goli team-features show-cert-details -a my-app -i my-instance  
      Display the certificate details for the instance "my-instance" bound to the application "my-app."

  goli team-features show-cert-details  
      Start an interactive mode to select an application and an instance, then display the certificate details.`,
		Aliases: []string{"scd"},
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return CheckCertFunc(cf, *apps, appName, InstanceName)
		},
	}
	cmd.SetHelpTemplate(cmd.Long)

	cmd.Flags().StringVarP(&appName, "app", "a", "", "the app for which the certificate will be checked")
	cmd.Flags().StringVarP(&InstanceName, "instance", "i", "", "the instance for which the certificate will be checked")

	return cmd
}

func CheckCertFunc(cf *client.Client, appsList *map[string]AppData, appName, instanceName string) error {
	app := helpers.GetApp(appsList, appName)
	if app == nil {
		return errors.New("app not found")
	}
	var instances = make(map[string][]*entities.Instance)
	VcapServices, err := app.GetVcapServices(cf)
	if err != nil {
		return err
	}
	var offersNames []string
	for offerName, services := range *VcapServices {
		offersNames = append(offersNames, offerName)
		for _, service := range services {
			(instances)[offerName] = append((instances)[offerName], service)
		}
	}

	instance := helpers.GetInstance(cf, &instances, "", instanceName)
	if instance == nil {
		return errors.New("instance not found")
	}

	cred, err := instance.GetCredentials(cf)
	if err != nil {
		return errors.New("failed to get credentials")
	}
	if cred["certificate"] == nil {
		return errors.New("binding is not of type x509")
	}
	block, _ := pem.Decode([]byte(cred["certificate"].(string)))
	if block == nil {
		return errors.New("failed to decode PEM block containing certificate")
	}

	// Parse the certificate
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return errors.New("failed to parse certificate")
	}

	// Print certificate information
	outputUtils.PrintInfoMessage("Certificate Information:")
	outputUtils.PrintInfoMessage(fmt.Sprintf("  Subject: %s", cert.Subject))
	outputUtils.PrintInfoMessage(fmt.Sprintf("  Issuer: %s", cert.Issuer))
	outputUtils.PrintInfoMessage(fmt.Sprintf("  Not Before: %s", cert.NotBefore.Format(time.RFC1123)))
	outputUtils.PrintInfoMessage(fmt.Sprintf("  Not After: %s", cert.NotAfter.Format(time.RFC1123)))
	outputUtils.PrintInfoMessage(fmt.Sprintf("  DNS Names: %v", cert.DNSNames))
	outputUtils.PrintInfoMessage(fmt.Sprintf("  Key Usage: %v", cert.KeyUsage))
	outputUtils.PrintInfoMessage(fmt.Sprintf("  ExtKeyUsage: %v", cert.ExtKeyUsage))
	outputUtils.PrintInfoMessage(fmt.Sprintf("  Serial Number: %s", cert.SerialNumber))
	outputUtils.PrintInfoMessage(fmt.Sprintf("  Signature Algorithm: %s", cert.SignatureAlgorithm))
	outputUtils.PrintInfoMessage(fmt.Sprintf("  Public Key Algorithm: %s", cert.PublicKeyAlgorithm))

	return nil
}

func RunSAAFun(cf *client.Client, instances *map[string][]*entities.Instance, apps *map[string]AppData, payloadPath, chunksValue string) error {
	var payload map[string][]string
	var chunks int
	var postgresCred *ConnectionInfo
	var token, portalUrl string
	var err error

	postgresName := "portal-postgresql-db-dt"
	wg := sync.WaitGroup{}
	wg.Add(3)
	go func() {
		defer wg.Done()
		token, err = helpers.GenerateClientToken(instances, cf)
		if err != nil {
			outputUtils.PrintErrorMessage("An error occurred:", err.Error())
			return

		}
	}()
	go func() {
		defer wg.Done()
		portalUrl, err = helpers.GetPortalUrl(instances, cf)
		if err != nil {
			outputUtils.PrintErrorMessage("An error occurred:", err.Error())
			return
		}
	}()
	go func() {
		defer wg.Done()
		postgresCred, err = teamFunctionsUtils.GetPostgresCredentials(cf, instances, postgresName, nil)
		if err != nil {
			outputUtils.PrintErrorMessage("An error occurred:", err.Error())
			return
		}
	}()
	if chunksValue != "" {
		chunks, err = strconv.Atoi(chunksValue)
		if err != nil {
			return errors.New("invalid chunks value")
		}
	}

	if payloadPath != "" {
		payload, err = teamFunctionsUtils.GetPayloadFile(payloadPath)
		if err != nil {
			return err
		}
	} else {
		fmt.Println("No payload path provided, please enter the tenants, contexts and entityTypes manually (for array values, use comma separated values[','])")
		payload = teamFunctionsUtils.GetPayloadUser(chunks)
	}
	if len(payload["tenants"]) == 0 {
		fmt.Println("No tenants provided, full Subaccount action will be performed")
		if !utils.PresentSecurityQuestion() {
			return nil
		}
	}
	for key, value := range payload {
		fmt.Println(key, ":", value)
	}
	wg.Wait()

	if token == "" || portalUrl == "" {
		return nil
	}

	if chunks > 0 {
		outputUtils.PrintInfoMessage("running SAA by", chunksValue, "chunks.")
		fmt.Println("getting all of the tenants for which we need to run the SAA")
		cdmStoreApp := helpers.GetApp(apps, "portal-cf-cdm-store-service")
		if cdmStoreApp == nil {
			return errors.New("cdm-store-service app not found")
		}

		stopChan, err := db.OpenConnectionToService(cf, postgresCred, cdmStoreApp.GUID, "postgres", cdmStoreApp.Name)
		if err != nil {
			return err
		}
		rows, err := db.RunQuery(postgresCred, `select distinct("CDM_ENTITIES"."identityZoneId") from "cdm"."CDM_ENTITIES"`, false)

		stopChan <- os.Interrupt
		if err != nil {
			return err
		}
		var tenantsFromDb, jobIds []string
		for _, row := range rows {
			tenantsFromDb = append(tenantsFromDb, row[0])
		}
		for i := 0; i < len(rows); i += chunks {
			payload["tenants"] = tenantsFromDb[i:min(i+chunks, len(rows))]
			jobId, err := runSAAJobFunc(payload, token, portalUrl)
			jobIds = append(jobIds, jobId)
			if err != nil {
				return err
			}
		}
		str := strings.Join(jobIds, "\n")
		err = os.WriteFile("job-id.txt", []byte(str), 0644)
	} else {
		jobId, err := runSAAJobFunc(payload, token, portalUrl)
		if err != nil {
			return err
		}
		err = os.WriteFile("job-id.txt", []byte(jobId), 0644)
	}
	return err

}

func GetStatusOfSAAJobFun(cf *client.Client, instances *map[string][]*entities.Instance, id, interval string) error {
	var token, jobId, portalUrl string
	var jobIds []string
	var jobIdRaw []byte
	var err error
	wg := sync.WaitGroup{}
	wg.Add(2)
	go func() {
		defer wg.Done()
		token, err = helpers.GenerateClientToken(instances, cf)
		if err != nil {
			outputUtils.PrintErrorMessage("An error occurred:", err.Error())
			return

		}
	}()
	go func() {
		defer wg.Done()
		portalUrl, err = helpers.GetPortalUrl(instances, cf)
		if err != nil {
			outputUtils.PrintErrorMessage("An error occurred:", err.Error())
			return
		}
	}()
	if id != "" {
		jobId = id
	} else {
		jobId = utils.StringPrompt("For which job id?\n(for the last job, press 'Enter'):")
		if jobId == "" {
			jobIdRaw, err = os.ReadFile("job-id.txt")
			if err != nil {
				jobIdRaw = []byte("")
			}
			jobId = string(jobIdRaw)
		}
	}
	if jobId == "" {
		outputUtils.PrintErrorMessage("No job id found")
		return nil
	}
	jobIds = strings.Split(jobId, "\n")
	fmt.Println("JobId:", jobIds)
	wg.Wait()

	if interval != "" {
		// get the status of the job every interval seconds
		var intervalInt int
		intervalInt, err = strconv.Atoi(interval)
		if err != nil {
			return errors.New("invalid job interval value")
		}
		stop := false
		jobs := strings.Join(jobIds, ",")
		fmt.Printf("Getting the status of jobs %s every %s seconds.\nPress on '%s' to stop\n", color.HiCyanString(jobs), color.HiGreenString(interval), color.HiRedString("Enter"))
		go func() {
			for !stop {
				err = teamFunctionsUtils.GetAndPrintJobStatus(token, portalUrl, jobIds)
				if err != nil {
					fmt.Println("error getting job status: ", err)
					break
				}
				if err != nil {
					fmt.Println("error parsing interval: ", err)
					break
				}
				time.Sleep(time.Duration(intervalInt) * time.Second)
			}
		}()
		utils.StopUntilEnter()
		stop = true
	} else {
		err = teamFunctionsUtils.GetAndPrintJobStatus(token, portalUrl, jobIds)
	}
	return err
}

func RunQueryAllFun(instances *map[string][]*entities.Instance, queryInput, db string) error {
	var query, dbName string
	var postgresNames []string
	var landscapesInfo map[string]struct {
		API   string `json:"api"`
		Org   string `json:"org"`
		Space string `json:"space"`
	}
	var dbUserInfo struct {
		DbCred struct {
			Username string `json:"username"`
			Password string `json:"password"`
		} `json:"dbCredentials"`
	}
	userConfRaw, err := os.ReadFile("config.json")
	if err != nil {
		outputUtils.PrintErrorMessage("config file should be exist")
		return err
	}
	err = json.Unmarshal(userConfRaw, &dbUserInfo)
	if dbUserInfo.DbCred.Username == "" || dbUserInfo.DbCred.Password == "" {
		return errors.New("user credentials are missing in config file")
	}
	infoRaw, err := os.ReadFile("resources/landscapesInfo.json")
	if err != nil {
		outputUtils.PrintErrorMessage("landscapesInfo should be exist in the resources folder with all of the landscapes")
		return err
	}
	err = json.Unmarshal(infoRaw, &landscapesInfo)

	if db != "" {
		dbName = db
	} else {
		for _, instance := range (*instances)["postgresql-db"] {
			postgresNames = append(postgresNames, instance.Name)
		}
		_, dbNameAsInt := utils.ListAndSelectItem(postgresNames, "Select the postgres name:", true)
		dbName = (*instances)["postgresql-db"][dbNameAsInt].Name
	}

	if queryInput != "" {
		query = queryInput
	} else {
		query = utils.StringPrompt("Enter the query:")
	}
	if !teamFunctionsUtils.ValidateQuery(query) {
		return errors.New("invalid query - query be a 'select' query and contain 'from' clauses")
	}

	var cf *client.Client
	for _, landscape := range landscapesInfo {
		outputUtils.PrintSuccessMessage("******", landscape.Space, "******")
		cfConf, err := config.New(landscape.API, config.UserPassword(dbUserInfo.DbCred.Username, dbUserInfo.DbCred.Password), config.SkipTLSValidation())
		if err != nil {
			return err
		}
		cf, err = client.New(cfConf)
		if err != nil {
			return err
		}
		space, _ := cf.Spaces.First(context.Background(), &client.SpaceListOptions{
			Names: client.Filter{Values: []string{landscape.Space}},
		})
		pgInstance, err := cf.ServiceInstances.First(context.Background(), &client.ServiceInstanceListOptions{
			Names:      client.Filter{Values: []string{dbName}},
			SpaceGUIDs: client.Filter{Values: []string{space.GUID}},
		})
		if err != nil {
			return err
		}
		tempInstance := &entities.Instance{
			Name: pgInstance.Name,
			GUID: pgInstance.GUID,
		}
		err = teamFunctionsUtils.ConnectAndPrintQuery(cf, tempInstance, query, "")
		if err != nil {
			return err
		}
	}
	return nil
}

func runSAAJobFunc(payload map[string][]string, token string, portalUrl string) (string, error) {

	client := &http.Client{}
	reqBody := map[string]interface{}{"x-attribute-username": "technical"}
	for key, value := range payload {
		reqBody[key] = value
	}
	body, err := json.Marshal(reqBody)
	req, err := http.NewRequest("POST", portalUrl+"/cdm_store_service/events/replay/", strings.NewReader(string(body)))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("x-attribute-tenant-id", "test")
	req.Header.Set("x-attribute-instance-id", "test")

	resp, err := client.Do(req)

	if err != nil {
		fmt.Println("HTTP call error:", err)
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return "", errors.New(fmt.Sprintf("HTTP error: %d - %s\n", resp.StatusCode, resp.Status))
	}

	jobId := resp.Header.Get("Location")
	jobId = strings.Split(jobId, "/")[3]
	fmt.Println("JobId:", jobId)
	return jobId, nil
}
