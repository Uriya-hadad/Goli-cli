package entities

import (
	"context"
	"encoding/json"
	"github.com/cloudfoundry/go-cfclient/v3/client"
	"github.com/cloudfoundry/go-cfclient/v3/resource"
	. "goli-cli/types"
	"goli-cli/utils/outputUtils"
	"sync"
)

type App struct {
	GUID          string
	Name          string
	_Env          *resource.AppEnvironment
	_VcapServices *map[string][]*Instance
}

func (app *App) GetEnv(cf *client.Client) (*resource.AppEnvironment, error) {
	err := app.LoadEnv(cf)
	return app._Env, err
}

func (app *App) LoadEnv(cf *client.Client) error {
	if app._Env == nil {
		var err error
		app._Env, err = getEnvForApp(cf, app)
		if err != nil {
			return err
		}
		loadVcapServices(app)
	}
	return nil
}

func loadVcapServices(app *App) {
	_ = json.Unmarshal(app._Env.SystemEnvVars["VCAP_SERVICES"], &app._VcapServices)
}

func (app *App) ResetEnv() {
	app._Env = nil
	app._VcapServices = nil
}

func (app *App) GetVcapServices(cf *client.Client) (*map[string][]*Instance, error) {
	if app._VcapServices == nil {
		env, err := app.GetEnv(cf)
		if err != nil {
			return nil, err
		}
		_ = json.Unmarshal(env.SystemEnvVars["VCAP_SERVICES"], &app._VcapServices)
	}
	return app._VcapServices, nil
}

func getEnvForApp(cf *client.Client, app *App) (*resource.AppEnvironment, error) {
	environment, err := cf.Applications.GetEnvironment(context.Background(), app.GUID)
	return environment, err
}

func NewApp(cf *client.Client, name string, guid string, updateLock *sync.WaitGroup, appList *map[string]AppData, silent bool) *App {
	_, err := cf.Applications.Get(context.Background(), guid)
	var app *App
	if err != nil && resource.IsResourceNotFoundError(err) {
		// App not found
		if !silent {
			outputUtils.PrintWarningMessage("App not found, please wait for the app GUID to be updated")
		}
		updateLock.Wait()
		app = &App{
			GUID: (*appList)[name].GUID,
			Name: name,
		}
	} else {
		// App found
		app = &App{
			GUID: guid,
			Name: name,
		}
	}
	go app.LoadEnv(cf)

	return app

}
