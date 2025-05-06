package applicationsUtils

import (
	"context"
	"fmt"
	"github.com/cloudfoundry/go-cfclient/v3/client"
	"github.com/cloudfoundry/go-cfclient/v3/resource"
	"github.com/fatih/color"
	"goli-cli/utils/outputUtils"
	"sync"
	"time"
)

func CreateDeployment(cf *client.Client, appGUID string, isRestage bool, dropletGUID string) error {
	fmt.Println("creating the deployment")
	deploymentResource := resource.DeploymentCreate{
		Relationships: resource.AppRelationship{
			App: resource.ToOneRelationship{
				Data: &resource.Relationship{
					GUID: appGUID,
				},
			},
		},
		Strategy: "rolling",
	}
	if isRestage {
		deploymentResource.Droplet.GUID = dropletGUID
	}
	deployment, err := cf.Deployments.Create(context.Background(), &deploymentResource)
	if err != nil {
		return err
	}
	fmt.Println("Waiting for deployment to finish...")
	isDeployed := "DEPLOYING"
	for isDeployed == "DEPLOYING" {
		time.Sleep(5 * time.Second)
		deploymentItem, err := cf.Deployments.Get(context.Background(), deployment.GUID)
		if err != nil {
			return err
		}
		isDeployed = deploymentItem.Status.Reason
		fmt.Println("deployment status: ", isDeployed)
	}
	fmt.Println("deployment finished")
	return nil
}

func BuildPackage(cf *client.Client, appGUID string) (string, error) {
	packageItem, err := cf.Packages.FirstForApp(context.Background(), appGUID, &client.PackageListOptions{
		ListOptions: &client.ListOptions{
			OrderBy: "created_at",
		},
		States: client.Filter{Values: []string{"READY"}},
	})
	if err != nil {
		return "", err
	}
	fmt.Println("creating the build")
	build, err := cf.Builds.Create(context.Background(), &resource.BuildCreate{
		Package: resource.Relationship{
			GUID: packageItem.GUID,
		},
	})
	if err != nil {
		return "", err
	}
	buildItem, err := cf.Builds.Get(context.Background(), build.GUID)
	isStage := buildItem.State
	for isStage == "STAGING" {
		time.Sleep(1 * time.Second)
		buildItem, err = cf.Builds.Get(context.Background(), build.GUID)
		if err != nil {
			return "", err
		}
		isStage = buildItem.State
		fmt.Println("build status: ", isStage)
	}
	fmt.Println("build finished")
	return buildItem.Droplet.GUID, nil
}

func CheckAppStatus(cf *client.Client, appGUID string, appName string) error {
	isRunning := "STARTING"
	for isRunning == "STARTING" {
		time.Sleep(2 * time.Second)
		app, err := cf.Processes.GetStatsForApp(context.Background(), appGUID, "web")
		if err != nil {
			return err
		}
		isRunning = app.Stats[0].State
		fmt.Println("app status: ", isRunning)
	}
	if isRunning == "CRASHED" {
		fmt.Println("app crashed!")
		err := GetRecentLogs(cf, appGUID, appName, 3, "", "")
		if err != nil {
			return err
		}
	} else {
		fmt.Println("app is running")
	}
	return nil
}

func EnableAppSsh(cf *client.Client, appGUID string) error {
	res, err := cf.Applications.SSHEnabled(context.Background(), appGUID)
	if err != nil {
		return err
	}
	if res.Enabled {
		fmt.Println("SSH is already enabled")
	} else {
		fmt.Println("Enabling SSH...")
		var updateRes *resource.AppFeature
		updateRes, err = cf.AppFeatures.UpdateSSH(context.Background(), appGUID, true)
		if err != nil {
			return err
		}
		fmt.Println("SSH Status: ", updateRes.Enabled)
	}
	return nil
}

func RestartAppRolling(cf *client.Client, appGUID, appName string) error {
	fmt.Println("restarting application - ", color.HiCyanString(appName))
	err := CreateDeployment(cf, appGUID, false, "")
	if err != nil {
		return err
	}
	err = CheckAppStatus(cf, appGUID, appName)
	return err
}

func GetFullAppStatus(cf *client.Client, appGUID string) ([]resource.ProcessStat, *resource.Process, error) {
	wg := sync.WaitGroup{}
	var err error
	var process *resource.Process
	var stats *resource.ProcessStats

	wg.Add(2)
	go func() {
		defer wg.Done()
		process, err = cf.Processes.FirstForApp(context.Background(), appGUID, nil)
	}()
	go func() {
		defer wg.Done()
		stats, err = cf.Processes.GetStatsForApp(context.Background(), appGUID, "web")
	}()
	wg.Wait()
	if err != nil {
		return nil, nil, err
	}
	return stats.Stats, process, err
}

func PrintStatus(cf *client.Client, appGUID string) error {
	var err error
	var state string

	const memory = 1024 * 1024
	stats, process, err := GetFullAppStatus(cf, appGUID)
	if err != nil {
		return err
	}

	for instanceIndex, stat := range stats {
		fmt.Println(instanceIndex, ":")
		switch stat.State {
		case "RUNNING":
			state = color.GreenString("RUNNING")
		case "DOWN":
			state = color.HiBlackString("DOWN")
			outputUtils.PrintInfoMessage("App Status: " + state)
			continue
		case "STARTING":
			state = color.YellowString("STARTING")
		case "CRASHED":
			state = color.RedString("CRASHED")
			outputUtils.PrintInfoMessage("App Status: " + state)
			continue
		}
		outputUtils.PrintInfoMessage("App Status: " + state)
		outputUtils.PrintInfoMessage(fmt.Sprintf("App CPU: %.1f%%", stat.Usage.CPU*100))
		outputUtils.PrintInfoMessage(fmt.Sprintf("App Memory: %.1fM / %dM", float64(stat.Usage.Memory)/memory, process.MemoryInMB))
		outputUtils.PrintInfoMessage(fmt.Sprintf("App Disk: %.1fM / %dM", float64(stat.Usage.Disk)/memory, process.DiskInMB))
	}

	return err
}
