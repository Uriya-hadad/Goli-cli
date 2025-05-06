package shanyTeam

import (
	"github.com/cloudfoundry/go-cfclient/v3/client"
	"goli-cli/cli/teamFunctions/shared"
	"goli-cli/entities"
	. "goli-cli/types"
	"goli-cli/utils"
	"goli-cli/utils/outputUtils"
)

func ShanyTeamCli(cf *client.Client, apps *map[string]AppData, instances *map[string][]*entities.Instance, names *[]string) {
	const (
		Back = "Back"
	)
	teamOptions := []string{Back}
	options := shared.GetAllOptions(&teamOptions)

	var err error
	for {
		option, _ := utils.ListAndSelectItem(options, "select an option:", false)
		if option == Back {
			return
		}
		err = executeOption(cf, option, instances)
		if err != nil {
			outputUtils.PrintErrorMessage("An error occurred:", err.Error())
			err = nil
		}
	}
}

func executeOption(cf *client.Client, option string, instances *map[string][]*entities.Instance) error {
	var err error

	isExecuted, err := shared.ExecuteOption(cf, option, instances)
	if isExecuted {
		return err
	}

	switch option {
	}
	return err
}
