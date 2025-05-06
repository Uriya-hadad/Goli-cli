package instancesTypes

import (
	"github.com/cloudfoundry/go-cfclient/v3/client"
	"goli-cli/utils/instanceUtils"
	"goli-cli/utils/outputUtils"
	"sync"
)

type PostgresInstance struct {
	Name          string
	Plan          string
	GUID          string
	LastOperation struct {
		Type  string
		State string
	}
	Token       map[string]string
	keyGUID     string
	Credentials map[string]interface{}
	FullDetails map[string]interface{}
	getterMutex sync.Mutex
}

func (p *PostgresInstance) GetBoundDetails(cf *client.Client) (map[string]interface{}, error) {
	if p.FullDetails != nil {
		return p.FullDetails, nil
	}

	// TODO check if needed
	//param, err := cf.ServiceInstances.GetManagedParameters(context.Background(), p.GUID)
	//var data map[string]interface{}
	//err = json.Unmarshal(*param, &data)
	//if err != nil {
	//	fmt.Println("Error:", err)
	//}
	//outputUtils.PrintColoredJSON(data, nil, nil)

	_, err := p.GetCredentials(cf)
	return p.FullDetails, err
}

func (p *PostgresInstance) GetCredentials(cf *client.Client) (map[string]interface{}, error) {
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

func (p *PostgresInstance) GetGUID() string {
	return p.GUID
}

func (p *PostgresInstance) GetName() string {
	return p.Name
}

func (p *PostgresInstance) SetToken(subdomain, token string) {
	outputUtils.PrintWarningMessage("No token available for this instance type")
}

func (p *PostgresInstance) GetToken(cf *client.Client, subdomain string) (string, error) {
	outputUtils.PrintWarningMessage("No token available for this instance type")
	return "", nil
}

func (p *PostgresInstance) ListOptions(cf *client.Client) {
	outputUtils.PrintWarningMessage("No options available for this instance type")
}

func (p *PostgresInstance) CleanUp(cf *client.Client) {
	if p.keyGUID == "" {
		return
	}
	err := instanceUtils.DeleteKey(cf, p.keyGUID)
	if err != nil {
		outputUtils.PrintWarningMessage("an error occurred while deleting the key:", err.Error())
	}
}
