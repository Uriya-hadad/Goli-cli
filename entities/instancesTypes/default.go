package instancesTypes

import (
	"github.com/cloudfoundry/go-cfclient/v3/client"
	"goli-cli/utils/instanceUtils"
	"goli-cli/utils/outputUtils"
	"sync"
)

type DefaultInstance struct {
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

func (p *DefaultInstance) GetBoundDetails(cf *client.Client) (map[string]interface{}, error) {
	if p.FullDetails != nil {
		return p.FullDetails, nil
	}
	_, err := p.GetCredentials(cf)
	return p.FullDetails, err
}

func (p *DefaultInstance) GetCredentials(cf *client.Client) (map[string]interface{}, error) {
	defer p.getterMutex.Unlock()
	p.getterMutex.Lock()
	if p.Credentials != nil {
		return p.Credentials, nil
	}
	BindingDetails, err := instanceUtils.GetCredFromBinding(cf, p.GUID)
	if BindingDetails != nil {
		p.FullDetails = BindingDetails
		if BindingDetails["uaa"] != nil {
			p.Credentials = BindingDetails["uaa"].(map[string]interface{})
		} else if BindingDetails["x509"] != nil {
			p.Credentials = BindingDetails["x509"].(map[string]interface{})
		} else {
			p.Credentials = BindingDetails
		}
		return p.Credentials, nil
	}
	BindingDetails, keyGUID, err := instanceUtils.CreateKeyForCred(cf, p.GUID, true)
	if err != nil {
		return nil, err
	}
	p.keyGUID = keyGUID
	p.FullDetails = BindingDetails
	if BindingDetails["uaa"] != nil {
		p.Credentials = BindingDetails["uaa"].(map[string]interface{})
	} else {
		p.Credentials = BindingDetails
	}
	return BindingDetails, nil
}

func (p *DefaultInstance) GetGUID() string {
	return p.GUID
}

func (p *DefaultInstance) GetName() string {
	return p.Name
}

func (p *DefaultInstance) SetToken(subdomain, token string) {
	p.Token[subdomain] = token
}

func (p *DefaultInstance) GetToken(cf *client.Client, subdomain string) (string, error) {
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

func (p *DefaultInstance) ListOptions(cf *client.Client) {
	outputUtils.PrintWarningMessage("No options available for this instance type")
}

func (p *DefaultInstance) CleanUp(cf *client.Client) {
	if p.keyGUID == "" {
		return
	}
	err := instanceUtils.DeleteKey(cf, p.keyGUID)
	if err != nil {
		outputUtils.PrintWarningMessage("an error occurred while deleting the key:", err.Error())
	}
}
