package applications

import (
	"context"
	"fmt"
	"github.com/cloudfoundry/go-cfclient/v3/client"
	"github.com/spf13/cobra"
	. "goli-cli/entities"
	"goli-cli/utils"
	"goli-cli/utils/outputUtils"
	"sort"
	"strings"
)

func NewCreateEnvCmd(cf *client.Client) *cobra.Command {
	var envName, envValue string

	cmd := &cobra.Command{
		Use:   "create-env APP_NAME",
		Short: "Create a new environment variable for a specific application.",
		Long: `Create a new environment variable for a specified application in the current Cloud Foundry space.
This command allows you to add a new environment variable to an application.
If the flags '--name' or '--value' are not provided, the command will switch to interactive mode where you can specify the environment variable name and value interactively.
After the environment variable is created, a restart of the application is required for the change to take effect.

Usage:
  goli applications create-env APP_NAME [OPTIONS]

Arguments:
  APP_NAME                
      The name of the application for which you want to create the environment variable.  
      This is a required argument and must be specified before any options.

Options:
  --name <env-var-name>      
      (Optional) The name of the environment variable you want to create.  
      If not specified, the command will enter interactive mode.

  --value <env-var-value>    
      (Optional) The value for the new environment variable.  
      If not specified, the command will enter interactive mode.

  -h, --help                 
      Display this help message and exit.  

Examples:
  goli applications create-env my-app --name NEW_VAR --value new-value
      Create the environment variable "NEW_VAR" with the value "new-value" for the "my-app" application.

  goli applications create-env my-app --name NEW_VAR
      Enter interactive mode to set the value of the environment variable "MY_VAR" .
  
  goli applications create-env my-app
      Enter interactive mode to specify the environment variable name and value.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO: make sure application is restricting this to must have one arg
			app := cmd.Context().Value("app").(*App)
			if !utils.PresentSecurityQuestion() {
				return nil
			}
			return CreateAppEnv(cf, app, envName, envValue)
		},
	}

	cmd.Flags().StringVarP(&envName, "key", "k", "", "Environment variable name (optional).")
	cmd.Flags().StringVarP(&envValue, "value", "v", "", "New value for the environment variable (optional).")

	cmd.SetHelpTemplate(cmd.Long)

	return cmd
}

func NewChangeEnvCmd(cf *client.Client) *cobra.Command {
	var envName, envValue string

	cmd := &cobra.Command{
		Use:   "change-env APP_NAME",
		Short: "Change the environment variables for a specific application.",
		Long: `Change an environment variable for a specified application in the current Cloud Foundry space.
This command allows you to update or add a new environment variable to an application.
If the flags '--name' or '--value' are not provided, the command will switch to an interactive mode where you can specify the environment variable name and value interactively.
After the environment variable is updated, a restart of the application is required for the change to take effect.

Usage:
  goli applications change-env APP_NAME [OPTIONS]

Arguments:
  APP_NAME                
      The name of the application for which you want to change the environment variable.  
      This is a required argument and must be specified before any options.

Options:
  --name <env-var-name>      
      (Optional) The name of the environment variable you want to change or add.  
      If not specified, the command will enter interactive mode.

  --value <env-var-value>    
      (Optional) The new value for the environment variable.  
      If not specified, the command will enter interactive mode.

  -h, --help                 
      Display this help message and exit.  

Examples:
  goli applications change-env my-app --name MY_VAR --value new-value
      Change the value of the environment variable "MY_VAR" to "new-value" for the "my-app" application.

  goli applications change-env my-app --name DEBUG_MODE --value false
      Change the value of the environment variable "DEBUG_MODE" to "false" for the "my-app" application.  

  goli applications change-env my-app --name MY_VAR
      Enter interactive mode to set the value of the environment variable "MY_VAR".  

  goli applications change-env my-app
      Enter interactive mode to specify the environment variable name and value.  
`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO: make sure application is restricting this to must have one arg
			app := cmd.Context().Value("app").(*App)
			return ChangeAppEnv(cf, app, envName, envValue)
		},
	}
	cmd.Flags().StringVarP(&envName, "key", "k", "", "Environment variable name (optional).")
	cmd.Flags().StringVarP(&envValue, "value", "v", "", "New value for the environment variable (optional).")

	cmd.SetHelpTemplate(cmd.Long)

	return cmd
}

func NewShowEnvsCmd(cf *client.Client) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "envs APP_NAME",
		Short: "Display environment variables set for a specific application.",
		Long: `Display the environment variables set for the specified application in the current Cloud Foundry space.
This command allows developers to view all environment variables that have been defined for an application, including user-provided variables, system variables, and any custom configurations.
This can help troubleshoot and verify environment configurations for the application.

Usage:
  goli applications envs APP_NAME [OPTIONS]

Arguments:
  APP_NAME                
      The name of the application whose environment variables you want to view.  
      This is a required argument and must be specified before any options.

Options:
  -h, --help                 
      Display this help message and exit.  

Examples:
  goli applications envs my-app
      Display all environment variables for the application named "my-app."  

  goli applications envs my-app --help
      Show help for the envs command.  
`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO: make sure application is restricting this to must have one arg
			app := cmd.Context().Value("app").(*App)

			return ShowAppEnvs(cf, app)
		},
	}

	cmd.SetHelpTemplate(cmd.Long)

	return cmd
}

func ChangeAppEnv(cf *client.Client, app *App, envName, envValue string) error {
	env, err := app.GetEnv(cf)
	if err != nil {
		return err
	}
	envVariables := env.EnvVars

	selectedEnv := envName
	if selectedEnv == "" {
		maxLength := 0
		keys := make([]string, 0, len(envVariables))
		for key := range envVariables {
			if len(key) > maxLength {
				maxLength = len(key)
			}
			keys = append(keys, key)
		}

		i := 1
		sort.Strings(keys)
		for _, key := range keys {
			fmt.Printf("%d. %s%s : %s\n", i, key, strings.Repeat(" ", maxLength-len(key)), envVariables[key])
			i++
		}

		envNameAsInt := utils.IntPrompt("select a number of env:")
		if err != nil {
			return err
		}
		selectedEnv = keys[envNameAsInt-1]
	} else {
		if _, ok := envVariables[selectedEnv]; !ok {
			return fmt.Errorf("env %s not found", selectedEnv)
		}
	}
	fmt.Println("current value: ", envVariables[selectedEnv])

	newValue := envValue
	if newValue == "" {
		newValue = utils.StringPrompt("enter new value:")
	}

	if !utils.PresentSecurityQuestion() {
		return nil
	}
	_, err = cf.Applications.SetEnvironmentVariables(context.Background(), app.GUID, map[string]*string{selectedEnv: &newValue})
	if err != nil {
		return err
	}
	fmt.Println("env updated")
	app.ResetEnv()
	return nil
}

func CreateAppEnv(cf *client.Client, app *App, envName, envValue string) error {
	key := envName
	if key == "" {
		key = utils.StringPrompt("enter the key name:")
	}
	value := envValue
	if value == "" {
		value = utils.StringPrompt("enter the env value:")
	}
	_, err := cf.Applications.SetEnvironmentVariables(context.Background(), app.GUID, map[string]*string{key: &value})
	if err != nil {
		return err
	}
	outputUtils.PrintSuccessMessage("Env added:")
	outputUtils.PrintSuccessMessage(key, ":", value)
	app.ResetEnv()
	return nil
}

func ShowAppEnvs(cf *client.Client, app *App) error {
	env, err := app.GetEnv(cf)
	if err != nil {
		return err
	}
	envVariables := env.EnvVars

	keys := make([]string, 0, len(envVariables))
	maxLength := 0
	for key := range envVariables {
		if len(key) > maxLength {
			maxLength = len(key)
		}
		keys = append(keys, key)
	}

	sort.Strings(keys)

	for _, key := range keys {
		fmt.Printf("%s%s : %s\n", key, strings.Repeat(" ", maxLength-len(key)), envVariables[key])
	}
	return nil
}
