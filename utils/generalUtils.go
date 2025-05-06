package utils

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/fatih/color"
	. "goli-cli/types"
	"goli-cli/utils/outputUtils"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

var timeNow time.Time

func GetOrgAndSpaceFromConfig() (selectedSpace *CfSpace, selectedOrg *CfOrg, err error) {
	var config struct {
		OrganizationFields struct {
			Name string `json:"Name"`
			GUID string `json:"GUID"`
		} `json:"OrganizationFields"`
		SpaceFields struct {
			Name string `json:"Name"`
			GUID string `json:"GUID"`
		} `json:"SpaceFields"`
	}

	cfHome := os.Getenv("CF_HOME")
	if cfHome == "" {
		cfHome = os.Getenv("HOME")
	}
	configFile := filepath.Join(cfHome, ".cf", "config.json")

	f, err := os.Open(configFile)
	if err != nil {
		return selectedSpace, selectedOrg, err
	}
	defer f.Close()

	err = json.NewDecoder(f).Decode(&config)
	if err != nil {
		return selectedSpace, selectedOrg, fmt.Errorf("error decoding config: %w", err)
	}

	selectedSpace = &CfSpace{config.SpaceFields.Name, config.SpaceFields.GUID}
	selectedOrg = &CfOrg{Name: config.OrganizationFields.Name, GUID: config.OrganizationFields.GUID}

	return selectedSpace, selectedOrg, nil
}

func StringPrompt(label string) string {
	var s string
	r := bufio.NewReader(os.Stdin)
	outputUtils.PrintQuestionMessage(label)
	s, _ = r.ReadString('\n')
	return strings.TrimSpace(s)
}

func IntPrompt(label string) int {
	var s string
	r := bufio.NewReader(os.Stdin)
	for {
		outputUtils.PrintQuestionMessage(label)
		s, _ = r.ReadString('\n')
		s = strings.TrimSpace(s)
		regexpStr := regexp.MustCompile(`^[0-9]+`)
		if s != "" && regexpStr.MatchString(s) {
			break
		}
		outputUtils.PrintErrorMessage("Invalid input - please enter a number")
	}
	ansAsInt, _ := strconv.Atoi(s)
	return ansAsInt
}

func QuestionPrompt(label string) bool {
	ans := StringPrompt(label + " (yes / no)")
	return ans == "y" || ans == "yes"
}

func StopUntilEnter() {
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
}

func PresentSecurityQuestion() bool {
	scanner := bufio.NewReader(os.Stdin)
	for {
		fmt.Println("Are you sure you want to continue? (y/n)")
		input, _, err := scanner.ReadLine()
		if err != nil {
			fmt.Println(err)
			return false
		}
		if string(input) == "y" {
			return true
		}
		if string(input) == "n" {
			return false
		}
	}
}

func ExtractRegion(apiUrl string) string {
	parts := strings.Split(apiUrl, ".")
	if len(parts) <= 2 {
		outputUtils.PrintErrorMessage("Invalid target url", apiUrl)
		return ""
	}
	return parts[2]
}
func ExtractDomain(apiUrl string) string {
	return apiUrl[15:]
}

func ListAndSelectItem(items []string, question string, sortKeys bool) (string, int) {
	var ansAsInt int
	if len(items) == 1 {
		ansAsInt = 1
	} else {
		if sortKeys {
			sort.Strings(items)
		}
		for {
			i := 1
			for _, name := range items {
				fmt.Printf("%d. %s\n", i, name)
				i++
			}
			ansAsInt = IntPrompt(question)
			if ansAsInt > i || ansAsInt < 1 {
				fmt.Println("Invalid number - please select a number from the list")
				continue
			}
			break
		}
	}
	fmt.Println("\nSelected: ", color.HiCyanString(items[ansAsInt-1]))
	return items[ansAsInt-1], ansAsInt - 1

}
func ListAndSelectItemMap(items []map[string]interface{}, question string, sortKeys bool, key string) (map[string]interface{}, int) {
	var ansAsInt int
	var err error
	if len(items) == 1 {
		ansAsInt = 1
	} else {
		if sortKeys {
			sort.Slice(items, func(i, j int) bool {
				return items[i][key].(string) < items[j][key].(string)
			})
		}
		for {
			i := 1
			for _, name := range items {
				fmt.Printf("%d. %s\n", i, name[key])
				i++
			}
			ansAsInt = IntPrompt(question)
			if err != nil || ansAsInt > i || ansAsInt < 1 {
				fmt.Println("Invalid number - please select a number from the list")
				continue
			}
			break
		}
	}

	fmt.Println("\nSelected: ", color.HiCyanString(items[ansAsInt-1][key].(string)))
	return items[ansAsInt-1], ansAsInt - 1

}

func GetStringFromMap(m map[string]interface{}, key string) string {
	if val, ok := m[key]; ok {
		return val.(string)
	}
	return ""
}

func GetVersion() string {
	file, err := os.ReadFile("./version.txt")
	if err != nil {
		return ""
	}
	return strings.Replace(string(file), "\n", "", -1)
}

func InterfaceToString(v []interface{}) []string {
	var result []string
	for _, value := range v {
		result = append(result, fmt.Sprint(value))
	}
	return result
}

func RemoveKeyFromArray[K comparable](array []K, key K) []K {
	for index, value := range array {
		if value == key {
			return append(array[:index], array[index+1:]...)
		}
	}
	return array
}

func SetTime() {
	timeNow = time.Now()
}

func PrintTime(line ...string) {
	if line == nil {
		println(time.Now().Sub(timeNow).Milliseconds())
		return
	}
	println(line[0], time.Now().Sub(timeNow).Milliseconds())
}
