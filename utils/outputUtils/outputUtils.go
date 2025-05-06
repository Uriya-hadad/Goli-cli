package outputUtils

import (
	"fmt"
	"github.com/fatih/color"
	"os"
	"reflect"
	"sort"
	"strings"
)

func PrintColoredJSON(obj map[string]interface{}, keyColor, valueColor func(format string, a ...interface{}) string) {
	if keyColor == nil {
		keyColor = color.GreenString
	}
	if valueColor == nil {
		valueColor = color.RedString
	}
	printJSON(obj, "", keyColor, valueColor)
	fmt.Println("")
}

func printJSON(obj map[string]interface{}, indent string, keyColor, valueColor func(format string, a ...interface{}) string) {
	keys := make([]string, 0, len(obj))
	for key := range obj {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	fmt.Println(indent + "{")
	for _, key := range keys {
		if reflect.TypeOf(obj[key]) == reflect.TypeOf(map[string]interface{}{}) {
			fmt.Println(indent+"    "+keyColor(key), ":")
			printJSON(obj[key].(map[string]interface{}), indent+"    ", keyColor, valueColor)
			println("")
		} else if reflect.TypeOf(obj[key]) == reflect.TypeOf([]interface{}{}) {
			currentIntent := indent + "    "
			fmt.Print(currentIntent+keyColor(key), ": [")
			if len(obj[key].([]interface{})) != 0 {
				fmt.Println("")
			} else {
				fmt.Println(" ]")
				continue
			}
			for _, item := range obj[key].([]interface{}) {
				if reflect.TypeOf(item) == reflect.TypeOf(map[string]interface{}{}) {
					printJSON(item.(map[string]interface{}), currentIntent+"    ", keyColor, valueColor)
				} else {
					fmt.Print(currentIntent+keyColor(key), ":", valueColor(fmt.Sprintf("%v", item)))
				}
				fmt.Println(", ")
			}
			fmt.Println(currentIntent + "]")

		} else {
			fmt.Println(indent+"    "+keyColor(key), ":", valueColor(fmt.Sprintf("%v", obj[key])))
		}
	}

	fmt.Print(indent + "}")
}

func PrintInterface(s interface{}) {
	interfaceFormatted := make(map[string]interface{})
	v := reflect.ValueOf(s)
	t := v.Type()

	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		if field.Kind() == reflect.String && field.String() != "" {
			interfaceFormatted[t.Field(i).Name] = field.String()
		}
	}

	// print json pretty

	fmt.Println(color.BlackString("--------------------------------------------------"))
	PrintColoredJSON(interfaceFormatted, color.GreenString, color.HiCyanString)
	fmt.Println(color.BlackString("--------------------------------------------------"))
}

func PrintColoredJsons(jsons []map[string]interface{}) {
	for _, json := range jsons {
		PrintColoredJSON(json, nil, nil)
	}
}

func PrintInfoMessage(stringArgs ...string) {
	var message string
	message = strings.Join(stringArgs, " ")
	fmt.Println(color.HiBlueString(message))
}

func PrintErrorMessage(stringArgs ...string) {
	var message string
	message = strings.Join(stringArgs, " ")
	fmt.Println(color.HiRedString(message))
}

func PrintSuccessMessage(stringArgs ...string) {
	var message string
	message = strings.Join(stringArgs, " ")
	fmt.Println(color.HiCyanString(message))
}

func PrintItemsMessage(stringArgs ...string) {
	var message string
	message = strings.Join(stringArgs, " ")
	fmt.Println(color.HiYellowString(message))
}

func PrintWarningMessage(stringArgs ...string) {
	var message string
	message = strings.Join(stringArgs, " ")
	fmt.Println(color.HiBlackString(message))
}

func PrintQuestionMessage(stringArgs ...string) {
	var message string
	message = strings.Join(stringArgs, " ")
	fmt.Println(color.GreenString(message))
}

func Panic(stringArgs ...string) {
	PrintErrorMessage(stringArgs...)
	os.Exit(0)
}
