package setUpUtils

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/cloudfoundry/go-cfclient/v3/client"
	"github.com/cloudfoundry/go-cfclient/v3/config"
	"github.com/cloudfoundry/go-cfclient/v3/resource"
	"goli-cli/entities"
	"goli-cli/migrations"
	. "goli-cli/types"
	"goli-cli/utils"
	"goli-cli/utils/outputUtils"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strings"
	"sync"
)

const (
	MACOS   = "darwin"
	WINDOWS = "windows"
)

func UpdateLandscapesData(cf *client.Client, landscapeData *LandscapeData, mutex *sync.Mutex, selectedTarget, domain string, selectedOrg *CfOrg, selectedSpace *CfSpace) {
	defer mutex.Unlock()
	ws := sync.WaitGroup{}
	ws.Add(2)
	var err error
	var apps []*resource.App
	var instancesByOffer *map[string][]*entities.Instance
	go func() {
		defer ws.Done()
		apps, err = cf.Applications.ListAll(context.Background(), &client.AppListOptions{
			SpaceGUIDs:  client.Filter{Values: []string{selectedSpace.GUID}},
			ListOptions: client.NewListOptions(),
		})
		if err != nil {
			if strings.Contains(err.Error(), "invalid_token") {
				outputUtils.Panic("Invalid token, please login again to CF")
			}
			outputUtils.Panic("An error occurred while getting the applications")
		}
	}()
	go func() {
		defer ws.Done()
		instancesByOffer, _, err = getInstances(cf, domain, selectedSpace.GUID)
		if err != nil {
			outputUtils.Panic("An error occurred while getting the instances")
		}
	}()
	updatedApps := make(map[string]AppData)
	updatedInstances := make(OfferData)
	m1 := regexp.MustCompile(`-(blue|green)$`)
	ws.Wait()
	ws.Add(1)
	go func() {
		defer ws.Done()
		InsertDataInstances(instancesByOffer, updatedInstances)
	}()
	for _, app := range apps {
		originalName := m1.ReplaceAllString(app.Name, "")
		InsertAppData(updatedApps, app.Name, originalName, app.GUID)
	}
	mutex.Lock()
	if (*landscapeData) == nil {
		*landscapeData = make(LandscapeData)
	}
	if (*landscapeData)[selectedTarget] == nil {
		(*landscapeData)[selectedTarget] = make(OrgData)
	}
	if (*landscapeData)[selectedTarget][selectedOrg.Name] == nil {
		(*landscapeData)[selectedTarget][selectedOrg.Name] = make(SpaceData)
	}
	if (*landscapeData)[selectedTarget][selectedOrg.Name][selectedSpace.Name] == nil {
		(*landscapeData)[selectedTarget][selectedOrg.Name][selectedSpace.Name] = &CliData{}
	}
	(*landscapeData)[selectedTarget][selectedOrg.Name][selectedSpace.Name].Apps = updatedApps
	(*landscapeData)[selectedTarget][selectedOrg.Name][selectedSpace.Name].Instances = &updatedInstances
	ws.Wait()
	file, _ := json.Marshal(landscapeData)
	err = os.WriteFile("landscapesData.json", file, 0644)
	if err != nil {
		fmt.Println("error writing to file: ", err)
	}
}

