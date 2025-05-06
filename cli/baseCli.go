package cli

import (
	"fmt"
	"github.com/cloudfoundry/go-cfclient/v3/client"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"goli-cli/cli/applications"
	"goli-cli/cli/changeTarget"
	"goli-cli/cli/instances"
	"goli-cli/cli/teamFunctions"
	"goli-cli/entities"
	. "goli-cli/types"
	"goli-cli/utils"
	"goli-cli/utils/outputUtils"
	"goli-cli/utils/setUpUtils"
	"os"
	"sync"
)

var (
	landscapesData                                          LandscapeData
	landscapes                                              Landscape
	updateDataLock, updateLandLock, appsLock, instancesLock sync.WaitGroup
	currentUser                                             *CfUser
	selectedSpace                                           *CfSpace
	selectedOrg                                             *CfOrg
	instancesByOffer                                        *map[string][]*entities.Instance
	offerNames                                              *[]string
	apps                                                    *map[string]AppData
	selectedTarget, domain                                  string
	cf                                                      *client.Client
)

var baseCmd = &cobra.Command{
	Use:   "goli",
	Short: "Cloud Foundry CLI tool for managing applications, instances, and team-specific operations.",
	Long: `The goli command serves as the entry point to manage Cloud Foundry resources and interact with specific team features.
Running the goli command provides you with an interactive menu where you can:

Applications
Manage and interact with applications within the current space.

Instances
View and manipulate service instances.

Custom Team Options
Access team-specific functionality based on your role, such as running queries or triggering specific actions.

Change Targeted Org or Space
Switch between different Cloud Foundry organizations or spaces to manage resources across environments`,
	Version: GetVersion(),
	Run: func(cmd *cobra.Command, args []string) {
		cli(cf)
	},
}

func Execute() {
	baseCmd.CompletionOptions.HiddenDefaultCmd = true
	err := baseCmd.Execute()
	if err != nil {
		outputUtils.Panic(err.Error())
	}
}

func init() {
	var err error
	var isValid bool
	utils.SetTime()
	setUpUtils.DetermineCliFolder()
	currentUser, isValid, err = setUpUtils.LoadConfig()
	if err != nil {
		outputUtils.Panic("Error while validating user", err.Error())
	}
	if !isValid {
		outputUtils.Panic("Please validate your account before using the CLI")
	}
	cf = setUpUtils.CreateCfClient()

	setUpUtils.GetLandscapesFile(&landscapes)
	selectedTarget = utils.ExtractRegion(cf.Config.ApiURL(""))
	domain = utils.ExtractDomain(cf.Config.ApiURL(""))

	selectedSpace, selectedOrg, _ = utils.GetOrgAndSpaceFromConfig()
	if selectedSpace == nil && selectedSpace.Name == "" {
		if !(len(os.Args) > 1 && os.Args[1] == "__complete") {
			fmt.Println("Welcome to the Cloud Foundry CLI\n")
		}
		selectedSpace, selectedOrg, err = changeTarget.ChangeTarget(cf, &landscapes, &updateLandLock, selectedTarget, nil, nil, "", "")
		if err != nil {
			outputUtils.Panic(err.Error())
		}
	} else {
		if !(len(os.Args) > 1 && (os.Args[1] == "__complete" || os.Args[1] == "-v" || os.Args[1] == "--version")) {
			fmt.Println("Welcome to the Cloud Foundry CLI\n")
			fmt.Println("Running the Cli in org:", color.HiCyanString(selectedOrg.Name), "and space:", color.HiCyanString(selectedSpace.Name))
		}
	}
	go setUpUtils.CheckAuthorizations(cf, selectedSpace.GUID, currentUser.Email)
	setUpUtils.GetLandscapesDataFile(&landscapesData)

	GetAndUpdateLandscape(cf)

	baseCmd.AddCommand(
		applications.NewCmd(cf, &apps, &updateDataLock, &appsLock),           // applications command with its subcommands
		instances.NewCmd(cf, &instancesByOffer, &offerNames, &instancesLock), // instances command with its subcommands
		teamFunctions.NewCmd(cf, currentUser.Role, &apps, &instancesByOffer, &offerNames, &appsLock, &instancesLock, &updateDataLock),
		changeTarget.NewCmd(cf, &landscapes, &updateLandLock, selectedTarget, selectedOrg, selectedSpace),
	)
}

