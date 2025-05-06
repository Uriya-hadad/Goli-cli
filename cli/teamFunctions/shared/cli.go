package shared

import (
	"fmt"
	"github.com/cloudfoundry/go-cfclient/v3/client"
	"goli-cli/entities"
	. "goli-cli/types"
	"goli-cli/utils"
	"goli-cli/utils/outputUtils"
)

const (
	RunQuery    = "Run query on the DB"
	ConnectToDB = "Connect to DB"
	Back        = "Back"
)

var sharedOptions = []string{RunQuery, ConnectToDB}

func SharedCli(cf *client.Client, apps *map[string]AppData, instances *map[string][]*entities.Instance, names *[]string) {
	var err error
	teamOptions := []string{RunQuery, ConnectToDB, Back}

	for {
		option, _ := utils.ListAndSelectItem(teamOptions, "select an option:", false)
		if option == Back {
			return
		}
		_, err = ExecuteOption(cf, option, instances)
		if err != nil {
			outputUtils.PrintErrorMessage("An error occurred:", err.Error())
			err = nil
		}
	}
}

func GetAllOptions(options *[]string) []string {
	return append(sharedOptions, *options...)
}

func ExecuteOption(cf *client.Client, option string, instances *map[string][]*entities.Instance) (bool, error) {
	var err error
	switch option {
	case RunQuery:
		fmt.Println("Running query on DB")
		err = RunQueryFun(cf, instances, "", "", "")
		return true, err
	case ConnectToDB:
		fmt.Println("Connecting to DB")
		err = ConnectToDbCmdFunc(cf, instances, "")
		return true, err
	default:
		return false, nil
	}
}
