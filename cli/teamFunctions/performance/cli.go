package performance

import (
	"github.com/cloudfoundry/go-cfclient/v3/client"
	"goli-cli/cli/teamFunctions/shared"
	"goli-cli/entities"
	. "goli-cli/types"
	"goli-cli/utils"
	"goli-cli/utils/outputUtils"
	"sync"
)

const (
	PrintSatusOfApps    = "Print status of apps"
	PrintLogLevelOfApps = "Print log level of all apps"
	Back                = "Back"
)

func PerformanceCli(cf *client.Client, apps *map[string]AppData, instances *map[string][]*entities.Instance, updateLock *sync.WaitGroup) {
	teamOptions := []string{PrintSatusOfApps, PrintLogLevelOfApps, Back}
	options := shared.GetAllOptions(&teamOptions)

	var err error
	for {
		option, _ := utils.ListAndSelectItem(options, "select an option:", false)
		if option == Back {
			return
		}
		err = executeOption(cf, option, apps, instances, updateLock)
		if err != nil {
			outputUtils.PrintErrorMessage("An error occurred:", err.Error())
			err = nil
		}
	}
}

func executeOption(cf *client.Client, option string, apps *map[string]AppData, instances *map[string][]*entities.Instance, updateLock *sync.WaitGroup) error {
	var err error

	isExecuted, err := shared.ExecuteOption(cf, option, instances)
	if isExecuted {
		return err
	}

	switch option {
	case PrintSatusOfApps:
		err = PrintStatusFunc(cf, apps)
	case PrintLogLevelOfApps:
		err = PrintAppsLogLevelFunc(cf, apps, updateLock)
	}
	return err
}
