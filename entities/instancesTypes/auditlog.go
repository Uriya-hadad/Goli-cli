package instancesTypes

import (
	"github.com/cloudfoundry/go-cfclient/v3/client"
	"goli-cli/utils/instanceUtils"
	"goli-cli/utils/outputUtils"
	"sync"
)

type AuditlogInstance struct {
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

func (p *AuditlogInstance) GetBoundDetails(cf *client.Client) (map[string]interface{}, error) {
	if p.FullDetails != nil {
		return p.FullDetails, nil
	}
	_, err := p.GetCredentials(cf)
	return p.FullDetails, err
}

func (p *AuditlogInstance) GetCredentials(cf *client.Client) (map[string]interface{}, error) {
	defer p.getterMutex.Unlock()
	p.getterMutex.Lock()
	if p.Credentials != nil {
		return p.Credentials, nil
	}
	BindingDetails, err := instanceUtils.GetCredFromBinding(cf, p.GUID)
	if BindingDetails != nil {
		p.FullDetails = BindingDetails["uaa"].(map[string]interface{})
		p.Credentials = BindingDetails["uaa"].(map[string]interface{})
		return p.Credentials, nil
	}
	BindingDetails, keyGUID, err := instanceUtils.CreateKeyForCred(cf, p.GUID, true)
	if err != nil {
		return nil, err
	}
	p.keyGUID = keyGUID
	p.FullDetails = BindingDetails
	p.Credentials = BindingDetails["uaa"].(map[string]interface{})
	return p.Credentials, nil

}

func (p *AuditlogInstance) GetGUID() string {
	return p.GUID
}

func (p *AuditlogInstance) GetName() string {
	return p.Name
}

func (p *AuditlogInstance) SetToken(subdomain, token string) {
	p.Token[subdomain] = token
}

func (p *AuditlogInstance) GetToken(cf *client.Client, subdomain string) (string, error) {
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

func (p *AuditlogInstance) ListOptions(cf *client.Client) {
	outputUtils.PrintWarningMessage("not implemented yet")
}

func (p *AuditlogInstance) CleanUp(cf *client.Client) {
	if p.keyGUID == "" {
		return
	}
	err := instanceUtils.DeleteKey(cf, p.keyGUID)
	if err != nil {
		outputUtils.PrintWarningMessage("an error occurred while deleting the key:", err.Error())
	}
}
