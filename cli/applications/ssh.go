package applications

import (
	"github.com/cloudfoundry/go-cfclient/v3/client"
	"github.com/spf13/cobra"
	. "goli-cli/entities"
	"goli-cli/utils/applicationsUtils"
)

func NewSshCmd(cf *client.Client) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "enable-ssh APP_NAME",
		Short: "Enable SSH access to a specific application",
		Long: `Enable SSH access for the specified application in the current Cloud Foundry space.
By default, Cloud Foundry applications do not have SSH enabled.
Developers can use this command to allow SSH access to an application for debugging or direct interaction with the application’s running environment.
Once SSH is enabled, a restart of the application is required before you can successfully SSH into it.

Usage:
  goli applications enable-ssh APP_NAME [OPTIONS]

Arguments:
  APP_NAME                
      The name of the application for which SSH access should be enabled.  
      This is a required argument and must be specified before any options.

Options:
  -h, --help                 
      Display this help message and exit.  

Examples:
  goli applications enable-ssh my-app
      Enable SSH access for the application named "my-app."  
      After enabling, you will need to restart the app to SSH into it.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO: make sure application is restricting this to must have one arg
			app := cmd.Context().Value("app").(*App)

			return EnableAppSsh(cf, app)
		},
	}

	cmd.SetHelpTemplate(cmd.Long)

	return cmd
}
func EnableAppSsh(cf *client.Client, app *App) error {
	return applicationsUtils.EnableAppSsh(cf, app.GUID)
}
