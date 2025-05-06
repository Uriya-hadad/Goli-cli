package entities

import (
	"fmt"
	"github.com/fatih/color"
	"goli-cli/utils"
	"sort"
	"strings"
)

type Instance struct {
	Name          string `json:"name"`
	Plan          string
	GUID          string `json:"instance_guid"`
	LastOperation struct {
		Type  string
		State string
	}
	Credentials map[string]interface{} `json:"credentials"`
}

func PrintInstances(servicesByOffer *map[string][]*Instance, offersNames *[]string) {
	maxLength := 0
	for key := range *servicesByOffer {
		for _, val := range (*servicesByOffer)[key] {
			if len(val.Name) > maxLength {
				maxLength = len(val.Name)
			}
		}
	}

	ups := (*servicesByOffer)["user-provided"]
	color.HiBlue(strings.ToUpper("user-provided"))
	for _, val := range ups {
		fmt.Printf("\t%s%s : %s\n", color.HiGreenString(val.Name), strings.Repeat(" ", maxLength-len(val.Name)), color.HiYellowString(val.Plan))
	}
	delete(*servicesByOffer, "user-provided")
	sort.Strings(*offersNames)
	*offersNames = utils.RemoveKeyFromArray(*offersNames, "user-provided")
	fmt.Printf(color.HiCyanString("--------------------\n"))

	for _, value := range *offersNames {
		color.HiBlue(strings.ToUpper(value))
		servicesByName := (*servicesByOffer)[value]
		for _, val := range servicesByName {
			fmt.Printf("\t%s%s : %s\n", color.HiGreenString(val.Name), strings.Repeat(" ", maxLength-len(val.Name)), color.HiYellowString(val.Plan))
		}
	}
}
