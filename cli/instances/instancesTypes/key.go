package instancesTypes

import (
	"github.com/cloudfoundry/go-cfclient/v3/client"
	"github.com/spf13/cobra"
	"goli-cli/types"
	"goli-cli/utils"
	"goli-cli/utils/instanceUtils"
	"goli-cli/utils/outputUtils"
)

func NewCreateKeyCmd(cf *client.Client) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "create-key INSTANCE_NAME",
		Aliases: []string{"ck"},
		Short:   "Generate a new service key for a bound instance to enable secure external access.",
		Long: `Generate a new service key for a bound instance, enabling secure external access or integrations.
This key can be used to authenticate and interact with the service instance programmatically.

Usage:
  goli instances create-key INSTANCE_NAME [OPTIONS]

Aliases:
  create-key, ck

Arguments:
  INSTANCE_NAME
      The name of the service instance for which to create a new key.
      This is a required argument and must be specified before any options.

Options:
  -h, --help  
      Display this help message and exit.

Examples:
  goli instances create-key my-instance  
      Generate a new key for the service instance "my-instance."`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO: make sure instances is restricting this to must have one arg
			instance := cmd.Context().Value("instance").(types.ManagedInstance)
			if !utils.PresentSecurityQuestion() {
				return nil
			}
			return CreateAndPrintKey(cf, instance)
		},
	}

	cmd.SetHelpTemplate(cmd.Long)

	return cmd
}

func CreateAndPrintKey(cf *client.Client, instance types.ManagedInstance) error {
	keyName := utils.StringPrompt("Enter the key name: ")
	ans := utils.StringPrompt("Is it an x509 key? (yes/no): ")
	x509Key := ans == "yes" || ans == "y"
	err := instanceUtils.CreateKey(cf, keyName, x509Key, instance.GetGUID())
	if err != nil {
		return err
	}
	ans = utils.StringPrompt("Do you want to see the key? (yes/no): ")
	if ans == "yes" || ans == "y" {
		key, _, err := instanceUtils.GetKey(cf, keyName, instance.GetGUID())
		if err != nil {
			return err
		}
		outputUtils.PrintColoredJSON(key.Credentials, nil, nil)
	}
	return nil
}
