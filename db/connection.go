package db

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"github.com/cloudfoundry/go-cfclient/v3/client"
	"github.com/go-redis/redis"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/olekukonko/tablewriter"
	"golang.org/x/crypto/ssh"
	. "goli-cli/types"
	"goli-cli/utils"
	"goli-cli/utils/applicationsUtils"
	"goli-cli/utils/outputUtils"
	"io"
	"net"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"time"
)

// create type for service name - redis postgres
const (
	REDIS    = "redis"
	POSTGRES = "postgres"
)

func OpenPostgresConnection(cf *client.Client, connectionInfo *ConnectionInfo, appGUID, appName string) error {

	stopChan, err := OpenConnectionToService(cf, connectionInfo, appGUID, POSTGRES, appName)
	if err != nil {
		return err
	}

	var localCred = *connectionInfo

	localCred.Hostname = "127.0.0.1"
	localCred.Port = "5432"

	outputUtils.PrintInterface(localCred)

	err = OpenTablePlusClient(POSTGRES, connectionInfo)

	if err != nil && runtime.GOOS == "darwin" {
		outputUtils.PrintErrorMessage("Error opening the connection:", err.Error())
	} else {
		// Wait for the user to close the connection
		utils.StringPrompt("press enter to close the connection...")
	}
	stopChan <- os.Interrupt
	return err
}

func OpenRedisConnection(cf *client.Client, connectionInfo *ConnectionInfo, appGUID, appName string, isMasterNode bool) error {
	stopChan, err := OpenConnectionToService(cf, connectionInfo, appGUID, REDIS, appName)
	if err != nil {
		return err
	}
	var localCred = *connectionInfo

	localCred.Hostname = "127.0.0.1"
	localCred.Port = "6380"

	// Test connection
	if !isMasterNode {
		rdb := redis.NewClient(&redis.Options{
			Addr:     localCred.Hostname + ":" + localCred.Port,
			Password: localCred.Password,
			TLSConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		})
		pong, err := rdb.Do("ROLE").Result()
		if err != nil {
			outputUtils.PrintErrorMessage("Error connecting to Redis:", err.Error())
			return err
		}

		fmt.Println("Checking the role of the Redis instance...")
		if pong.([]any)[0].(string) == "slave" {
			outputUtils.PrintWarningMessage("You are connected to a Redis slave instance!, connecting to the master instance...")
			fmt.Println("Closing the connection to the slave instance...")
			stopChan <- os.Interrupt
			<-stopChan
			connectionInfo.Hostname = pong.([]any)[1].(string)
			return OpenRedisConnection(cf, connectionInfo, appGUID, appName, true)
		}
	}
	outputUtils.PrintInterface(localCred)

	err = OpenTablePlusClient(REDIS, connectionInfo)

	if err != nil && runtime.GOOS == "darwin" {
		outputUtils.PrintErrorMessage("Error opening the connection:", err.Error())
	} else {
		// Wait for the user to close the connection
		utils.StringPrompt("press enter to close the connection...")
	}
	stopChan <- os.Interrupt
	return err
}

func OpenConnectionToService(cf *client.Client, serviceCredentials *ConnectionInfo, appGUID, serviceName, appName string) (chan os.Signal, error) {
	domain := utils.ExtractDomain(cf.Config.ApiURL(""))
	var localPort string

	switch serviceName {
	case POSTGRES:
		localPort = "5432"
	case REDIS:
		localPort = "6380"
	}

	if !isPortFree(localPort) {
		return nil, errors.New(fmt.Sprintf("Port %s is occupied!", localPort))
	}

	// stopChan is used for closing signal
	stopChan, err := createConnection(cf, serviceCredentials, appGUID, serviceName, domain, appName)
	return stopChan, err
}

func OpenTablePlusClient(dbType string, serviceCredentials *ConnectionInfo) error {
	var command string
	switch dbType {
	case POSTGRES:
		command = fmt.Sprintf("postgresql://%s:%s@127.0.0.1/%s?statusColor=686B6F&env=local&name=temp&tLSMode=0&usePrivateKey=false&safeModeLevel=0&advancedSafeModeLevel=0&driverVersion=0",
			serviceCredentials.Username, serviceCredentials.Password, serviceCredentials.Dbname)
	case REDIS:
		command = fmt.Sprintf("redis://:%s@127.0.0.1:6380?statusColor=686B6F&env=local&name=local&tLSMode=1&usePrivateKey=false&safeModeLevel=0&advancedSafeModeLevel=0&driverVersion=0&lazyload=true",
			serviceCredentials.Password)
	}

	var cmd string
	var args []string
	switch runtime.GOOS {
	case "windows":
		cmd = "cmd"
		args = []string{"/c", "start", command}
		//fmt.Println("Unfortunately, opening TablePlus is not supported on Windows")
		//fmt.Println("Please copy the following connection string and paste it in the browser")
		//outputUtils.PrintSuccessMessage(command)
	case "darwin":
		cmd = "open"
		args = []string{command}
	}
	err := exec.Command(cmd, args...).Run()
	if err != nil && runtime.GOOS == "windows" {
		err = nil
	}
	return err
}

