package applications

import (
	"fmt"
	"github.com/cloudfoundry/go-cfclient/v3/client"
	"github.com/cloudfoundry/go-cfclient/v3/resource"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"goli-cli/db"
	. "goli-cli/entities"
	"goli-cli/utils"
	"goli-cli/utils/outputUtils"
)

func NewPostgresCmd(cf *client.Client) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "postgres APP_NAME",
		Short: "Create an SSH tunnel to the PostgreSQL bound instance",
		Long: `Search for a PostgreSQL bound instance to the specified application, create an SSH tunnel to it, and automatically open a new connection in **TablePlus** with the provided credentials.
This command simplifies the process of accessing a PostgreSQL database that is bound to a Cloud Foundry application by securely connecting to it through an SSH tunnel and launching TablePlus for easy interaction with the database.

Usage:
  goli applications postgres APP_NAME [OPTIONS]

Arguments:
  APP_NAME                
      The name of the application for which you want to find the PostgreSQL bound instance.  
      This is a required argument and must be specified before any options.

Options:
  -h, --help                 
      Display this help message and exit.  

Examples:
  goli applications postgres my-app
      Create an SSH tunnel to the PostgreSQL bound instance for the "my-app" application and open a new connection in TablePlus.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO: make sure application is restricting this to must have one arg
			app := cmd.Context().Value("app").(*App)

			return ConnectAppToPostgres(cf, app)
		},
	}

	cmd.SetHelpTemplate(cmd.Long)

	return cmd
}
func ConnectAppToPostgres(cf *client.Client, app *App) error {
	vcapServices, err := app.GetVcapServices(cf)
	if err != nil {
		if resource.IsNotAuthorizedError(err) {
			outputUtils.PrintErrorMessage("You are not authorized to access this resource")
			return nil
		}
		return err
	}
	postgresServices, ok := (*vcapServices)["postgresql-db"]
	if !ok {
		fmt.Println("No postgresql service found...")
		return nil
	}
	postgresService := postgresServices[0]
	if len(postgresServices) > 1 {
		for index, o := range postgresServices {
			fmt.Printf("%d. %s\n", index+1, o.Name)
		}
		serviceNumberAsInt := utils.IntPrompt("select a number of service:")
		postgresService = postgresServices[serviceNumberAsInt-1]
	}
	connectionInfo, err := db.GetPostgresConnectionInfo(postgresService.Credentials)
	if err != nil {
		return err
	}

	fmt.Println("Open connection to", color.HiCyanString(postgresService.Name))

	err = db.OpenPostgresConnection(cf, connectionInfo, app.GUID, app.Name)
	return err
}
