package applications

import (
	"fmt"
	"github.com/cloudfoundry/go-cfclient/v3/client"
	"github.com/spf13/cobra"
	"goli-cli/db"
	. "goli-cli/entities"
	. "goli-cli/types"
	"goli-cli/utils"
	"strconv"
)

func NewRedisCmd(cf *client.Client) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "redis APP_NAME",
		Short: "Create an SSH tunnel to the Redis bound instance.",
		Long: `Search for a Redis bound instance to the specified application, create an SSH tunnel to it, and optionally open a connection in a Redis client with the provided credentials.
This command simplifies secure access to a Redis instance bound to a Cloud Foundry application.

Usage:
  goli applications redis APP_NAME [OPTIONS]

Arguments:
  APP_NAME                
      The name of the application for which you want to find the Redis bound instance.  
      This is a required argument and must be specified before any options.

Options:
  -h, --help                 
      Display this help message and exit.  

Examples:
  goli applications redis my-app
      Create an SSH tunnel to the Redis bound instance for the "my-app" application and open a connection in a Redis client.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO: make sure application is restricting this to must have one arg
			app := cmd.Context().Value("app").(*App)

			return ConnectAppToRedis(cf, app)
		},
	}

	cmd.SetHelpTemplate(cmd.Long)

	return cmd
}
func ConnectAppToRedis(cf *client.Client, app *App) error {
	vcapServices, err := app.GetVcapServices(cf)
	if err != nil {
		return err
	}

	redisServices, ok := (*vcapServices)["redis-cache"]
	if !ok {
		fmt.Println("No redis service found...")
		return nil
	}

	redisService := redisServices[0]
	if len(redisServices) > 1 {
		for index, o := range redisServices {
			fmt.Printf("%d. %s\n", index+1, o.Name)
		}
		serviceNumberAsInt := utils.IntPrompt("select a number of service:")
		redisService = redisServices[serviceNumberAsInt-1]
	}
	connectionInfo := &ConnectionInfo{
		Hostname: redisService.Credentials["hostname"].(string),
		Port:     strconv.Itoa(int(redisService.Credentials["port"].(float64))),
		Username: "",
		Password: redisService.Credentials["password"].(string),
		Dbname:   "",
	}
	err = db.OpenRedisConnection(cf, connectionInfo, app.GUID, app.Name, false)
	return err
}