func handleForwarding(client *ssh.Client, remoteHost string, remotePort string, listener net.Listener, openNewConnection chan bool) {
	conn, err := listener.Accept()
	if err != nil && strings.Contains(err.Error(), "use of closed network connection") {
		return
	} else if err != nil {
		fmt.Println("Error accepting connection: ", err)
		return
	}
	openNewConnection <- true
	defer conn.Close()

	// Create a remote connection
	remoteConn, err := client.Dial("tcp", fmt.Sprintf("%s:%s", remoteHost, remotePort))
	if err != nil {
		fmt.Println("Error dialing remote host: ", err)
		return
	}
	defer remoteConn.Close()

	go io.Copy(remoteConn, conn)
	io.Copy(conn, remoteConn)
}

func createConnection(cf *client.Client, serviceCredentials *ConnectionInfo, appGUID, serviceName, domain, appName string) (chan os.Signal, error) {
	port := "2222"
	var localPort string

	switch serviceName {
	case POSTGRES:
		localPort = "5432"
	case REDIS:
		localPort = "6380"
	}

	server := fmt.Sprintf("ssh.cf.%s", domain)
	process, err := cf.Processes.First(context.Background(), &client.ProcessListOptions{
		AppGUIDs: client.Filter{Values: []string{appGUID}},
	})
	if err != nil {
		return nil, err
	}
	user := fmt.Sprintf("cf:%s/0", process.GUID)
	password, err := cf.SSHCode(context.Background())
	if err != nil {
		return nil, err
	}
	config := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			ssh.Password(password),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}
	sshClient, err := ssh.Dial("tcp", fmt.Sprintf("%s:%s", server, port), config)
	if err != nil {
		if strings.Contains(err.Error(), "unable to authenticate") {
			res, internalErr := cf.Processes.GetStatsForApp(context.Background(), appGUID, "web")
			if internalErr != nil {
				return nil, internalErr
			}
			for _, instance := range res.Stats {
				if instance.State != "RUNNING" {
					return nil, errors.New("app is not running... please start the app and try again")
				}
			}
			outputUtils.PrintWarningMessage("you are not authorized to perform the requested action (maybe SSH access is off?)")
			ans := utils.QuestionPrompt("Do you want to enable ssh?")
			if !ans {
				return nil, err
			}
			err = applicationsUtils.EnableAppSsh(cf, appGUID)
			if err != nil {
				return nil, err
			}
			err = applicationsUtils.RestartAppRolling(cf, appGUID, appName)
			if err != nil {
				return nil, err
			}
			return createConnection(cf, serviceCredentials, appGUID, serviceName, domain, appName)
		} else {
			return nil, err
		}
	}

	// Remote host and port to forward to
	remoteHost := serviceCredentials.Hostname
	remotePort := serviceCredentials.Port

	// Create a listener on the local port
	listener, err := net.Listen("tcp", fmt.Sprintf(":%s", localPort))
	if err != nil {
		return nil, err
	}

	stop := false

	// Start a goroutine to handle incoming connections
	go func() {
		openNewConnection := make(chan bool, 1)
		openNewConnection <- true
		for !stop {
			if <-openNewConnection {
				go handleForwarding(sshClient, remoteHost, remotePort, listener, openNewConnection)
			}
		}
	}()

	stopChan := make(chan os.Signal, 1)

	go func() {
		signal.Notify(stopChan,
			os.Interrupt,
			syscall.SIGINT,
			syscall.SIGKILL,
			syscall.SIGTERM,
			syscall.SIGQUIT)
		<-stopChan
		stop = true
		homeDir, _ := os.UserHomeDir()
		if runtime.GOOS == "windows" {
			os.RemoveAll(filepath.Join(homeDir, "AppData", "Roaming", "postgresql"))
		} else {
			os.RemoveAll(filepath.Join(homeDir, ".postgresql"))
		}
		listener.Close()
		sshClient.Close()
		stopChan <- nil
	}()

	return stopChan, nil
}

