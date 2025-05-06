package applications

import (
	"context"
	"fmt"
	"github.com/cloudfoundry/go-cfclient/v3/client"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
	. "goli-cli/entities"
	"goli-cli/utils/applicationsUtils"
	"goli-cli/utils/outputUtils"
	"strconv"
	"sync"
)

func NewDetailsCmd(cf *client.Client) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "details APP_NAME",
		Aliases: []string{"d", "D"},
		Short:   "Display the current details and health of an application",
		Long: `Retrieve and display the current status, health, and runtime details of a Cloud Foundry application.
This command provides an overview of the application's instances, including their states, uptime, memory and CPU usage.
It simplifies monitoring by consolidating key information about the application's deployment and operational health.

Usage:
  goli applications details APP_NAME [OPTIONS]

Aliases:
  details, d, D

Arguments:
  APP_NAME                
      The name of the application for which to retrieve the details.
      This is a required argument and must be specified before any options.

Options:
  -h, --help                   
      Display this help message and exit.

Examples:
  goli applications details my-app
      Display the overall details and summary information for the application "my-app."`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO: make sure application is restricting this to must have one arg
			app := cmd.Context().Value("app").(*App)
			return PrintDetails(cf, app)
		},
	}

	cmd.SetHelpTemplate(cmd.Long)

	return cmd
}

func PrintDetails(cf *client.Client, app *App) error {
	var err error
	var state string
	var mutex sync.WaitGroup
	var url string

	mutex.Add(1)

	go func() {
		defer mutex.Done()
		appRoute, _ := cf.Routes.FirstForApp(context.Background(), app.GUID, nil)
		url = appRoute.Protocol + "://" + appRoute.URL
	}()

	const memory = 1024 * 1024
	stats, process, err := applicationsUtils.GetFullAppStatus(cf, app.GUID)
	if err != nil {
		return err
	}
	mutex.Wait()

	outputUtils.PrintInfoMessage("App Name: " + app.Name)
	outputUtils.PrintInfoMessage("App GUID: " + app.GUID)
	outputUtils.PrintInfoMessage("App URL: " + url)
	fmt.Println()
	outputUtils.PrintInfoMessage("App Instances:")
	for instanceIndex, stat := range stats {
		outputUtils.PrintInfoMessage(strconv.Itoa(instanceIndex), ":")
		switch stat.State {
		case "RUNNING":
			state = color.GreenString("RUNNING")
		case "DOWN":
			state = color.HiBlackString("DOWN")
			outputUtils.PrintInfoMessage("App Status: " + state)
			continue
		case "STARTING":
			state = color.YellowString("STARTING")
		case "CRASHED":
			state = color.RedString("CRASHED")
			outputUtils.PrintInfoMessage("App Status: " + state)
			continue
		}
		outputUtils.PrintInfoMessage("App Status: " + state)
		outputUtils.PrintInfoMessage(fmt.Sprintf("App CPU: %.1f%%", stat.Usage.CPU*100))
		outputUtils.PrintInfoMessage(fmt.Sprintf("App Memory: %.1fM / %dM", float64(stat.Usage.Memory)/memory, process.MemoryInMB))
		outputUtils.PrintInfoMessage(fmt.Sprintf("App Disk: %.1fM / %dM", float64(stat.Usage.Disk)/memory, process.DiskInMB))
	}

	return err
}
