package instancesTypes

import (
	"errors"
	"github.com/cloudfoundry/go-cfclient/v3/client"
	"goli-cli/utils/instanceUtils"
	"goli-cli/utils/outputUtils"
	"sync"
)

type UpsInstance struct {
	Name          string
	Plan          string
	GUID          string
	LastOperation struct {
		Type  string
		State string
	}
	Token       map[string]string
	Credentials map[string]interface{}
	FullDetails map[string]interface{}
	getterMutex sync.Mutex
}

func (p *UpsInstance) GetBoundDetails(cf *client.Client) (map[string]interface{}, error) {
	if p.FullDetails != nil {
		return p.FullDetails, nil
	}
	_, err := p.GetCredentials(cf)
	return p.FullDetails, err
}

func (p *UpsInstance) GetCredentials(cf *client.Client) (map[string]interface{}, error) {
	defer p.getterMutex.Unlock()
	p.getterMutex.Lock()
	if p.Credentials != nil {
		return p.Credentials, nil
	}
	BindingDetails, _ := instanceUtils.GetCredFromBinding(cf, p.GUID)
	if BindingDetails != nil {
		p.FullDetails = BindingDetails
		p.Credentials = BindingDetails
		return p.Credentials, nil
	}
	return nil, errors.New("no bound details found for this user provided service")
}

func (p *UpsInstance) GetGUID() string {
	return p.GUID
}

func (p *UpsInstance) GetName() string {
	return p.Name
}

func (p *UpsInstance) SetToken(subdomain, token string) {
	p.Token[subdomain] = token
}

func (p *UpsInstance) GetToken(cf *client.Client, subdomain string) (string, error) {
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

func (p *UpsInstance) ListOptions(cf *client.Client) {
	outputUtils.PrintWarningMessage("No options available for this instance type")
}

func (p *UpsInstance) CleanUp(cf *client.Client) {
	return
}