func UpdateLandscapes(cf *client.Client, landscapes *Landscape, selectedTarget string) {
	ws := sync.WaitGroup{}
	ws.Add(2)
	var spaces []*resource.Space
	var orgs []*resource.Organization
	var err error

	go func() {
		defer ws.Done()
		var err error
		spaces, err = cf.Spaces.ListAll(context.Background(), nil)
		if err != nil {
			if strings.Contains(err.Error(), "invalid_token") {
				outputUtils.Panic("Invalid token, please login again to CF")
			}
			outputUtils.Panic("An error occurred while getting the spaces")
		}
	}()
	go func() {
		defer ws.Done()
		var err error
		orgs, err = cf.Organizations.ListAll(context.Background(), nil)
		if err != nil {
			if strings.Contains(err.Error(), "invalid_token") {
				outputUtils.Panic("Invalid token, please login again to CF")
			}
			outputUtils.Panic("An error occurred while getting the organizations")
		}
	}()
	ws.Wait()

	sort.Slice(spaces, func(i, j int) bool {
		return spaces[i].Relationships.Organization.Data.GUID < spaces[j].Relationships.Organization.Data.GUID
	})

	sort.Slice(orgs, func(i, j int) bool {
		return orgs[i].GUID < orgs[j].GUID
	})
	var orgsArr []*CfOrg
	i := 0
	spacesLen := len(spaces)

	for _, org := range orgs {
		var spacesArr = make([]*CfSpace, 0)
		for {
			if i < spacesLen && spaces[i].Relationships.Organization.Data.GUID == org.GUID {
				spacesArr = append(spacesArr, &CfSpace{Name: spaces[i].Name, GUID: spaces[i].GUID})
				i++
			} else {
				break
			}
		}
		orgsArr = append(orgsArr, &CfOrg{Name: org.Name, GUID: org.GUID, Spaces: spacesArr})
	}
	if (*landscapes) == nil {
		*landscapes = make(Landscape)
	}
	sort.Slice(orgsArr, func(i, j int) bool {
		return orgsArr[i].Name < orgsArr[j].Name
	})
	(*landscapes)[selectedTarget] = orgsArr
	file, _ := json.Marshal(landscapes)
	err = os.WriteFile("landscapes.json", file, 0644)
	if err != nil {
		fmt.Println("error writing to file: ", err)
	}
}

func InsertDataInstances(instances *map[string][]*entities.Instance, upInstances OfferData) {
	for offer, instances := range *instances {
		var instancesData []InstanceData
		for _, instance := range instances {
			instancesData = append(instancesData, InstanceData{
				Name: instance.Name,
				GUID: instance.GUID,
				Plan: instance.Plan,
			})
		}
		upInstances[offer] = instancesData

	}
}

func InsertAppData(upApps map[string]AppData, originalAppName string, formatedAppName string, appGuid string) {
	if _, ok := upApps[formatedAppName]; ok && formatedAppName == originalAppName {
		return
	}

	upApps[formatedAppName] = AppData{
		Name: originalAppName,
		GUID: appGuid,
	}
}

func getInstances(cf *client.Client, domain string, spaceGUID string) (*map[string][]*entities.Instance, *[]string, error) {
	instances := make(map[string][]*entities.Instance)
	var offersNames []string

	var instancesResParsed InstancesResponse

	servicePlanToGUID := map[string]struct {
		Name                string
		ServiceOfferingGUID string
	}{}

	GuidToName := map[string]string{}

	var serviceName string

	resp, err := http.NewRequest("GET", fmt.Sprintf("https://api.cf.%s/v3/service_instances?fields[service_plan.service_offering.service_broker]=guid,name&fields[service_plan.service_offering]=guid,name,relationships.service_broker&fields[service_plan]=guid,name,relationships.service_offering&order_by=name&per_page=5000&space_guids=%s", domain, spaceGUID), nil)

	instancesRes, err := cf.ExecuteAuthRequest(resp)

	if err != nil {
		if strings.Contains(err.Error(), "invalid_token") {
			outputUtils.Panic("Invalid token, please login again to CF")
		}
		return nil, nil, errors.New("error getting instances for updating landscape")
	}

	responseBody, err := io.ReadAll(instancesRes.Body)
	if err != nil {
		return nil, nil, errors.New("error retrieving instances")
	}
	err = json.Unmarshal(responseBody, &instancesResParsed)

	if err != nil {
		return nil, nil, err
	}

	for _, service := range instancesResParsed.Included.ServicePlans {
		servicePlanToGUID[service.GUID] = struct {
			Name                string
			ServiceOfferingGUID string
		}{Name: service.Name, ServiceOfferingGUID: service.Relationships.ServiceOffering.Data.GUID}
	}
	for _, service := range instancesResParsed.Included.ServiceOfferings {
		GuidToName[service.GUID] = service.Name
		offersNames = append(offersNames, service.Name)
	}
	sort.Strings(offersNames)
	for _, instance := range instancesResParsed.Resources {
		tempStruct := entities.Instance{
			Name: instance.Name,
			GUID: instance.GUID,
			LastOperation: struct{ Type, State string }{
				Type:  instance.LastOperation.Type,
				State: instance.LastOperation.State,
			},
		}

		if instance.Type == "user-provided" {
			tempStruct.Plan = "UPS"
			instances["user-provided"] = append(instances["user-provided"], &tempStruct)
			continue
		}
		plan := servicePlanToGUID[instance.Relationships.ServicePlan.Data.GUID]
		serviceName = GuidToName[plan.ServiceOfferingGUID]
		tempStruct.Plan = plan.Name
		instances[serviceName] = append(instances[serviceName], &tempStruct)
	}
	return &instances, &offersNames, nil
}

