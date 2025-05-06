package applications

import (
	"context"
	"fmt"
	"github.com/cloudfoundry/go-cfclient/v3/client"
	"github.com/spf13/cobra"
	. "goli-cli/entities"
	"goli-cli/utils"
	"goli-cli/utils/applicationsUtils"
)

func NewRestageCmd(cf *client.Client) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "restage APP_NAME",
		Short: "Restage a specific application",
		Long: `Restage an application in the current Cloud Foundry space.
This command restages the application, which involves rebuilding the application from the latest source code and dependencies.
It essentially performs a fresh deployment of the app, which can be useful when making updates or troubleshooting issues.
You can use the rolling strategy to minimize downtime during the restage process.

Usage:
  goli applications restage APP_NAME [OPTIONS]

Arguments:
  APP_NAME                
      The name of the application you want to restage.  
      This is a required argument and must be specified before any options.

Options:
  -r, --rolling              
      Perform a rolling restage, restarting application instances incrementally.  
      This strategy ensures that some instances are kept running while others are restaged, reducing downtime.  

  -h, --help                 
      Display this help message and exit.  

Examples:
  goli applications restage my-app
      Restage the application named "my-app" using the default strategy.  

  goli applications restage my-app --rolling
      Restage the application named "my-app" using a rolling strategy, reducing downtime during the process.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO: make sure application is restricting this to must have one arg
			app := cmd.Context().Value("app").(*App)
			isRolling, err := cmd.Flags().GetBool("rolling")
			if err != nil {
				return err
			}
			if !utils.PresentSecurityQuestion() {
				return nil
			}

			if isRolling {
				return RestageAppRolling(cf, app)
			}

			return RestageApp(cf, app)
		},
	}
	cmd.Flags().BoolP("rolling", "r", false, "Perform a rolling restage to reduce downtime.")

	cmd.SetHelpTemplate(cmd.Long)

	return cmd
}

func RestageApp(cf *client.Client, app *App) error {
	dropletGUID, err := applicationsUtils.BuildPackage(cf, app.GUID)
	if err != nil {
		return err
	}
	fmt.Println("stopping the app")
	_, err = cf.Applications.Stop(context.Background(), app.GUID)
	if err != nil {
		return err
	}
	fmt.Println("setting the droplet")
	_, err = cf.Droplets.SetCurrentAssociationForApp(context.Background(), app.GUID, dropletGUID)
	if err != nil {
		return err
	}
	fmt.Println("starting the app")
	_, err = cf.Applications.Start(context.Background(), app.GUID)
	if err != nil {
		return err
	}
	err = applicationsUtils.CheckAppStatus(cf, app.GUID, app.Name)
	return err
}

func RestageAppRolling(cf *client.Client, app *App) error {
	dropletGUID, err := applicationsUtils.BuildPackage(cf, app.GUID)
	if err != nil {
		return err
	}
	err = applicationsUtils.CreateDeployment(cf, app.GUID, true, dropletGUID)
	if err != nil {
		return err
	}
	err = applicationsUtils.CheckAppStatus(cf, app.GUID, app.Name)
	return err
}
