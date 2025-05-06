package instancesTypes

import (
	"fmt"
	"github.com/cloudfoundry/go-cfclient/v3/client"
	"github.com/spf13/cobra"
	"goli-cli/types"
	"goli-cli/utils"
	"goli-cli/utils/instanceUtils"
	"goli-cli/utils/outputUtils"
	"sort"
)

func NewUnBindCmd(cf *client.Client) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "unbind INSTANCE_NAME",
		Aliases: []string{"ub"},
		Short:   "Unbind a service instance from a specific Cloud Foundry application.",
		Long: `Unbind a specified service instance from a given application in Cloud Foundry.
This command removes the binding between the service instance and the application, ensuring the app no longer consumes the service.
A restage of the application may be required to apply the changes.

Usage:
  goli instances unbind INSTANCE_NAME [OPTIONS]

Aliases:
  unbind, ub

Arguments:
  INSTANCE_NAME            
      The name of the service instance to unbind.
      This is a required argument and must be specified before any options.

Options:
  -h, --help                   
      Display this help message and exit.

Examples:
  goli instances unbind my-instance
      Unbind the service instance named "my-instance", the command will switch to interactive mode to select the app to unbind.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO: make sure instances is restricting this to must have one arg
			instance := cmd.Context().Value("instance").(types.ManagedInstance)
			return UnBindFromApp(cf, instance)
		},
	}

	cmd.SetHelpTemplate(cmd.Long)

	return cmd
}

func NewBoundAppsCmd(cf *client.Client) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "show-bound-apps INSTANCE_NAME",
		Aliases: []string{"sba"},
		Short:   "List applications bound to a specific Cloud Foundry service instance.",
		Long: `Retrieve and display a list of all applications currently bound to a specified Cloud Foundry service instance.
This command helps you understand which applications are consuming a given service instance.

Usage:
  goli instances show-bound-apps INSTANCE_NAME [OPTIONS]

Aliases:
  show-bound-apps, sba

Arguments:
  INSTANCE_NAME            
      The name of the service instance for which to list bound applications.
      This is a required argument and must be specified before any options.

Options:
  -h, --help                   
      Display this help message and exit.

Examples:
  goli instances show-bound-apps my-instance
      List all applications bound to the service instance named "my-instance."`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO: make sure instances is restricting this to must have one arg
			instance := cmd.Context().Value("instance").(types.ManagedInstance)
			return ShowInstanceBoundApps(cf, instance)
		},
	}

	cmd.SetHelpTemplate(cmd.Long)

	return cmd
}

func NewShowBoundDetailsCmd(cf *client.Client) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "show-credentials INSTANCE_NAME",
		Aliases: []string{"sc", "credentials"},
		Short:   "Display credentials for a specific service instance.",
		Long: `Retrieve and display the credentials of a specified Cloud Foundry service instance.
This command allows you to access the credentials for a service instance.

Usage:
  goli instances show-credentials INSTANCE_NAME [OPTIONS]

Aliases:
  show-credentials, sc, credentials

Arguments:
  INSTANCE_NAME            
      The name of the service instance for which to retrieve credentials.  
      This is a required argument and must be specified before any options.

Options:
  -h, --help                   
      Display this help message and exit.

Examples:
  goli instances show-credentials my-instance
      Display the credentials of the service instance named "my-instance."`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO: make sure instances is restricting this to must have one arg
			instance := cmd.Context().Value("instance").(types.ManagedInstance)
			return ShowBoundDetails(cf, instance)
		},
	}

	cmd.SetHelpTemplate(cmd.Long)

	return cmd
}

func ShowBoundDetails(cf *client.Client, instance types.ManagedInstance) error {
	boundDetails, err := instance.GetBoundDetails(cf)
	if err != nil {
		return err
	}
	outputUtils.PrintColoredJSON(boundDetails, nil, nil)
	return nil
}

func ShowInstanceBoundApps(cf *client.Client, instance types.ManagedInstance) error {
	boundApps, err := instanceUtils.GetBoundApps(cf, instance.GetGUID())
	if err != nil {
		return err
	}
	sort.Slice(boundApps, func(i, j int) bool {
		return boundApps[i].Name < boundApps[j].Name
	})
	for _, app := range boundApps {
		outputUtils.PrintItemsMessage(app.Name)
	}
	return nil
}

func UnBindFromApp(cf *client.Client, instance types.ManagedInstance) error {
	boundApps, err := instanceUtils.GetBoundApps(cf, instance.GetGUID())
	if err != nil {
		return err
	}
	for index, app := range boundApps {
		fmt.Printf("%d. %s\n", index+1, app.Name)
	}
	appNumberAsInt := utils.IntPrompt("Select a number of app:")
	if err != nil {
		return err
	}
	if !utils.PresentSecurityQuestion() {
		return err
	}
	err = instanceUtils.UnBindService(cf, boundApps[appNumberAsInt-1].GUID, instance.GetGUID())
	return nil
}