func GetLandscapesDataFile(landscapes *LandscapeData) {
	landscapeFile, err := os.ReadFile("landscapesData.json")
	if os.IsNotExist(err) {
		return
	}
	if err != nil {
		fmt.Println(err)
		return
	}
	err = json.Unmarshal(landscapeFile, landscapes)
	if err != nil {
		fmt.Println(err)
	}
}

func GetLandscapesFile(landscapes *Landscape) {
	landscapeFile, err := os.ReadFile("landscapes.json")
	if os.IsNotExist(err) {
		return
	}
	if err != nil {
		fmt.Println(err)
		return
	}
	err = json.Unmarshal(landscapeFile, landscapes)
	if err != nil {
		fmt.Println(err)
	}
}

func UpdateCli() {
	var cmd *exec.Cmd

	osType := runtime.GOOS
	if osType == WINDOWS {
		cmd = exec.Command("powershell", "-File", "./autoUpdate.ps1")
	} else {
		cmd = exec.Command("bash", "autoUpdate.sh")
	}
	if err := cmd.Start(); err != nil {
		fmt.Println(err)
	}
}

func LoadConfig() (user *CfUser, isValid bool, err error) {
	file, err := os.OpenFile("./config.json", os.O_RDWR, 0644)
	defer file.Close()

	if err == nil {
		var goliConfig LocalConfig
		err = json.NewDecoder(file).Decode(&goliConfig)
		if err != nil {
			return nil, false, err
		}
		runMigrations(&goliConfig)
		return &CfUser{Email: goliConfig.Email, Role: decodeRole(goliConfig.Role)}, goliConfig.IUA, nil
	}

	user, err = getUserFromCfConf()
	if err != nil {
		return nil, false, err
	}

	res, err := http.Get("https://goli-cli.cfapps.eu12.hana.ondemand.com/goli/user?mail=" + user.Email)
	if res.StatusCode != 200 {
		return nil, false, errors.New("service might be down, please try again later")
	}
	if err != nil {
		return nil, false, err
	}

	resRaw, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, false, err
	}
	userInfo := UserInfo{}
	_ = json.Unmarshal(resRaw, &userInfo)

	userConfig := LocalConfig{
		PresentByTeams: false,
		Email:          user.Email,
		MigVar:         utils.GetVersion(),
		Role:           userInfo.Role,
		IUA:            userInfo.Auth,
	}
	if !userInfo.Auth {
		return nil, false, nil
	}
	user.Role = decodeRole(userConfig.Role)
	userConfigJson, err := json.Marshal(userConfig)
	if err != nil {
		return nil, false, err
	}
	err = os.WriteFile("config.json", userConfigJson, 0644)
	return user, userConfig.IUA, err
}

func runMigrations(localConfig *LocalConfig) {
	if localConfig.MigVar == "" {
		localConfig.MigVar = "0.0.0"
	}
	var goliConfig LocalConfig
	var err error
	var ran = false

	migrationsList := migrations.GetMigrations(localConfig)
	sort.Slice(migrationsList, func(i, j int) bool {
		return migrations.IsLowerVer(migrationsList[i].Version, migrationsList[j].Version)
	})

	for _, migration := range migrationsList {
		if migrations.IsLowerVer(localConfig.MigVar, migration.Version) {
			outputUtils.PrintInfoMessage("Running migration:", migration.Name)
			err = migration.Up()
			if err != nil {
				break
			}
			ran = true
		}
	}
	if err != nil {
		fmt.Println("Error running migrations: ", err)
		return
	}
	if !ran {
		return
	}
	file, _ := os.ReadFile("./config.json")
	err = json.Unmarshal(file, &goliConfig)
	if err != nil {
		fmt.Println("Error unmarshalling user config: ", err)
	}
	goliConfig.MigVar = strings.TrimSuffix(utils.GetVersion(), "\n")
	userConfigJson, err := json.Marshal(goliConfig)
	if err != nil {
		fmt.Println("Error marshalling user config: ", err)
	}
	err = os.WriteFile("config.json", userConfigJson, 0644)
	if err != nil {
		fmt.Println("Error writing to file: ", err)
	}
}