func cli(cf *client.Client) {

	const (
		Applications  = "Applications"
		Instances     = "Instances"
		TeamFunctions = "Custom team options"
		ChangeTarget  = "Change targeted org or space"
		Exit          = "Exit"
	)
	options := []string{Applications, Instances, TeamFunctions, ChangeTarget, Exit}
	for {
		option, _ := utils.ListAndSelectItem(options, "Select an option:", false)
		var err error

		switch option {
		case Applications:
			fmt.Println("applications")
			appsLock.Wait()
			if len(*apps) == 0 {
				fmt.Println("No apps found...")
				break
			}
			applications.ApplicationCli(cf, apps, &updateDataLock, "")
		case Instances:
			fmt.Println("instances")
			instancesLock.Wait()
			if len(*offerNames) == 0 {
				fmt.Println("No instances found...")
				break
			}
			instances.InstanceCli(cf, instancesByOffer, offerNames, "")
		case TeamFunctions:
			appsLock.Wait()
			instancesLock.Wait()
			fmt.Println("Team functions of team:", currentUser.Role)
			teamFunctions.TeamFeaturesCli(cf, currentUser.Role, apps, instancesByOffer, offerNames, &updateDataLock)
		case ChangeTarget:
			selectedSpace, selectedOrg, err = changeTarget.ChangeTarget(cf, &landscapes, &updateLandLock, selectedTarget, nil, nil, "", "")
			GetAndUpdateLandscape(cf)
		case Exit:
			fmt.Println("Exiting")
			setUpUtils.UpdateCli()
			os.Exit(0)
		}
		if err != nil {
			fmt.Println("An error occurred:", err)
			err = nil
		}
		fmt.Println("Running the Cli in org:", color.HiCyanString(selectedOrg.Name), "and space:", color.HiCyanString(selectedSpace.Name))
	}
}
func GetAndUpdateLandscape(cf *client.Client) {
	var mutex = &sync.Mutex{}
	updateDataLock.Add(1)
	updateLandLock.Add(1)
	appsLock.Add(1)
	instancesLock.Add(1)
	go func() {
		defer updateDataLock.Done()
		setUpUtils.UpdateLandscapesData(cf, &landscapesData, mutex, selectedTarget, domain, selectedOrg, selectedSpace)
	}()
	go func() {
		// for change target command
		defer updateLandLock.Done()
		setUpUtils.UpdateLandscapes(cf, &landscapes, selectedTarget)
	}()
	mutex.Lock()
	target := landscapesData[selectedTarget]
	if target == nil {
		mutex.Unlock()
		updateDataLock.Wait()
		mutex.Lock()
		target = landscapesData[selectedTarget]
	}
	spaces := target[selectedOrg.Name]
	if spaces == nil {
		mutex.Unlock()
		updateDataLock.Wait()
		mutex.Lock()
		spaces = landscapesData[selectedTarget][selectedOrg.Name]
	}
	spaceData := spaces[selectedSpace.Name]
	if spaceData == nil || spaceData.Apps == nil || spaceData.Instances == nil {
		mutex.Unlock()
		updateDataLock.Wait()
		mutex.Lock()
		spaceData = spaces[selectedSpace.Name]
	}
	mutex.Unlock()
	go func() {
		defer appsLock.Done()
		apps = setUpUtils.GetAppsFromLandscape(spaceData, mutex)
	}()
	go func() {
		defer instancesLock.Done()
		instancesByOffer, offerNames = setUpUtils.GetInstancesFromLandscape(spaceData, mutex)
	}()
}

func GetVersion() string {
	// as it is running from the cli folder, we need to go back to the root folder
	setUpUtils.DetermineCliFolder()
	file, err := os.OpenFile("./version.txt", os.O_RDWR, 0644)
	if err != nil {
		return ""
	}
	defer file.Close()
	version := make([]byte, 100)
	_, err = file.Read(version)
	if err != nil {
		return ""
	}
	return string(version)
}
