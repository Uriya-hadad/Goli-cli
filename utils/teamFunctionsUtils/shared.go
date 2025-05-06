package teamFunctionsUtils

import (
	"encoding/csv"
	"errors"
	"fmt"
	"github.com/cloudfoundry/go-cfclient/v3/client"
	"github.com/fatih/color"
	"goli-cli/cli/instances/instancesTypes"
	"goli-cli/db"
	"goli-cli/entities"
	. "goli-cli/types"
	"goli-cli/utils"
	"goli-cli/utils/instanceUtils"
	"goli-cli/utils/outputUtils"
	"log"
	"os"
	"strings"
	"sync"
)

func ConnectAndPrintQuery(cf *client.Client, pgInstanceRaw *entities.Instance, query, fileName string) error {
	app := instanceUtils.GetFirstStartedApp(cf, pgInstanceRaw.GUID, true)
	if app == nil {
		return errors.New("no container is exist to create the ssh tunnel with")
	}
	pgInstance := instancesTypes.GetManagedInstance("postgresql-db", pgInstanceRaw, cf)
	postgresCred, err := GetPostgresCredentials(cf, nil, "", pgInstance)
	if err != nil {
		return err
	}
	stopChan, err := db.OpenConnectionToService(cf, postgresCred, app.GUID, "postgres", app.Name)
	if err != nil {
		return err
	}
	rows, err := db.RunQuery(postgresCred, query, true)
	stopChan <- os.Interrupt
	if err != nil {
		return err
	}

	if fileName != "" {
		if !strings.HasSuffix(fileName, ".csv") {
			fileName = fileName + ".csv"
		}
		// Save the result to a CSV file
		err = saveToScv(rows, fileName)
	} else {
		// print the query result if no file name is provided
		db.PrintQueryResult(rows)

	}
	return err
}

func GetPostgresCredentials(cf *client.Client, instances *map[string][]*entities.Instance, postgresName string, postgresIns ManagedInstance) (*ConnectionInfo, error) {
	if postgresIns == nil {
		for _, instance := range (*instances)["postgresql-db"] {
			if instance.Name == postgresName {
				postgresIns = instancesTypes.GetManagedInstance("postgresql-db", instance, cf)
				break
			}
		}
	}
	if postgresIns == nil {
		return nil, errors.New("postgres instance not found")
	}
	cred, err := postgresIns.GetBoundDetails(cf)
	if err != nil {
		return nil, err
	}
	connectionInfo, err := db.GetPostgresConnectionInfo(cred)
	if err != nil {
		return nil, err
	}
	return connectionInfo, err
}

func ConnectToDB(cf *client.Client, instances *map[string][]*entities.Instance, dbRawInstance *entities.Instance) error {
	ws := sync.WaitGroup{}
	var err error
	var postgresCred *ConnectionInfo
	var app *CFAppData

	ws.Add(2)
	go func() {
		defer ws.Done()
		postgresCred, err = GetPostgresCredentials(cf, instances, dbRawInstance.Name, nil)
	}()
	go func() {
		defer ws.Done()
		app = instanceUtils.GetFirstStartedApp(cf, dbRawInstance.GUID, true)
	}()

	ws.Wait()
	if err != nil {
		return err
	}
	fmt.Println("Opening SSH tunnel to", color.HiCyanString(app.Name))

	if app == nil {
		return errors.New("no container is exist to create the ssh tunnel with")
	}
	err = db.OpenPostgresConnection(cf, postgresCred, app.GUID, app.Name)
	return err
}

func GetPostgresInstance(dbName string, instances *map[string][]*entities.Instance) (pgInstanceRaw *entities.Instance, err error) {
	var postgresNames []string
	if dbName != "" {
		for _, instance := range (*instances)["postgresql-db"] {
			if instance.Name == dbName {
				pgInstanceRaw = instance
			}
		}
		if pgInstanceRaw == nil {
			return nil, errors.New("invalid db name provided")
		}
	} else {
		postgresServices, ok := (*instances)["postgresql-db"]
		if !ok {
			return nil, errors.New("no postgresql service found")
		}
		for _, instance := range postgresServices {
			postgresNames = append(postgresNames, instance.Name)
		}
		_, dbNameAsInt := utils.ListAndSelectItem(postgresNames, "Select the postgres name:", true)
		pgInstanceRaw = (*instances)["postgresql-db"][dbNameAsInt]
	}
	return pgInstanceRaw, nil
}

func saveToScv(rows [][]string, fileName string) error {
	// Create CSV file
	file, err := os.Create(fileName)
	if err != nil {
		return errors.New("error creating file")
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write header
	if err := writer.Write(rows[0]); err != nil {
		log.Fatalf("Failed to write header: %v", err)
	}

	// Write rows3
	for _, row := range rows[1:] {
		if err := writer.Write(row); err != nil {
			return errors.New("error writing row to CSV")
		}
	}
	fileLoc, _ := os.Getwd()
	outputUtils.PrintSuccessMessage("Query result saved to: " + fileLoc + "/" + fileName)
	return nil
}
