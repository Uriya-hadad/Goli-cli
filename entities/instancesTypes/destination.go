package instancesTypes

import (
	"fmt"
	"github.com/cloudfoundry/go-cfclient/v3/client"
	"goli-cli/utils"
	"goli-cli/utils/instanceUtils"
	"goli-cli/utils/instancesTypesUtils"
	"goli-cli/utils/outputUtils"
	"sync"
)

type DestinationInstance struct {
	Name          string
	Plan          string
	GUID          string
	LastOperation struct {
		Type  string
		State string
	}
	keyGUID     string
	Token       map[string]string
	Credentials map[string]interface{}
	FullDetails map[string]interface{}
	getterMutex sync.Mutex
}

func (p *DestinationInstance) GetBoundDetails(cf *client.Client) (map[string]interface{}, error) {
	if p.FullDetails != nil {
		return p.FullDetails, nil
	}
	_, err := p.GetCredentials(cf)
	return p.FullDetails, err
}

func (p *DestinationInstance) GetCredentials(cf *client.Client) (map[string]interface{}, error) {
	defer p.getterMutex.Unlock()
	p.getterMutex.Lock()
	if p.Credentials != nil {
		return p.Credentials, nil
	}
	BindingDetails, err := instanceUtils.GetCredFromBinding(cf, p.GUID)
	if BindingDetails != nil {
		p.FullDetails = BindingDetails
		p.Credentials = BindingDetails
		return p.Credentials, nil
	}
	BindingDetails, keyGUID, err := instanceUtils.CreateKeyForCred(cf, p.GUID, true)
	if err != nil {
		return nil, err
	}
	p.keyGUID = keyGUID
	p.FullDetails = BindingDetails
	p.Credentials = BindingDetails
	return BindingDetails, nil
}
func (p *DestinationInstance) GetGUID() string {
	return p.GUID
}

func (p *DestinationInstance) GetName() string {
	return p.Name
}

func (p *DestinationInstance) SetToken(subdomain, token string) {
	p.Token[subdomain] = token
}

func (p *DestinationInstance) GetToken(cf *client.Client, subdomain string) (string, error) {
	if subdomain == "" {
		subdomain = "current"
	}
	value, ok := p.Token[subdomain]
	if ok {
		return value, nil
	}
	cred, err := p.GetCredentials(cf)
	if err != nil {
		return "", err
	}

	token, err := instanceUtils.GenerateClientToken(cred, subdomain)
	if err != nil {
		return "", err
	}
	p.Token[subdomain] = token
	return token, nil
}

func (p *DestinationInstance) ListOptions(cf *client.Client) {
	const GetDestination = "Get a destination of a tenant"
	const GetDestinations = "Get all of the destinations of a tenant"
	const Back = "Back"

	options := []string{GetDestination, GetDestinations, Back}
	for {
		option, _ := utils.ListAndSelectItem(options, "Please select an option", false)
		var err error

		switch option {
		case GetDestination:
			var destinations []map[string]interface{}
			fmt.Println("Getting credentials...")
			destinations, err = p.getDestinations(cf)
			if err != nil {
				break
			}
			dest, _ := utils.ListAndSelectItemMap(destinations, "Please select a destination to view?", true, "Name")
			outputUtils.PrintColoredJSON(dest, nil, nil)
		case GetDestinations:
			var destinations []map[string]interface{}
			fmt.Println("Getting credentials...")
			destinations, err = p.getDestinations(cf)
			if err != nil {
				break
			}
			outputUtils.PrintColoredJsons(destinations)
		case Back:
			return
		}
		if err != nil {
			outputUtils.PrintErrorMessage("An error occurred:", err.Error())
			err = nil
		}
	}
}

func (p *DestinationInstance) getDestinations(cf *client.Client) ([]map[string]interface{}, error) {
	subdomain := utils.StringPrompt("For which subdomain do you want to generate a token?\n(for the current subdomain, press 'Enter'):")
	token, err := p.GetToken(cf, subdomain)
	if err != nil {
		return nil, err
	}

	destinations, err := instancesTypesUtils.GetDestinations(p.Credentials["uri"].(string), token)
	return destinations, err
}

func (p *DestinationInstance) CleanUp(cf *client.Client) {
	if p.keyGUID == "" {
		return
	}
	err := instanceUtils.DeleteKey(cf, p.keyGUID)
	if err != nil {
		outputUtils.PrintWarningMessage("an error occurred while deleting the key:", err.Error())
	}
}
