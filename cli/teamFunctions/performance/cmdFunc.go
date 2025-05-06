package performance

import (
	"context"
	"fmt"
	"github.com/cloudfoundry/go-cfclient/v3/client"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"goli-cli/entities"
	. "goli-cli/types"
	"goli-cli/utils"
	"goli-cli/utils/applicationsUtils"
	"sync"
	"time"
)

func NewStatusCmd(cf *client.Client, apps **map[string]AppData) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get-status",
		Short: "Retrieve the status of a service or application repeatedly every 5 seconds.",
		Long: `Retrieve the status of a service or application continuously by running the status check every 5 seconds.
This command provides real-time updates to monitor the status of services or applications over time, which is helpful for tracking issues or monitoring changes in status.

Usage:
  goli team-features get-status

Aliases:
  status, STATUS

Options:
  -h, --help  
      Display this help message and exit.

Examples:
  goli team-features get-status  
      Retrieve and display the status of the service/application every 5 seconds until manually interrupted.`,
		Aliases: []string{"status", "STATUS"},
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return PrintStatusFunc(cf, *apps)
		},
	}

	cmd.SetHelpTemplate(cmd.Long)

	return cmd
}

func NewGetLogLevelCmd(cf *client.Client, apps **map[string]AppData, updateLock *sync.WaitGroup) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get-log-level",
		Short: "Display the log level for each application.",
		Long: `Retrieve and display the current log level settings for all applications.
This command provides a quick overview to identify the logging verbosity for each app in the environment.

Usage:
  goli team-features get-log-level

Aliases:
  gll

Options:
  -h, --help  
      Display this help message and exit.

Examples:
  goli team-features get-log-level  
	  Print the log level (e.g., DEBUG, INFO, WARN, ERROR) for each application in the environment.`,
		Aliases: []string{"gll"},
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return PrintAppsLogLevelFunc(cf, *apps, updateLock)
		},
	}

	cmd.SetHelpTemplate(cmd.Long)

	return cmd
}

func PrintAppsLogLevelFunc(cf *client.Client, appsList *map[string]AppData, updateLock *sync.WaitGroup) error {
	ws := sync.WaitGroup{}
	for _, appData := range *appsList {
		go func() {
			defer ws.Done()
			ws.Add(1)
			app := entities.NewApp(cf, appData.Name, appData.GUID, updateLock, appsList, true)
			stats, _ := cf.Processes.GetStatsForApp(context.Background(), app.GUID, "web")
			if (stats.Stats[0].State) != "RUNNING" {
				return
			}
			env, err := app.GetEnv(cf)
			if err != nil {
				fmt.Println(err)
				return
			}
			envVariables := env.EnvVars
			logLevel := envVariables["CF_APP_LOG_LEVEL"]
			if logLevel != "" {
				fmt.Printf("App: %s, Log Level: %s\n", color.HiCyanString(app.Name), color.HiGreenString(logLevel))
			}
		}()
	}
	ws.Wait()
	return nil
}

func PrintStatusFunc(cf *client.Client, apps *map[string]AppData) error {
	var err error

	keys := make([]string, 0, len(*apps))
	for name, _ := range *apps {
		keys = append(keys, name)
	}
	selectedApp, _ := utils.ListAndSelectItem(keys, "select an app:", true)
	appData := (*apps)[selectedApp]

	app := &entities.App{
		GUID: appData.GUID,
		Name: appData.Name,
	}
	var stop bool
	fmt.Printf("Printing status of %s every %s seconds, click on '%s' to stop\n", color.HiCyanString(app.Name), color.HiCyanString("5"), color.HiRedString("Enter"))
	go func() {
		for {
			if stop {
				break
			}
			err = applicationsUtils.PrintStatus(cf, app.GUID)
			fmt.Println("------------------------------------")
			if err != nil {
				break
			}
			time.Sleep(5 * time.Second)
		}
	}()
	utils.StopUntilEnter()
	stop = true
	return err
}
