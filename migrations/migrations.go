package migrations

import (
	"encoding/json"
	"errors"
	. "goli-cli/types"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
)

type Migrations struct {
	Name    string
	Up      func() error
	Version string
}

var currentConfig *LocalConfig

var migrations = []Migrations{
	{
		Name:    "add role to each user",
		Version: "1.1.9",
		Up: func() error {
			var userInfo UserInfo
			res, err := http.Get("https://goli-cli.cfapps.eu12.hana.ondemand.com/goli/user?mail=" + currentConfig.Email)
			if err != nil {
				return err
			} else if res.StatusCode != 200 {
				return errors.New("error getting user info")
			}
			defer res.Body.Close()
			resRaw, err := io.ReadAll(res.Body)
			if err != nil {
				return err
			}
			_ = json.Unmarshal(resRaw, &userInfo)
			currentConfig.Role = userInfo.Role
			userConfigJson, err := json.Marshal(currentConfig)
			err = os.WriteFile("config.json", userConfigJson, 0644)
			return nil
		},
	}, {
		Name:    "replace landscape file",
		Version: "1.1.39",
		Up: func() error {
			_, err := os.Stat("landscapes.json")
			if err != nil {
				return nil
			}
			return os.Rename("landscapes.json", "landscapesData.json")
		},
	},
}

func GetMigrations(localConfig *LocalConfig) []Migrations {
	currentConfig = localConfig
	return migrations
}

func IsLowerVer(currentVersion string, migVersion string) bool {
	currentVersionSplit := strings.Split(currentVersion, ".")
	migVersionSplit := strings.Split(migVersion, ".")
	for i := 0; i < len(currentVersionSplit); i++ {
		currentVersionInt, _ := strconv.Atoi(currentVersionSplit[i])
		migVersionInt, _ := strconv.Atoi(migVersionSplit[i])
		if currentVersionInt < migVersionInt {
			return true
		} else if currentVersionInt > migVersionInt {
			return false
		}
	}
	return false
}
