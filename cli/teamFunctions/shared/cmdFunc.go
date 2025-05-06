package shared

import (
	"errors"
	"github.com/cloudfoundry/go-cfclient/v3/client"
	"github.com/spf13/cobra"
	"goli-cli/entities"
	"goli-cli/utils"
	"goli-cli/utils/teamFunctionsUtils"
)

func NewRunQueryCmd(cf *client.Client, instances **map[string][]*entities.Instance) *cobra.Command {
	var query, dbName, fileName string

	cmd := &cobra.Command{
		Use:   "run-query",
		Short: "Run a direct query on the database",
		Long: `Run a direct query on the database without the need to open it or interact with the database manually.
This command allows users to execute SQL queries (SELECT query only) directly from the CLI to retrieve data in the database.

Usage:
  goli team-features run-query [OPTIONS]

Aliases:
  run-query, query

Options:
  -q, --query <query>  
      The SQL query you wish to execute on the database.
      If not specified, an interactive mode will open to input the query.
  
  -d, --db <database>  
      The name of the database where the query will be executed.
      If not specified, an interactive mode will open to choose the database.

  -f, --file <file>
      The name of the file to save the query result to.
      If not specified, the query result will be displayed in the terminal.

  -h, --help  
      Display this help message and exit.
  
Examples:
  goli team-features query -q "SELECT * FROM users" -d "my_database"  
      Execute the query "SELECT * FROM users" on the "my_database" database.

  goli team-features query -q "SELECT * FROM users" -f "users.csv"
      Execute the query "SELECT * FROM users" and save the result to the "users.csv" file.

  goli team-features query  
      Open interactive mode to input the query and select the database.`,
		Aliases: []string{"query"},
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return RunQueryFun(cf, *instances, query, dbName, fileName)
		},
	}

	cmd.Flags().StringVarP(&query, "query", "q", "", "query to run on the DB")
	cmd.Flags().StringVarP(&dbName, "db", "d", "", "the DB on which the query will be run on")
	cmd.Flags().StringVarP(&fileName, "fileName", "f", "", "the name of the file to save the query result to")

	cmd.SetHelpTemplate(cmd.Long)

	return cmd
}
func NewConnectToDbCmd(cf *client.Client, instances **map[string][]*entities.Instance) *cobra.Command {
	var dbName string

	cmd := &cobra.Command{
		Use:   "connect-to-db",
		Short: "Automatically connect to a PostgreSQL database bound to an application",
		Long: `This command simplifies the process of accessing a PostgreSQL database bound to a Cloud Foundry application by automatically detecting the application with a PostgreSQL binding and securely connecting to it. 

Usage:
  goli connect-to-db [OPTIONS]

Aliases:
  connect-to-db, ctd

Options:
  -d, --db <database>  
      The name of the database 
      If not specified, interactive mode will open to choose the database.

  -h, --help              
      Display this help message and exit.  

Examples:
  goli connect-to-db
      list all the postgresql databases and connect to the selected one.

  goli connect-to-db --db-name custom-db-name
      Connect to the PostgreSQL database with the explicitly specified name "custom-db-name".`,
		Aliases: []string{"ctd"},
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return ConnectToDbCmdFunc(cf, *instances, dbName)
		},
	}

	cmd.Flags().StringVarP(&dbName, "db", "d", "", "the DB on which a connection will be established")

	cmd.SetHelpTemplate(cmd.Long)

	return cmd
}

func ConnectToDbCmdFunc(cf *client.Client, instances *map[string][]*entities.Instance, db string) error {
	var pgInstanceRaw *entities.Instance
	pgInstanceRaw, err := teamFunctionsUtils.GetPostgresInstance(db, instances)
	if err != nil {
		return err
	}
	err = teamFunctionsUtils.ConnectToDB(cf, instances, pgInstanceRaw)
	if err != nil {
		return err
	}
	return nil
}
func RunQueryFun(cf *client.Client, instances *map[string][]*entities.Instance, queryInput, db, fileName string) error {
	var query string
	var pgInstanceRaw *entities.Instance

	pgInstanceRaw, err := teamFunctionsUtils.GetPostgresInstance(db, instances)
	if err != nil {
		return err
	}

	if queryInput != "" {
		query = queryInput
	} else {
		query = utils.StringPrompt("Enter the query:")
	}
	if !teamFunctionsUtils.ValidateQuery(query) {
		return errors.New("invalid query - query be a 'select' query and contain 'from' clauses")
	}

	err = teamFunctionsUtils.ConnectAndPrintQuery(cf, pgInstanceRaw, query, fileName)
	if err != nil {
		return err
	}
	return nil
}
