package instancesTypes

import (
	"github.com/cloudfoundry/go-cfclient/v3/client"
	"github.com/spf13/cobra"
	"goli-cli/types"
)

func NewViewAllOptionsCmd(cf *client.Client) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "capabilities INSTANCE_NAME",
		Short: "Explore API capabilities of a service instance",
		Long: `Explore and interact with the API capabilities of a specified Cloud Foundry service instance.  
Some service instances, like message queues or SaaS registries for example, provide APIs for operations such as retrieving queue data or managing subscriptions.  
This command launches an interactive mode, allowing you to execute API functions directly through the CLI.

Usage:
  goli instances capabilities INSTANCE_NAME [OPTIONS]

Arguments:
  INSTANCE_NAME            
      The name of the service instance to explore.  
      This is a required argument and must be specified before any options.

Options:
  -h, --help                   
      Display this help message and exit.

Examples:
  goli instances explore-service-api my-instance
      Explore the API capabilities of the service instance "my-instance."`,
		Args: cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			// TODO: make sure instances is restricting this to must have one arg
			instance := cmd.Context().Value("instance").(types.ManagedInstance)
			ViewAllInstanceOptions(cf, instance)
		},
	}

	cmd.SetHelpTemplate(cmd.Long)

	return cmd
}

func ViewAllInstanceOptions(cf *client.Client, instance types.ManagedInstance) {
	instance.ListOptions(cf)
}
