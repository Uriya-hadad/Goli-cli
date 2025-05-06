package applications

import (
	"context"
	"errors"
	"fmt"
	"github.com/cloudfoundry/go-cfclient/v3/client"
	"github.com/spf13/cobra"
	. "goli-cli/entities"
	. "goli-cli/types"
	"goli-cli/utils"
	"goli-cli/utils/applicationsUtils"
	"goli-cli/utils/outputUtils"
	"strings"
	"sync"
)

const (
	Details             = "Details"
	Restart             = "Restart the app"
	RestartRolling      = "Restart the app --strategy rolling"
	Restage             = "Restage the app"
	RestageRolling      = "Restage the app --strategy rolling"
	ConnectToPostgres   = "Connect to postgres"
	ConnectToRedis      = "Connect to redis"
	ShowEnvs            = "Show envs"
	AddEnv              = "Add env"
	ChangeEnv           = "Change env"
	ShowRecentLogs      = "Show logs --recent"
	ShowLogs            = "Show logs"
	ShowInstances       = "Show bound instances"
	ManipulateInstances = "Manipulate instances"
	EnableSsh           = "Enable ssh"
	ChangeApp           = "Change app"
	Back                = "Return to the previous menu"
)

func NewCmd(cf *client.Client, apps **map[string]AppData, updateLock, appsLock *sync.WaitGroup) *cobra.Command {
	var appsList *map[string]AppData

	cmd := &cobra.Command{
		Use:     "applications [APP_NAME]",
		Aliases: []string{"a"},
		Short:   "Manage and perform actions on applications.",
		Long: `The 'applications' command allows you to interactively select and manage Cloud Foundry applications within your targeted space.
You can either provide the application name as an argument or navigate through an interactive mode to choose an app and perform actions.`,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			appsLock.Wait()
			appsList = *apps
			if len(args) == 1 {
				formatedAppName := args[0]
				if (*appsList)[formatedAppName].Name == "" {
					outputUtils.Panic("application do not exist")
				}
				if !strings.HasPrefix(cmd.Use, "applications") {
					rawApp := (*appsList)[formatedAppName]
					app := NewApp(cf, formatedAppName, rawApp.GUID, updateLock, appsList, false)
					ctx := context.WithValue(cmd.Context(), "app", app)
					cmd.SetContext(ctx)
				}
			}
		},
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) > 1 {
				return errors.New("only one arg is accepted")
			}
			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {
			isRaw, _ := cmd.Flags().GetBool("raw")
			if isRaw {
				appsCompletion(appsList)
				return
			}
			if len(args) == 1 {
				ApplicationCli(cf, appsList, updateLock, args[0])
				return
			}
			ApplicationCli(cf, appsList, updateLock, "")
		},
	}
	cmd.AddCommand(NewDetailsCmd(cf),
		NewRestartCmd(cf),
		NewRestageCmd(cf),
		NewPostgresCmd(cf),
		NewRedisCmd(cf),
		NewCreateEnvCmd(cf),
		NewChangeEnvCmd(cf),
		NewShowEnvsCmd(cf),
		NewShowLogsCmd(cf),
		NewShowBoundInstancesCmd(cf),
		NewManipulateInstanceCmd(cf),
		NewSshCmd(cf))

	// create a raw flag for retuning the raw applications data - for completion
	cmd.Flags().BoolP("raw", "", false, "return all of applications by name")
	cmd.Flags().MarkHidden("raw")

	return cmd
}

func ApplicationCli(cf *client.Client, appsList *map[string]AppData, updateLock *sync.WaitGroup, appName string) {
	var appData AppData
	if appName == "" {
		// full interactive mode
		keys := make([]string, 0, len(*appsList))
		for name, _ := range *appsList {
			keys = append(keys, name)
		}
		appName, _ = utils.ListAndSelectItem(keys, "select an app:", true)
	}
	appData = (*appsList)[appName]
	app := NewApp(cf, appName, appData.GUID, updateLock, appsList, false)
	listOptions(cf, app, appsList, updateLock)
}

func listOptions(cf *client.Client, app *App, appsList *map[string]AppData, updateLock *sync.WaitGroup) {
	var option string
	var err error

	options := []string{Details, Restart, RestartRolling, Restage, RestageRolling, ConnectToPostgres, ConnectToRedis, ShowEnvs, AddEnv, ChangeEnv, ShowLogs, ShowRecentLogs, ShowInstances, ManipulateInstances, EnableSsh, ChangeApp, Back}

	for {
		fmt.Println("selected app: ", app.Name)
		option, _ = utils.ListAndSelectItem(options, "select an option:", false)
		if option == Back {
			return
		} else if option == ChangeApp {
			ApplicationCli(cf, appsList, updateLock, "")
			return
		}
		err = executeOption(cf, app, option)
		if err != nil {
			outputUtils.PrintErrorMessage("An error occurred:", err.Error())
			err = nil
		}
		utils.StringPrompt("Press enter to continue")
	}
}

func executeOption(cf *client.Client, app *App, option string) error {
	var err error
	switch option {
	case Details:
		fmt.Println("getting details")
		err = PrintDetails(cf, app)
	case Restart:
		fmt.Println("restarting the app")
		if !utils.PresentSecurityQuestion() {
			break
		}
		err = RestartApp(cf, app)
	case RestartRolling:
		fmt.Println("restarting the app --strategy rolling")
		if !utils.PresentSecurityQuestion() {
			break
		}
		err = RestartAppRolling(cf, app)
	case Restage:
		fmt.Println("restaging the app")
		if !utils.PresentSecurityQuestion() {
			break
		}
		err = RestageApp(cf, app)
	case RestageRolling:
		fmt.Println("restaging the app --strategy rolling")
		if !utils.PresentSecurityQuestion() {
			break
		}
		err = RestageAppRolling(cf, app)
	case ConnectToPostgres:
		fmt.Println("connecting to postgres...")
		err = ConnectAppToPostgres(cf, app)
	case ConnectToRedis:
		fmt.Println("connecting to redis")
		err = ConnectAppToRedis(cf, app)
	case ShowEnvs:
		fmt.Println("showing envs")
		err = ShowAppEnvs(cf, app)
	case AddEnv:
		fmt.Println("adding env")
		if !utils.PresentSecurityQuestion() {
			break
		}
		err = CreateAppEnv(cf, app, "", "")
	case ChangeEnv:
		fmt.Println("changing env")
		err = ChangeAppEnv(cf, app, "", "")
	case ShowLogs:
		fmt.Println("showing logs")
		err = GetCurrentLogs(cf, app, "")
	case ShowRecentLogs:
		fmt.Println("showing recent logs")
		err = GetRecentLogs(cf, app, "", "")
	case ShowInstances:
		fmt.Println("showing bound instances")
		err = ShowBoundInstances(cf, app)
	case ManipulateInstances:
		fmt.Println("manipulating instances")
		err = ManipulateAppInstances(cf, app)
	case EnableSsh:
		err = applicationsUtils.EnableAppSsh(cf, app.GUID)
	}
	return err
}

func appsCompletion(appsList *map[string]AppData) {
	for appName, _ := range *appsList {
		fmt.Println(appName)
	}
}
