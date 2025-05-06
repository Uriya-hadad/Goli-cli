package instancesTypes

import (
	"github.com/cloudfoundry/go-cfclient/v3/client"
	"goli-cli/utils"
	"goli-cli/utils/instanceUtils"
	"goli-cli/utils/instancesTypesUtils"
	"goli-cli/utils/outputUtils"
	"sync"
	"time"
)

type SaasRegistryInstance struct {
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

func (p *SaasRegistryInstance) GetBoundDetails(cf *client.Client) (map[string]interface{}, error) {
	if p.FullDetails != nil {
		return p.FullDetails, nil
	}
	_, err := p.GetCredentials(cf)
	return p.FullDetails, err
}

func (p *SaasRegistryInstance) GetCredentials(cf *client.Client) (map[string]interface{}, error) {
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

func (p *SaasRegistryInstance) GetGUID() string {
	return p.GUID
}

func (p *SaasRegistryInstance) GetName() string {
	return p.Name
}

func (p *SaasRegistryInstance) SetToken(subdomain, token string) {
	p.Token[subdomain] = token
}

func (p *SaasRegistryInstance) GetToken(cf *client.Client, subdomain string) (string, error) {
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

func (p *SaasRegistryInstance) ListOptions(cf *client.Client) {
	const GetSubscription = "Get dependencies tree of a tenant"
	const UpdateSubsForTenant = "Update Subscription for a tenant"
	const UpdateSubsForLandscape = "Update Subscriptions for the entire landscape"
	const Back = "Back"

	options := []string{GetSubscription, UpdateSubsForTenant, UpdateSubsForLandscape, Back}
	for {
		option, _ := utils.ListAndSelectItem(options, "Please select an option", false)
		var err error

		switch option {
		case GetSubscription:
			var subscription map[string]interface{}

			tenantId := utils.StringPrompt("For which tenant do you want to get the dependencies tree?")
			subscription, err = p.getSubscription(cf, tenantId)
			if err != nil {
				break
			}
			outputUtils.PrintColoredJSON(subscription, nil, nil)
		case UpdateSubsForTenant:
			var subscription map[string]interface{}

			tenantId := utils.StringPrompt("For which tenant do you want to update the dependencies tree?")
			subscription, err = p.getSubscription(cf, tenantId)
			if err != nil {
				break
			}
			var subscriptionStatus string
			subscriptionStatus, err = p.updateSubscriptionAndGetState(cf, subscription["consumerTenantId"].(string), false)
			if err != nil {
				break
			}
			if subscriptionStatus == "SUBSCRIBED" {
				outputUtils.PrintSuccessMessage("Subscription updated successfully for tenant", subscription["consumerTenantId"].(string))
			} else {
				outputUtils.PrintErrorMessage("Subscription update failed for tenant", subscription["consumerTenantId"].(string), "with status", subscriptionStatus)
			}
		case UpdateSubsForLandscape:
			if !utils.PresentSecurityQuestion() {
				break
			}
			var subscriptions []map[string]interface{}
			wgStatusJobs := sync.WaitGroup{}
			wgStatusJobs.Add(1)
			type SubscriptionsStatus struct {
				tenantId string
				state    string
				err      error
			}
			subscriptions, err = p.getAllSubscriptions(cf)
			if err != nil {
				break
			}

			subscriptionsStatus := make(chan SubscriptionsStatus, len(subscriptions))
			wgUpdateJobs := sync.WaitGroup{}
			wgUpdateJobs.Add(len(subscriptions))
			for _, tenantSub := range subscriptions {
				go func() {
					defer wgUpdateJobs.Done()
					tenantId := tenantSub["consumerTenantId"].(string)
					state, err := p.updateSubscriptionAndGetState(cf, tenantId, true)
					subscriptionsStatus <- SubscriptionsStatus{tenantId, state, err}
				}()
				time.Sleep(2 * time.Second)
			}
			wgUpdateJobs.Wait()
			close(subscriptionsStatus)
			go func() {
				defer wgStatusJobs.Done()
				for stateData := range subscriptionsStatus {
					if stateData.err != nil {
						outputUtils.PrintErrorMessage("Subscription update failed for tenant", stateData.tenantId, "with status", stateData.state, "and error", stateData.err.Error())
					}
					if stateData.state != "SUBSCRIBED" {
						outputUtils.PrintSuccessMessage("Subscription updated successfully for tenant", stateData.tenantId)
					} else {
						outputUtils.PrintErrorMessage("Subscription update failed for tenant", stateData.tenantId, "with status", stateData.state)
					}
				}
			}()
			wgStatusJobs.Wait()
			outputUtils.PrintErrorMessage("All subscriptions updated successfully")
		case Back:
			return
		}
		if err != nil {
			outputUtils.PrintErrorMessage("An error occurred:", err.Error())
			err = nil
		}
	}
}

func (p *SaasRegistryInstance) getAllSubscriptions(cf *client.Client) (subscriptions []map[string]interface{}, err error) {
	// Get the token for the provider subdomain
	token, err := p.GetToken(cf, "")
	if err != nil {
		return nil, err
	}
	subscriptions, err = instancesTypesUtils.GetAllSubscriptions(p.Credentials["saas_registry_url"].(string), token)
	return subscriptions, err
}

func (p *SaasRegistryInstance) getSubscription(cf *client.Client, tenantId string) (subscription map[string]interface{}, err error) {
	token, err := p.GetToken(cf, "")
	if err != nil {
		return nil, err
	}
	subscription, err = instancesTypesUtils.GetSubscription(p.Credentials["saas_registry_url"].(string), tenantId, token)
	return subscription, err
}

func (p *SaasRegistryInstance) updateSubscriptionAndGetState(cf *client.Client, tenantId string, silent bool) (string, error) {
	token, err := p.GetToken(cf, "")
	if err != nil {
		return "", err
	}
	err = instancesTypesUtils.UpdateSubscription(p.Credentials["saas_registry_url"].(string), tenantId, token)
	if err != nil {
		return "", err
	}
	outputUtils.PrintInfoMessage("Subscription updated successfully, waiting for the status to be updated...")
	subscriptionStatus, err := p.getStatusOfSubscription(tenantId, silent)
	return subscriptionStatus, err
}

func (p *SaasRegistryInstance) CleanUp(cf *client.Client) {
	if p.keyGUID == "" {
		return
	}
	err := instanceUtils.DeleteKey(cf, p.keyGUID)
	if err != nil {
		outputUtils.PrintWarningMessage("an error occurred while deleting the key:", err.Error())
	}
}

func (p *SaasRegistryInstance) getStatusOfSubscription(tenantId string, silent bool) (string, error) {
	//Get the token for the provider subdomain
	var err error
	subscription := "IN_PROCESS"

	token, err := p.GetToken(nil, "")
	if err != nil {
		return "", err
	}
	for subscription == "IN_PROCESS" {
		time.Sleep(4 * time.Second)
		if !silent {
			outputUtils.PrintInfoMessage("Getting the status of the subscription...")
		}
		subData, err := instancesTypesUtils.GetSubscription(p.Credentials["saas_registry_url"].(string), tenantId, token)
		if err != nil {
			return "", err
		}
		subscription = subData["state"].(string)
	}
	return subscription, nil
}
