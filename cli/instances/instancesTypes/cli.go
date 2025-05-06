package instancesTypes

import (
	"fmt"
	"github.com/cloudfoundry/go-cfclient/v3/client"
	"github.com/fatih/color"
	"goli-cli/entities"
	"goli-cli/entities/instancesTypes"
	"goli-cli/types"
	"goli-cli/utils"
	"goli-cli/utils/outputUtils"
)

const (
	ShowCredentials     = "Show credentials"
	CreateKey           = "Create key"
	ShowBoundApps       = "Show bound applications"
	UnBind              = "Unbind from application"
	GenerateClientToken = "Generate client token"
	ViewAllOptions      = "View all service options"
	Back                = "Back"
)

const (
	SUBSCRIPTION_MANAGER = "subscription-manager"
	POSTGRES             = "postgresql-db"
	DESTINATION          = "destination"
	FEATURE_FLAGS        = "feature-flags"
	MESSAGE_QUEUING      = "message-queuing"
	SAAS_REGISTRY_APP    = "saas-registry-application"
	AUDIT_LOG            = "auditlog"
	UPS                  = "user-provided"
)

func ListOptions(cf *client.Client, instance types.ManagedInstance) {

	var err error
	var option string

	defer func() {
		go instance.CleanUp(cf)
	}()

	fmt.Println("Instance name: ", color.HiCyanString(instance.GetName()))
	options := []string{ShowCredentials, CreateKey, ShowBoundApps, UnBind, GenerateClientToken, ViewAllOptions, Back}
	for {
		option, _ = utils.ListAndSelectItem(options, "Select an option:", false)
		if option == Back {
			return
		}
		err = executeOption(cf, instance, option)
		if err != nil {
			outputUtils.PrintErrorMessage("An error occurred:", err.Error())
			err = nil
		}
	}
}

func executeOption(cf *client.Client, instance types.ManagedInstance, option string) error {
	var err error
	switch option {
	case ShowCredentials:
		err = ShowBoundDetails(cf, instance)
	case CreateKey:
		err = CreateAndPrintKey(cf, instance)
	case ShowBoundApps:
		err = ShowInstanceBoundApps(cf, instance)
	case UnBind:
		err = UnBindFromApp(cf, instance)
	case GenerateClientToken:
		err = GenClientToken(cf, instance, "")
	case ViewAllOptions:
		ViewAllInstanceOptions(cf, instance)
	}
	return err
}

func GetManagedInstance(offerName string, serviceInfo *entities.Instance, cf *client.Client) types.ManagedInstance {
	var instance types.ManagedInstance

	// by offer name
	instance = getInstanceByOffer(offerName, serviceInfo)
	if instance == nil {
		// by offer name and plan
		serviceId := offerName + "-" + serviceInfo.Plan
		instance = getInstanceByServiceId(serviceId, serviceInfo, cf)
	}

	if serviceInfo.Credentials == nil {
		go instance.GetCredentials(cf)
	}
	return instance
}

func getInstanceByServiceId(serviceId string, serviceInfo *entities.Instance, cf *client.Client) (instance types.ManagedInstance) {
	switch serviceId {
	case SAAS_REGISTRY_APP:
		instance = &instancesTypes.SaasRegistryInstance{
			Name:          serviceInfo.Name,
			Plan:          serviceInfo.Plan,
			GUID:          serviceInfo.GUID,
			LastOperation: serviceInfo.LastOperation,
			Credentials:   serviceInfo.Credentials,
			FullDetails:   serviceInfo.Credentials,
			Token:         make(map[string]string),
		}
	default:
		instance = &instancesTypes.DefaultInstance{
			Name:          serviceInfo.Name,
			Plan:          serviceInfo.Plan,
			GUID:          serviceInfo.GUID,
			LastOperation: serviceInfo.LastOperation,
			Credentials:   serviceInfo.Credentials,
			FullDetails:   serviceInfo.Credentials,
			Token:         make(map[string]string),
		}
	}
	return instance
}

func getInstanceByOffer(offerName string, serviceInfo *entities.Instance) (instance types.ManagedInstance) {
	switch offerName {
	case SUBSCRIPTION_MANAGER:
		instance = &instancesTypes.SubscriptionManagerInstance{
			Name:          serviceInfo.Name,
			Plan:          serviceInfo.Plan,
			GUID:          serviceInfo.GUID,
			LastOperation: serviceInfo.LastOperation,
			Credentials:   serviceInfo.Credentials,
			FullDetails:   serviceInfo.Credentials,
			Token:         make(map[string]string),
		}
	case POSTGRES:
		instance = &instancesTypes.PostgresInstance{
			Name:          serviceInfo.Name,
			Plan:          serviceInfo.Plan,
			GUID:          serviceInfo.GUID,
			LastOperation: serviceInfo.LastOperation,
			Credentials:   serviceInfo.Credentials,
			FullDetails:   serviceInfo.Credentials,
			Token:         make(map[string]string),
		}
	case DESTINATION:
		instance = &instancesTypes.DestinationInstance{
			Name:          serviceInfo.Name,
			Plan:          serviceInfo.Plan,
			GUID:          serviceInfo.GUID,
			LastOperation: serviceInfo.LastOperation,
			Credentials:   serviceInfo.Credentials,
			FullDetails:   serviceInfo.Credentials,
			Token:         make(map[string]string),
		}
	case FEATURE_FLAGS:
		credentials := serviceInfo.Credentials
		if credentials != nil {
			credentials = credentials["x509"].(map[string]interface{})
		}
		instance = &instancesTypes.FeatureFlagsInstance{
			Name:          serviceInfo.Name,
			Plan:          serviceInfo.Plan,
			GUID:          serviceInfo.GUID,
			LastOperation: serviceInfo.LastOperation,
			Credentials:   credentials,
			FullDetails:   serviceInfo.Credentials,
			Token:         make(map[string]string),
		}
	case MESSAGE_QUEUING:
		credentials := serviceInfo.Credentials
		if credentials != nil {
			credentials = credentials["uaa"].(map[string]interface{})
		}
		instance = &instancesTypes.MessageQueueInstance{
			Name:          serviceInfo.Name,
			Plan:          serviceInfo.Plan,
			GUID:          serviceInfo.GUID,
			LastOperation: serviceInfo.LastOperation,
			Credentials:   credentials,
			FullDetails:   serviceInfo.Credentials,
			Token:         make(map[string]string),
		}
	case AUDIT_LOG:
		credentials := serviceInfo.Credentials
		if credentials != nil {
			credentials = credentials["uaa"].(map[string]interface{})
		}
		instance = &instancesTypes.AuditlogInstance{
			Name:          serviceInfo.Name,
			Plan:          serviceInfo.Plan,
			GUID:          serviceInfo.GUID,
			LastOperation: serviceInfo.LastOperation,
			Credentials:   credentials,
			FullDetails:   serviceInfo.Credentials,
			Token:         make(map[string]string),
		}
	case UPS:
		instance = &instancesTypes.UpsInstance{
			Name:          serviceInfo.Name,
			Plan:          serviceInfo.Plan,
			GUID:          serviceInfo.GUID,
			LastOperation: serviceInfo.LastOperation,
			Credentials:   serviceInfo.Credentials,
			FullDetails:   serviceInfo.Credentials,
			Token:         make(map[string]string),
		}
	}
	return instance
}
