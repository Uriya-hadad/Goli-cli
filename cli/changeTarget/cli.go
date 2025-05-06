package changeTarget

import (
	"encoding/json"
	"fmt"
	"github.com/cloudfoundry/go-cfclient/v3/client"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
	. "goli-cli/types"
	"goli-cli/utils"
	"goli-cli/utils/outputUtils"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
)

func NewCmd(cf *client.Client, landscapes *Landscape, updateLandLock *sync.WaitGroup, selectedTarget string, currentOrg *CfOrg, currentSpace *CfSpace) *cobra.Command {
	var selectedSpace, selectedOrg string

	cmd := &cobra.Command{
		Use:     "change-target",
		Aliases: []string{"ct"},
		Short:   "Switch between different Cloud Foundry organizations or spaces to manage resources across environments.",
		Long: `The 'change-target' command allows you to switch between different Cloud Foundry organizations or spaces to manage resources across environments.
Running this command opens an interactive mode where you can select the organization and space you want to manage.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			isSpacesRaw, _ := cmd.Flags().GetBool("sRaw")
			isOrgsRaw, _ := cmd.Flags().GetBool("oRaw")
			if isOrgsRaw {
				orgsCompletion(landscapes, updateLandLock, selectedTarget)
				return nil
			}
			if isSpacesRaw {
				targetedOrg := cmd.Flag("orgTar").Value.String()
				targetedOrg = strings.Replace(targetedOrg, "%20", " ", -1)
				if targetedOrg == "" {
					targetedOrg = currentOrg.Name
				}
				spacesCompletion(landscapes, updateLandLock, selectedTarget, targetedOrg)
				return nil
			}
			selectedOrg = strings.Replace(selectedOrg, "%20", " ", -1)
			selectedSpace = strings.Replace(selectedSpace, "%20", " ", -1)
			selectedSpace, selectedOrg, err := ChangeTarget(cf, landscapes, updateLandLock, selectedTarget, currentSpace, currentOrg, selectedOrg, selectedSpace)
			fmt.Println("Running the Cli in org:", color.HiCyanString(selectedOrg.Name), "and space:", color.HiCyanString(selectedSpace.Name))
			return err
		},
	}
	cmd.Flags().StringVarP(&selectedOrg, "org", "o", "", "The organization to manage.")
	cmd.Flags().StringVarP(&selectedSpace, "space", "s", "", "The space to manage.")
	// create a raw flags for retuning the raw spaces/orgs data - for completion
	cmd.Flags().StringP("orgTar", "", "", "The organization to fetch from.")
	cmd.Flags().MarkHidden("orgTar")
	cmd.Flags().BoolP("sRaw", "", false, "return all of orgs by name")
	cmd.Flags().MarkHidden("sRaw")
	cmd.Flags().BoolP("oRaw", "", false, "return all of spaces by name")
	cmd.Flags().MarkHidden("oRaw")
	return cmd
}

func ChangeTarget(cf *client.Client, landscapes *Landscape, updateLandLock *sync.WaitGroup, selectedTarget string, currentSpace *CfSpace, currentOrg *CfOrg, selectedOrgName, selectedSpaceName string) (selectedSpace *CfSpace, selectedOrg *CfOrg, err error) {
	if (*landscapes)[selectedTarget] == nil {
		updateLandLock.Wait()
	}

	var orgs = (*landscapes)[selectedTarget]

	if selectedOrgName == "" {
		if selectedSpaceName == "" || currentOrg.Name == "" {
			// both org and space are empty

			if len(orgs) == 1 {
				selectedOrg = &CfOrg{Name: orgs[0].Name, GUID: orgs[0].GUID}
			} else {
				for index, o := range orgs {
					fmt.Printf("%d. %s\n", index+1, o.Name)

				}
				orgNameAsInt := utils.IntPrompt("Please select an organization from the list:")
				org := orgs[orgNameAsInt-1]
				selectedOrg = &CfOrg{Name: org.Name, GUID: org.GUID}
			}

			var spaces []*CfSpace
			for _, org := range orgs {
				if org.Name == selectedOrg.Name {
					spaces = org.Spaces
					break
				}
			}

			sort.Slice(spaces, func(i, j int) bool {
				return spaces[i].Name < spaces[j].Name
			})

			if len(spaces) == 0 {
				outputUtils.PrintErrorMessage("No spaces found in the organization")
				return ChangeTarget(cf, landscapes, updateLandLock, selectedTarget, currentSpace, currentOrg, "", "")
			} else if len(spaces) == 1 {
				selectedSpace = &CfSpace{Name: spaces[0].Name, GUID: spaces[0].GUID}
			} else {
				for index, o := range spaces {
					fmt.Printf("%d. %s\n", index+1, o.Name)
				}
				spaceNameAsInt := utils.IntPrompt("Please select a space from the list:")
				space := spaces[spaceNameAsInt-1]
				selectedSpace = &CfSpace{Name: space.Name, GUID: space.GUID}
			}

		} else {
			// org is empty and space is not empty
			var org *CfOrg
			for _, tOrg := range orgs {
				if tOrg.Name == currentOrg.Name {
					org = tOrg
					break
				}
			}
			var space *CfSpace
			for _, s := range org.Spaces {
				if s.Name == selectedSpaceName {
					space = s
					break
				}
			}

			if space == nil {
				outputUtils.PrintWarningMessage("Space", selectedSpaceName, "is not found")
				// TODO check logic here!
				return ChangeTarget(cf, landscapes, updateLandLock, selectedTarget, currentSpace, currentOrg, currentOrg.Name, "")
			}
			selectedSpace = &CfSpace{Name: space.Name, GUID: space.GUID}
			selectedOrg = currentOrg
		}
	} else {
		if selectedSpaceName == "" {
			// org is not empty and space is empty
			var org *CfOrg
			for _, tOrg := range orgs {
				if tOrg.Name == selectedOrgName {
					org = tOrg
					break
				}
			}

			if org == nil {
				outputUtils.PrintWarningMessage("Organization", selectedOrgName, "is not found")
				return ChangeTarget(cf, landscapes, updateLandLock, selectedTarget, currentSpace, currentOrg, "", "")
			}
			selectedOrg = &CfOrg{Name: org.Name, GUID: org.GUID}

			spaces := org.Spaces
			sort.Slice(spaces, func(i, j int) bool {
				return spaces[i].Name < spaces[j].Name
			})

			if len(spaces) == 0 {
				outputUtils.PrintErrorMessage("No spaces found in the organization")
				return ChangeTarget(cf, landscapes, updateLandLock, selectedTarget, currentSpace, currentOrg, "", "")
			} else if len(spaces) == 1 {
				selectedSpace = &CfSpace{Name: spaces[0].Name, GUID: spaces[0].GUID}
			} else {
				for index, o := range spaces {
					fmt.Printf("%d. %s\n", index+1, o.Name)
				}
				spaceNameAsInt := utils.IntPrompt("Please select a space from the list:")
				space := spaces[spaceNameAsInt-1]
				selectedSpace = &CfSpace{Name: space.Name, GUID: space.GUID}
			}
		} else {
			// org and space are not empty

			var org *CfOrg
			for _, tOrg := range orgs {
				if tOrg.Name == selectedOrgName {
					org = tOrg
					break
				}
			}
			if org == nil {
				outputUtils.PrintWarningMessage("Organization", selectedOrgName, "is not found")
				return ChangeTarget(cf, landscapes, updateLandLock, selectedTarget, currentSpace, currentOrg, "", "")
			}
			selectedOrg = &CfOrg{Name: org.Name, GUID: org.GUID}

			var space *CfSpace
			for _, s := range org.Spaces {
				if s.Name == selectedSpaceName {
					space = s
					break
				}
			}
			if space == nil {
				outputUtils.PrintWarningMessage("Space", selectedSpaceName, "is not found")
				return ChangeTarget(cf, landscapes, updateLandLock, selectedTarget, currentSpace, currentOrg, selectedOrgName, "")
			}
			selectedSpace = &CfSpace{Name: space.Name, GUID: space.GUID}

		}
	}

	fmt.Println("Moving to org:", color.HiCyanString(selectedOrg.Name), "and space:", color.HiCyanString(selectedSpace.Name))

	// Update the config
	go updateCFConfig(selectedOrg, selectedSpace)
	return selectedSpace, selectedOrg, nil
}

func updateCFConfig(org *CfOrg, space *CfSpace) {

	cfHome := os.Getenv("CF_HOME")
	if cfHome == "" {
		cfHome = os.Getenv("HOME")
	}
	configFile := filepath.Join(cfHome, ".cf", "config.json")

	f, err := os.Open(configFile)
	if err != nil {
		return
	}
	defer f.Close()
	var config map[string]interface{}

	err = json.NewDecoder(f).Decode(&config)
	if err != nil {
		return
	}

	config["OrganizationFields"].(map[string]interface{})["Name"] = org.Name
	config["OrganizationFields"].(map[string]interface{})["GUID"] = org.GUID
	config["SpaceFields"].(map[string]interface{})["Name"] = space.Name
	config["SpaceFields"].(map[string]interface{})["GUID"] = space.GUID

	f, err = os.Create(configFile)
	if err != nil {
		return
	}
	defer f.Close()

	err = json.NewEncoder(f).Encode(config)
	if err != nil {
		return
	}
}

func spacesCompletion(landscapes *Landscape, updateLandLock *sync.WaitGroup, selectedTarget string, targetOrg string) {
	if landscapes == nil {
		updateLandLock.Wait()
	}
	var spaces []string
	for _, org := range (*landscapes)[selectedTarget] {
		if org.Name == targetOrg {
			for _, space := range org.Spaces {
				spaces = append(spaces, space.Name)
			}
			break
		}
	}
	for _, space := range spaces {
		fmt.Println(space)
	}
}

func orgsCompletion(landscapes *Landscape, updateLandLock *sync.WaitGroup, selectedTarget string) {
	if landscapes == nil {
		updateLandLock.Wait()
	}
	for _, org := range (*landscapes)[selectedTarget] {
		fmt.Println(org.Name)
	}
}
