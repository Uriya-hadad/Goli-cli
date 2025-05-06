package helpers

import (
	"errors"
	"github.com/cloudfoundry/go-cfclient/v3/client"
	"goli-cli/cli/instances/instancesTypes"
	"goli-cli/entities"
	"goli-cli/types"
	"goli-cli/utils"
)

func GenerateClientToken(instances *map[string][]*entities.Instance, cf *client.Client) (string, error) {
	xsuaaInstance := GetInstance(cf, instances, "xsuaa", "portal-xsuaa-broker")
	if xsuaaInstance == nil {
		return "", errors.New("xsuaa instance not found")
	}
	token, err := xsuaaInstance.GetToken(cf, "")
	if err != nil {
		return "", err
	}
	return token, nil
}

func GetPortalUrl(instances *map[string][]*entities.Instance, cf *client.Client) (string, error) {
	generalConfig := GetInstance(cf, instances, "user-provided", "portal-cf-general-configuration")
	if generalConfig == nil {
		return "", errors.New("general configuration instance not found")
	}
	cred, err := generalConfig.GetBoundDetails(cf)
	if err != nil {
		return "", err
	}
	return cred["portalServiceUrl"].(string), nil
}

func GetApp(appsList *map[string]types.AppData, appName string) *entities.App {
	var reqApp *entities.App
	if appName == "" {
		keys := make([]string, 0, len(*appsList))
		for name, _ := range *appsList {
			keys = append(keys, name)
		}
		appName, _ = utils.ListAndSelectItem(keys, "select an app:", true)
	}

	appData := (*appsList)[appName]
	if appData.Name == "" {
		return nil
	}
	reqApp = &entities.App{
		GUID: appData.GUID,
		Name: appData.Name,
	}
	return reqApp
}

func GetInstance(cf *client.Client, instances *map[string][]*entities.Instance, offerName, instanceName string) types.ManagedInstance {
	var reqInstance types.ManagedInstance
	if instanceName != "" {
		if (*instances)[offerName] != nil {
			for _, instance := range (*instances)[offerName] {
				if instance.Name == instanceName {
					reqInstance = instancesTypes.GetManagedInstance(offerName, instance, cf)
					break
				}
			}
		} else {
			for key, _ := range *instances {
				for _, instanceRaw := range (*instances)[key] {
					if instanceRaw.Name == instanceName {
						reqInstance = instancesTypes.GetManagedInstance(offerName, instanceRaw, cf)
						break
					}
				}
			}
		}
	} else {
		offerNames := make([]string, 0)
		for key, _ := range *instances {
			offerNames = append(offerNames, key)
		}
		offerName, _ := utils.ListAndSelectItem(offerNames, "Select an offer:", true)
		servicesNames := make([]string, 0)
		var serviceNum int

		if len((*instances)[offerName]) > 1 {
			for _, instance := range (*instances)[offerName] {
				servicesNames = append(servicesNames, instance.Name)
			}
			_, serviceNum = utils.ListAndSelectItem(servicesNames, "Select an instance:", false)
		}
		reqInstance = instancesTypes.GetManagedInstance(offerName, (*instances)[offerName][serviceNum], cf)
	}
	return reqInstance
}

//func SelectInstance(cf *client.Client, instances *map[string][]*entities.Instance, offerNames *[]string) *types.ManagedInstance {
//	offerName, _ := utils.ListAndSelectItem(*offerNames, "Select an offer to manipulate:", true)
//	servicesNames := make([]string, 0)
//	var serviceNum int
//
//	if len((*instances)[offerName]) > 1 {
//		for _, instance := range (*instances)[offerName] {
//			servicesNames = append(servicesNames, instance.Name)
//		}
//		_, serviceNum = utils.ListAndSelectItem(servicesNames, "Select an instance to manipulate:", false)
//	}
//
//	instance := instancesTypes.GetManagedInstance(offerName, &(*instances)[offerName][serviceNum], cf)
//
//	return &instance
//}
