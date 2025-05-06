package applications

import (
	"context"
	"fmt"
	"github.com/cloudfoundry/go-cfclient/v3/client"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
	. "goli-cli/entities"
	"goli-cli/utils"
	"goli-cli/utils/applicationsUtils"
)

func NewRestartCmd(cf *client.Client) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "restart APP_NAME",
		Short: "Restart a specific application",
		Long: `Restart an application in the current Cloud Foundry space.
This command stops and then restarts the specified application, ensuring that it is reloaded with the latest configurations and dependencies.
By default, the restart process uses a full stop-start approach, but you can enable the rolling strategy to minimize downtime during the restart.

Usage:
  goli applications restart APP_NAME [OPTIONS]

Arguments:
  APP_NAME                
      The name of the application you want to restart.  
      This is a required argument and must be specified before any options.

Options:
  -r, --rolling              
      Perform a rolling restart, where application instances are restarted incrementally.  
      This strategy reduces downtime by ensuring some instances remain running while others restart.  

  -h, --help                 
      Display this help message and exit.  

Examples:
  goli applications restart my-app
      Restart the application named "my-app" using the default stop-start strategy.  

  goli applications restart my-app --rolling
      Restart the application named "my-app" using a rolling strategy, minimizing downtime during the process.  
`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			app := cmd.Context().Value("app").(*App)
			isRolling, err := cmd.Flags().GetBool("rolling")
			if err != nil {
				return err

			}
			if !utils.PresentSecurityQuestion() {
				return nil
			}
			if isRolling {
				return RestartAppRolling(cf, app)
			}

			return RestartApp(cf, app)
		},
	}
	cmd.Flags().BoolP("rolling", "r", false, "Perform a rolling restart to reduce downtime.")

	cmd.SetHelpTemplate(cmd.Long)

	return cmd
}

func RestartApp(cf *client.Client, app *App) error {
	fmt.Println("restarting application - ", color.HiCyanString(app.Name))
	_, err := cf.Applications.Restart(context.Background(), app.GUID)
	if err != nil {
		return err
	}
	err = applicationsUtils.CheckAppStatus(cf, app.GUID, app.Name)
	return err
}

func RestartAppRolling(cf *client.Client, app *App) error {
	return applicationsUtils.RestartAppRolling(cf, app.GUID, app.Name)
}
