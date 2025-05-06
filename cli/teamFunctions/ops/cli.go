package ops

import (
	"fmt"
	"github.com/cloudfoundry/go-cfclient/v3/client"
	"goli-cli/cli/teamFunctions/shared"
	"goli-cli/entities"
	. "goli-cli/types"
	"goli-cli/utils"
	"goli-cli/utils/outputUtils"
)

const (
	RunSAA                  = "Run SAA"
	GetStatusOfSAAJob       = "Get status of SAA job"
	RunQueryOnAllLandscapes = "Run query on all landscapes"
	CheckCert               = "Check certificate"
	Back                    = "Back"
)

func OpsCli(cf *client.Client, apps *map[string]AppData, instances *map[string][]*entities.Instance, names *[]string) {

	teamOptions := []string{RunSAA, GetStatusOfSAAJob, RunQueryOnAllLandscapes, CheckCert, Back}
	options := shared.GetAllOptions(&teamOptions)
	var err error
	var option string
	for {
		option, _ = utils.ListAndSelectItem(options, "select an option:", false)
		if option == Back {
			return
		}
		err = executeOption(cf, option, apps, instances, names)
		if err != nil {
			outputUtils.PrintErrorMessage("An error occurred:", err.Error())
			err = nil
		}
	}
}

func executeOption(cf *client.Client, option string, apps *map[string]AppData, instances *map[string][]*entities.Instance, names *[]string) error {
	var err error

	isExecuted, err := shared.ExecuteOption(cf, option, instances)
	if isExecuted {
		return err
	}

	switch option {
	case RunSAA:
		fmt.Println("Running SAA")
		err = RunSAAFun(cf, instances, apps, "", "")
	case GetStatusOfSAAJob:
		fmt.Println("Getting status of SAA job")
		err = GetStatusOfSAAJobFun(cf, instances, "", "")
	case RunQueryOnAllLandscapes:
		fmt.Println("Running query on all landscapes")
		err = RunQueryAllFun(instances, "", "")
	case CheckCert:
		fmt.Println("Checking certificate")
		err = CheckCertFunc(cf, apps, "", "")
	}
	return err
}
