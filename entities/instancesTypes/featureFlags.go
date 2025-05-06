package instancesTypes

import (
	"fmt"
	"github.com/cloudfoundry/go-cfclient/v3/client"
	"github.com/fatih/color"
	"goli-cli/utils"
	"goli-cli/utils/instanceUtils"
	"goli-cli/utils/instancesTypesUtils"
	"goli-cli/utils/outputUtils"
	"sort"
	"strconv"
	"strings"
	"sync"
)

type FeatureFlagsInstance struct {
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

func (p *FeatureFlagsInstance) GetBoundDetails(cf *client.Client) (map[string]interface{}, error) {
	if p.FullDetails != nil {
		return p.FullDetails, nil
	}
	_, err := p.GetCredentials(cf)
	return p.FullDetails, err
}

func (p *FeatureFlagsInstance) GetCredentials(cf *client.Client) (map[string]interface{}, error) {
	defer p.getterMutex.Unlock()
	p.getterMutex.Lock()
	if p.Credentials != nil {
		return p.Credentials, nil
	}
	BindingDetails, err := instanceUtils.GetCredFromBinding(cf, p.GUID)
	if BindingDetails != nil {
		p.FullDetails = BindingDetails
		p.Credentials = BindingDetails["x509"].(map[string]interface{})
		return p.Credentials, nil
	}
	BindingDetails, keyGUID, err := instanceUtils.CreateKeyForCred(cf, p.GUID, true)
	if err != nil {
		return nil, err
	}
	p.keyGUID = keyGUID
	p.FullDetails = BindingDetails
	p.Credentials = BindingDetails["x509"].(map[string]interface{})
	return p.Credentials, nil
}

func (p *FeatureFlagsInstance) GetGUID() string {
	return p.GUID
}

func (p *FeatureFlagsInstance) GetName() string {
	return p.Name
}

func (p *FeatureFlagsInstance) SetToken(subdomain, token string) {
	p.Token[subdomain] = token
}

func (p *FeatureFlagsInstance) GetToken(cf *client.Client, subdomain string) (string, error) {
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

func (p *FeatureFlagsInstance) ListOptions(cf *client.Client) {
	const GetFeatureFlag = "Get Feature Flag of provider"
	const ChangeFlag = "Change value of a Flag"
	const Back = "Back"

	options := []string{GetFeatureFlag, ChangeFlag, Back}
	for {
		option, _ := utils.ListAndSelectItem(options, "Please select an option", false)
		var err error

		switch option {
		case GetFeatureFlag:
			var featureFlags map[string]bool
			featureFlags, err = p.getFeatureFlags(cf)
			if err != nil {
				break
			}
			maxLength := 0
			keys := make([]string, 0, len(featureFlags))
			for key := range featureFlags {
				keys = append(keys, key)
				if len(key) > maxLength {
					maxLength = len(key)
				}
			}
			sort.Strings(keys)
			for _, key := range keys {
				var flagVal string
				if featureFlags[key] {
					flagVal = color.GreenString("Enabled")
				} else {
					flagVal = color.RedString("Disabled")
				}
				fmt.Printf("%s%s : %s\n", color.GreenString(key), strings.Repeat(" ", maxLength-len(key)), flagVal)
			}
		case ChangeFlag:
			var featureFlags map[string]bool
			featureFlags, err = p.getFeatureFlags(cf)
			if err != nil {
				break
			}
			keys := make([]string, 0, len(featureFlags))
			for name, _ := range featureFlags {
				keys = append(keys, name)
			}
			selectedFlag, _ := utils.ListAndSelectItem(keys, "select a flag:", true)
			outputUtils.PrintInfoMessage("Current value of the flag is:", strconv.FormatBool(featureFlags[selectedFlag]))
			if !utils.PresentSecurityQuestion() {
				break
			}
		case Back:
			return
		}
		if err != nil {
			outputUtils.PrintErrorMessage("An error occurred:", err.Error())
			err = nil
		}
	}
}

func (p *FeatureFlagsInstance) getFeatureFlags(cf *client.Client) (map[string]bool, error) {
	domain := utils.ExtractDomain(cf.Config.ApiURL(""))
	uri := fmt.Sprintf("https://feature-flags.cfapps.%s/api/v1/features", domain)
	token, err := p.GetToken(cf, "current")
	if err != nil {
		return nil, err
	}
	featureFlagsRaw, err := instancesTypesUtils.GetFeatureFlags(uri, token)
	featureFlags := make(map[string]bool, len(featureFlagsRaw))
	for _, ff := range featureFlagsRaw {
		featureFlags[ff.ID] = ff.Enabled
	}
	return featureFlags, err
}

func (p *FeatureFlagsInstance) CleanUp(cf *client.Client) {
	if p.keyGUID == "" {
		return
	}
	err := instanceUtils.DeleteKey(cf, p.keyGUID)
	if err != nil {
		outputUtils.PrintWarningMessage("An error occurred while deleting the key:", err.Error())
	}
}
