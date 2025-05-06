package applications

import (
	"fmt"
	"github.com/cloudfoundry/go-cfclient/v3/client"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"goli-cli/cli/instances/instancesTypes"
	. "goli-cli/entities"
	"goli-cli/utils"
	"sort"
	"strings"
)

func NewShowBoundInstancesCmd(cf *client.Client) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "show-bound-instances APP_NAME",
		Aliases: []string{"show-instances", "si"},
		Short:   "Show bound instances for a specific application",
		Long: `Show all bound instances for the specified application in the current Cloud Foundry space.
This command provides details of the services bound to the application, such as instances of databases, caches, or other services that the application is connected to.

Bound instances are used by the application to interact with external services and data.
This command helps you track and manage those connections.

Usage:
  goli applications show-bound-instances APP_NAME [OPTIONS]

Aliases:
  show-bound-instances, show-instances, si

Arguments:
  APP_NAME                
      The name of the application for which you want to show bound instances.  
      This is a required argument and must be specified before any options.

Options:
  -h, --help                 
      Display this help message and exit.  

Examples:
  goli applications instances my-app
      Show all bound instances for the "my-app" application. 
`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			app := cmd.Context().Value("app").(*App)

			return ShowBoundInstances(cf, app)
		},
	}

	cmd.SetHelpTemplate(cmd.Long)

	return cmd
}

func NewManipulateInstanceCmd(cf *client.Client) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "manipulate-instances APP_NAME",
		Aliases: []string{"mi"},
		Short:   "Perform actions on bound instances of the application",
		Long: `Interact with the bound instances of a Cloud Foundry application and perform actions such as viewing credentials, generating keys, unbinding instances, and more.
This command provides a centralized interface for managing and interacting with service instances bound to an application.

Usage:
  goli applications manipulate-instances APP_NAME [OPTIONS]

Aliases:
  manipulate-instances, mi

Arguments:
  APP_NAME
      The name of the application for which to manipulate bound instances.
      This is a required argument and must be specified before any options.

Options:
  -h, --help 
      Display this help message and exit.

Examples:
  goli applications manipulate-instances my-app
      Start an interactive mode to choose an action for managing bound instances of "my-app."`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			app := cmd.Context().Value("app").(*App)

			return ManipulateAppInstances(cf, app)
		},
	}

	cmd.SetHelpTemplate(cmd.Long)

	return cmd
}

func ShowBoundInstances(cf *client.Client, app *App) error {
	VcapServices, err := app.GetVcapServices(cf)
	if err != nil {
		return err
	}
	var offersNames []string
	maxLength := 0
	for offerName, services := range *VcapServices {
		offersNames = append(offersNames, offerName)
		for _, val := range services {
			if len(val.Name) > maxLength {
				maxLength = len(val.Name)
			}
		}
	}

	sort.Strings(offersNames)
	fmt.Printf(color.HiCyanString("--------------------\n"))

	for _, value := range offersNames {
		color.HiBlue(strings.ToUpper(value))
		servicesByName := (*VcapServices)[value]
		for _, val := range servicesByName {
			fmt.Printf("\t%s%s : %s\n", color.HiGreenString(val.Name), strings.Repeat(" ", maxLength-len(val.Name)), color.HiYellowString(val.Plan))
		}
	}
	return nil
}

func ManipulateAppInstances(cf *client.Client, app *App) error {
	vcapServices, err := app.GetVcapServices(cf)
	if err != nil {
		return err
	}

	keys := make([]string, 0, len(*vcapServices))

	for name, _ := range *vcapServices {
		keys = append(keys, name)
	}

	selectedService, _ := utils.ListAndSelectItem(keys, "select a service:", true)
	services := (*vcapServices)[selectedService]
	serviceRaw := services[0]
	if len(services) > 1 {
		for index, o := range services {
			fmt.Printf("%d. %s\n", index+1, o.Name)
		}
		serviceNumberAsInt := utils.IntPrompt("select a number of service:")
		if err != nil {
			return err
		}
		serviceRaw = services[serviceNumberAsInt-1]
	}
	instance := instancesTypes.GetManagedInstance(selectedService, serviceRaw, cf)
	instancesTypes.ListOptions(cf, instance)
	return nil
}
