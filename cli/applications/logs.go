package applications

import (
	"github.com/cloudfoundry/go-cfclient/v3/client"
	"github.com/spf13/cobra"
	. "goli-cli/entities"
	"goli-cli/utils"
	"goli-cli/utils/applicationsUtils"
)

func NewShowLogsCmd(cf *client.Client) *cobra.Command {
	var level, correlationId string
	var isRecent bool

	cmd := &cobra.Command{
		Use:   "logs APP_NAME",
		Short: "View logs for a specific application.",
		Long: `Retrieve and analyze logs for a specific application in the current Cloud Foundry space.
This command allows you to monitor and troubleshoot application behavior by accessing real-time or historical log data.
Use filters such as correlation ID or log levels to focus on specific events or issues.

Usage:
  goli applications logs APP_NAME [OPTIONS]

Arguments:
  APP_NAME                
      The name of the application for which you want to view logs.  
      This is a required argument and must be specified before any options.

Options:
  -r, --recent               
      Display only the most recent logs for the specified application without streaming in real-time.  
      Use this option to quickly review the last set of log entries, such as after an application crash, restart or deployment.  

  -c, --correlationId <id>   
      Filter logs to display only entries that match the specified correlation ID.  
      A correlation ID is often used to track a specific request or transaction across multiple components, making this  
      option particularly useful for debugging distributed systems.  

  -l, --level <level>        
      Filter logs based on their severity level. Supported values include INFO, WARN, ERROR, and others depending  
      on your logging setup. Use this option to focus on logs of a specific importance, such as errors that require  
      immediate attention or warnings indicating potential issues.  

  -h, --help                 
      Display this help message and exit.  

Examples:
  goli applications logs my-app
      Stream live logs for the application named "my-app," providing real-time insights into its operation.  

  goli applications logs my-app --recent
      Display only the most recent logs for "my-app," ideal for quickly reviewing events after a deployment.  

  goli applications logs my-app --correlationId abc123
      Filter logs for "my-app" to display only those associated with the correlation ID "abc123,"  
      allowing you to trace specific transactions or requests.  

  goli applications logs my-app --level ERROR
      Display only error-level logs for "my-app," focusing on critical issues that need immediate resolution.  

  goli applications logs my-app --level WARN --correlationId abc123
      Combine filters to display warning-level logs associated with the correlation ID "abc123,"  
      offering a targeted view of potential problems for that specific transaction.  
`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			app := cmd.Context().Value("app").(*App)

			if isRecent {
				return GetRecentLogs(cf, app, level, correlationId)
			}

			return GetCurrentLogs(cf, app, level)
		},
	}

	cmd.Flags().BoolVarP(&isRecent, "recent", "r", false, "Display recent logs without streaming.")
	cmd.Flags().StringVarP(&level, "level", "l", "", "Filter logs by severity level (e.g., INFO, WARN, ERROR).")
	cmd.Flags().StringVarP(&correlationId, "correlationId", "c", "", "Filter logs by correlation ID.")

	cmd.SetHelpTemplate(cmd.Long)
	return cmd
}

func GetRecentLogs(cf *client.Client, app *App, level, correlationId string) error {
	numberOfMinutes := utils.IntPrompt("Number of minutes to get logs for ('0' for earliest logs): ")
	err := applicationsUtils.GetRecentLogs(cf, app.GUID, app.Name, numberOfMinutes, level, correlationId)
	return err
}

func GetCurrentLogs(cf *client.Client, app *App, level string) error {
	err := applicationsUtils.GetCurrentLogs(cf, app.GUID, app.Name, level)
	return err
}