func getUserFromCfConf() (*CfUser, error) {
	var err error
	var cfConfig struct {
		AccessToken string
	}

	cfHome := os.Getenv("CF_HOME")
	if cfHome == "" {
		cfHome, err = os.UserHomeDir()
	}
	cfHome = filepath.Join(cfHome, ".cf")
	f, err := os.Open(filepath.Join(cfHome, "config.json"))
	if err != nil {
		return nil, err
	}
	defer f.Close()
	err = json.NewDecoder(f).Decode(&cfConfig)

	if err != nil {
		return nil, err
	}
	parts := strings.Split(cfConfig.AccessToken, ".")
	if len(parts) != 3 {
		outputUtils.PrintErrorMessage("Invalid CF token")
		return nil, errors.New("invalid CF token")
	}

	// Decode the payload part
	payloadBytes, err := base64.RawStdEncoding.DecodeString(parts[1])
	if err != nil {
		outputUtils.PrintErrorMessage("Invalid CF token")
		return nil, err
	}

	// Parse the JSON payload
	var userData *CfUser
	err = json.Unmarshal(payloadBytes, &userData)
	return userData, err
}

func CheckAuthorizations(cf *client.Client, spaceGUID string, userEmail string) {
	user, err := cf.Users.ListAll(context.Background(), &client.UserListOptions{
		UserNames: client.Filter{
			Values: []string{userEmail},
		}})
	if err != nil {
		outputUtils.PrintErrorMessage("Cannot check authorization - error while fetching user")
		return
	}
	if user == nil {
		outputUtils.PrintInfoMessage("You don't have any roles in this space - you may not have access to all the features")
		return
	}

	roles, err := cf.Roles.ListAll(context.Background(), &client.RoleListOptions{
		SpaceGUIDs: client.Filter{
			Values: []string{spaceGUID},
		},
		UserGUIDs: client.Filter{
			Values: []string{user[0].Resource.GUID},
		},
		ListOptions: client.NewListOptions(),
	})
	if err != nil {
		outputUtils.PrintErrorMessage("Cannot check authorization - error while fetching user")
		return
	}
	isThereDev := false
	for _, role := range roles {
		if role.Type == "space_developer" {
			isThereDev = true
			break
		}
	}
	if !isThereDev {
		outputUtils.PrintWarningMessage("You don't have any roles in this space - you may not have access to all the features")
	}

}

func DetermineCliFolder() {
	exePath, err := os.Executable()
	if err != nil {
		fmt.Println("Error getting executable path:", err)
		os.Exit(1)
	}

	// Get the directory of the binary
	err = os.Chdir(filepath.Dir(exePath))
	if err != nil {
		fmt.Println("Error changing directory:", err)
		os.Exit(1)
	}
}

func GetAppsFromLandscape(spaceData *CliData, mutex *sync.Mutex) *map[string]AppData {
	defer mutex.Unlock()
	mutex.Lock()
	return &spaceData.Apps
}

func GetInstancesFromLandscape(spaceData *CliData, mutex *sync.Mutex) (*map[string][]*entities.Instance, *[]string) {
	defer mutex.Unlock()
	mutex.Lock()
	instancesByOffer := make(map[string][]*entities.Instance)
	var offerNames []string
	for offer, instances := range *spaceData.Instances {
		var instancesData []*entities.Instance
		for _, instance := range instances {
			instancesData = append(instancesData, &entities.Instance{
				Name: instance.Name,
				GUID: instance.GUID,
				Plan: instance.Plan,
			})
		}
		instancesByOffer[offer] = instancesData
		offerNames = append(offerNames, offer)
	}
	return &instancesByOffer, &offerNames

}

func decodeRole(role string) string {
	decoded, _ := base64.StdEncoding.DecodeString(role)
	return string(decoded)
}

func CreateCfClient() *client.Client {

	cfg, _ := config.NewFromCFHome()

	// there is times when the client id is not set
	if cfg.SSHOAuthClientID() == "" {
		_ = config.SSHOAuthClient("ssh-proxy")(cfg)
	}

	cf, err := client.New(cfg)
	if err != nil {
		outputUtils.Panic(err.Error())
	}

	return cf
}
