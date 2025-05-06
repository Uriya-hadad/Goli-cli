package instances

import (
	"context"
	"errors"
	"fmt"
	"github.com/cloudfoundry/go-cfclient/v3/client"
	"github.com/spf13/cobra"
	"goli-cli/cli/instances/instancesTypes"
	"goli-cli/entities"
	"goli-cli/types"
	"goli-cli/utils"
	"goli-cli/utils/outputUtils"
	"strings"
	"sync"
)

func NewCmd(cf *client.Client, instancesP **map[string][]*entities.Instance, offerNamesP **[]string, instancesLock *sync.WaitGroup) *cobra.Command {
	var instances *map[string][]*entities.Instance
	var offerNames *[]string

	cmd := &cobra.Command{
		Use:     "instances [INSTANCE_NAME]",
		Aliases: []string{"i"},
		Short:   "View or interact with instances or interact with them, such as managing credentials or generating keys.",
		Long: `The goli instances command provides a menu-driven interface to list and manipulate service instances.
When no specific instance is provided, you will be presented with the following options:

Show All Instances
List all available instances, grouped by their offer name and plan, to gain an overview of the resources in your space.

Manipulate Instances
Select a specific instance to perform various actions, such as viewing credentials, creating a client token, and more.`,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			var instance types.ManagedInstance
			instancesLock.Wait()
			instances = *instancesP
			offerNames = *offerNamesP

			if len(args) == 1 {
				for _, offerName := range *offerNames {
					for _, instanceRaw := range (*instances)[offerName] {
						if instanceRaw.Name == args[0] {
							instance = instancesTypes.GetManagedInstance(offerName, instanceRaw, cf)
							break
						}
					}
				}
				if instance == nil {
					outputUtils.Panic("instance do not exist")
				}
				if !strings.HasPrefix(cmd.Use, "instances") {

					ctx := context.WithValue(cmd.Context(), "instance", instance)
					cmd.SetContext(ctx)
				}
			}
		},
		// TODO clashes with line 63.. need to rethink about that
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) > 1 {
				return errors.New("only one arg is accepted")
			}
			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {
			isRaw, _ := cmd.Flags().GetBool("raw")
			if isRaw {
				instancesCompletion(instances, offerNames)
				return
			}

			if len(args) == 1 {
				InstanceCli(cf, instances, offerNames, args[0])
				return
			}
			InstanceCli(cf, instances, offerNames, "")
		},
	}
	cmd.AddCommand(instancesTypes.NewShowBoundDetailsCmd(cf),
		instancesTypes.NewCreateKeyCmd(cf),
		instancesTypes.NewBoundAppsCmd(cf),
		instancesTypes.NewUnBindCmd(cf),
		instancesTypes.NewGenClientTokenCmd(cf),
		instancesTypes.NewViewAllOptionsCmd(cf))

	// create a raw flag for retuning the raw applications data - for completion
	cmd.Flags().BoolP("raw", "", false, "return all of applications by name")
	cmd.Flags().MarkHidden("raw")

	return cmd
}

func InstanceCli(cf *client.Client, instances *map[string][]*entities.Instance, offerNames *[]string, instanceName string) {
	listOptions(cf, instances, offerNames, instanceName)
}

func listOptions(cf *client.Client, instances *map[string][]*entities.Instance, offerNames *[]string, instanceName string) {
	const (
		ShowAllInstances    = "Show All instances"
		ManipulateInstances = "Manipulate instances"
		Back                = "Returning to the previous menu"
	)

	if instanceName != "" {
		for _, offerName := range *offerNames {
			for _, instance := range (*instances)[offerName] {
				if instance.Name == instanceName {
					instance := instancesTypes.GetManagedInstance(offerName, instance, cf)
					instancesTypes.ListOptions(cf, instance)
					return
				}
			}
		}
		outputUtils.Panic("Instance not found...")
	}
	options := []string{ShowAllInstances, ManipulateInstances, Back}
	for {
		selectedOption, _ := utils.ListAndSelectItem(options, "select an option:", false)
		var err error

		switch selectedOption {
		case ShowAllInstances:
			fmt.Println("Show all instances")
			entities.PrintInstances(instances, offerNames)
		case ManipulateInstances:
			fmt.Println("Manipulate instances")
			var offerName string
			offerName, _ = utils.ListAndSelectItem(*offerNames, "Select an offer to manipulate:", true)
			if err != nil {
				break
			}
			var servicesNames []string
			serviceNum := 0
			if len((*instances)[offerName]) > 1 {
				for _, instance := range (*instances)[offerName] {
					servicesNames = append(servicesNames, instance.Name)
				}
				_, serviceNum = utils.ListAndSelectItem(servicesNames, "Select an instance to manipulate:", false)
				if err != nil {
					break
				}
			}
			instance := instancesTypes.GetManagedInstance(offerName, (*instances)[offerName][serviceNum], cf)
			instancesTypes.ListOptions(cf, instance)
		case Back:
			return
		}
		if err != nil {
			outputUtils.PrintErrorMessage("An error occurred:", err.Error())
			err = nil
		}
	}
}

func instancesCompletion(instances *map[string][]*entities.Instance, offerNames *[]string) {
	for _, key := range *offerNames {
		for _, val := range (*instances)[key] {
			fmt.Println(val.Name)
		}
	}
}