func RunQuery(cred *ConnectionInfo, query string, forPrint bool) ([][]string, error) {
	databaseURL := fmt.Sprintf("postgresql://%s:%s@127.0.0.1:5432/%s",
		cred.Username, cred.Password, cred.Dbname)

	// Create a connection pool
	config, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, err
	}

	// Set connection pool settings (optional)
	config.MaxConns = 10 // max 10 connections
	config.MaxConnLifetime = 30 * time.Minute

	// Connect to the database
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	dbpool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, err
	}
	defer dbpool.Close()

	fmt.Println("Connected to the database successfully!")

	// Example Query: Fetch data from a table
	rows, err := dbpool.Query(ctx, query)
	defer rows.Close()

	if err != nil {
		return nil, err
	}

	columnDescriptions := rows.FieldDescriptions()
	numColumns := len(columnDescriptions)

	// Slice to store all rows
	var results [][]string
	if forPrint {
		columnNames := make([]string, numColumns)
		for i, column := range columnDescriptions {
			columnNames[i] = column.Name
		}
		results = append(results, columnNames)
	}

	// Iterate through each row
	for rows.Next() {
		// Create a slice to hold the values of the current row
		rowData := make([]interface{}, numColumns)
		rowPointers := make([]interface{}, numColumns)

		// Initialize the row pointers to point to each element in the rowData slice
		for i := range rowData {
			rowPointers[i] = &rowData[i]
		}

		// Scan the row into the rowPointers (which will populate rowData)
		err := rows.Scan(rowPointers...)
		if err != nil {
			return nil, fmt.Errorf("unable to scan the row: %v", err)
		}
		stringRow := make([]string, numColumns)
		for i, col := range rowData {
			if col != nil {
				stringRow[i] = fmt.Sprintf("%v", col)
				if strings.HasPrefix(stringRow[i], "[") {
					stringRow[i] = convertUUID(col)
				}
			} else {
				stringRow[i] = "NULL" // Handle null values
			}
		}

		// Append the scanned row (rowData) to the results
		results = append(results, stringRow)
	}

	// Check for any errors during the iteration
	if rows.Err() != nil {
		return nil, rows.Err()
	}
	return results, nil
}

func convertUUID(v interface{}) string {
	bytes, err := v.([16]byte)
	if !err {
		return ""
	}
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%04x%08x",
		uint32(bytes[0])<<24|uint32(bytes[1])<<16|uint32(bytes[2])<<8|uint32(bytes[3]),
		uint16(bytes[4])<<8|uint16(bytes[5]),
		uint16(bytes[6])<<8|uint16(bytes[7]),
		uint16(bytes[8])<<8|uint16(bytes[9]),
		uint16(bytes[10])<<8|uint16(bytes[11]),
		uint32(bytes[12])<<24|uint32(bytes[13])<<16|uint32(bytes[14])<<8|uint32(bytes[15]),
	)
}

func PrintQueryResult(rows [][]string) {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader(rows[0])
	rows = rows[1:]
	//for i := 0; i < len(rows); i++ {
	//	for j := 0; j < len(rows[i]); j++ {
	//		str := ""
	//		s := 0
	//		for ; s < len(rows[i][j])-40; s += 40 {
	//			str += rows[i][j][s:s+40] + "\n"
	//		}
	//		str += rows[i][j][s:]
	//		rows[i][j] = str
	//	}
	//}
	table.SetRowLine(true)
	table.SetColWidth(40)
	table.SetCenterSeparator("")
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetTablePadding("\n") // pad with tabs
	table.AppendBulk(rows)
	table.Render()

}

func isPortFree(port string) bool {
	address := ":" + port
	ln, err := net.Listen("tcp", address)
	if err != nil {
		return false // Port is not free
	}
	_ = ln.Close() // Release the port
	return true
}

func GetPostgresConnectionInfo(cred map[string]interface{}) (*ConnectionInfo, error) {
	info := &ConnectionInfo{
		Hostname: cred["hostname"].(string),
		Port:     cred["port"].(string),
		Username: cred["username"].(string),
		Password: cred["password"].(string),
		Dbname:   cred["dbname"].(string),
	}
	if cred["server_ca"] != nil {
		err := SaveCertsToPostgresDir(cred["sslkey"].(string), cred["sslcert"].(string), cred["server_ca"].(string))
		return info, err
	}
	return info, nil

}

func SaveCertsToPostgresDir(sslkey, sslcert, sslrootcert string) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get user home directory: %v", err)
	}

	var postgresDir string
	if runtime.GOOS == "windows" {
		postgresDir = filepath.Join(homeDir, "AppData", "Roaming", "postgresql")
	} else {
		postgresDir = filepath.Join(homeDir, ".postgresql")
	}
	if err := os.MkdirAll(postgresDir, 0700); err != nil {
		return fmt.Errorf("failed to create .postgresql directory: %v", err)
	}

	certPath := filepath.Join(postgresDir, "postgresql.crt")
	if err := os.WriteFile(certPath, []byte(sslcert), 0600); err != nil {
		return fmt.Errorf("failed to write client.crt: %v", err)
	}

	keyPath := filepath.Join(postgresDir, "postgresql.key")
	if err := os.WriteFile(keyPath, []byte(sslkey), 0600); err != nil {
		return fmt.Errorf("failed to write client.key: %v", err)
	}

	rootCertPath := filepath.Join(postgresDir, "root.crt")
	if err := os.WriteFile(rootCertPath, []byte(sslrootcert), 0600); err != nil {
		return fmt.Errorf("failed to write root.crt: %v", err)
	}

	return nil
}
