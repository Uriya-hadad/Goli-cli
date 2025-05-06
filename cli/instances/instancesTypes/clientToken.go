package instancesTypes

import (
	"github.com/atotto/clipboard"
	"github.com/cloudfoundry/go-cfclient/v3/client"
	"github.com/spf13/cobra"
	"goli-cli/types"
	"goli-cli/utils/outputUtils"
)

func NewGenClientTokenCmd(cf *client.Client) *cobra.Command {
	var subdomain string

	cmd := &cobra.Command{
		Use:     "generate-client-token INSTANCE_NAME",
		Aliases: []string{"gct"},
		Short:   "Generate a client token for a service instance.",
		Long: `Generate a client token for a specified service instance in Cloud Foundry.  
This command retrieves a token that can be used for secure interactions with the service instance, such as authentication or API access.

Usage:
  goli instances generate-client-token INSTANCE_NAME [OPTIONS]

Aliases:
  generate-client-token, gct

Arguments:
  INSTANCE_NAME            
      The name of the service instance for which to generate a client token.  
      This is a required argument and must be specified before any options.

Options:
  -s, --subdomain <subdomain>
      The subdomain to generate the token for.

  -h, --help                   
      Display this help message and exit.

Examples:
  goli instances generate-client-token my-instance -s my-subdomain
      Generate a client token for the service instance named "my-instance" with
      the subdomain "my-subdomain".

  goli instances generate-client-token my-instance
      Generate a client token for the service instance named "my-instance".`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO: make sure instances is restricting this to must have one arg
			instance := cmd.Context().Value("instance").(types.ManagedInstance)
			return GenClientToken(cf, instance, subdomain)
		},
	}

	cmd.Flags().StringVarP(&subdomain, "subdomain", "s", "", "The subdomain to generate the token for")
	cmd.SetHelpTemplate(cmd.Long)

	return cmd
}

func GenClientToken(cf *client.Client, instance types.ManagedInstance, subdomain string) error {
	token, err := instance.GetToken(cf, subdomain)
	if err != nil {
		return err
	}
	clipboard.WriteAll(token)

	outputUtils.PrintSuccessMessage("Client token generated and copied to clipboard.")
	return nil
}
