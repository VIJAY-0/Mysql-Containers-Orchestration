package MysqlServer

import (
	"MYSQL-orchestration-API/config"
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

func waitForMySQL(dsn string, retries int, delay time.Duration) (*sql.DB, error) {
	var db *sql.DB
	var err error

	for i := 0; i < retries; i++ {
		fmt.Printf("Attempting to connect to MySQL (attempt %d/%d)...\n", i+1, retries)
		db, err = sql.Open("mysql", dsn)
		if err == nil {
			// Try pinging the database
			err = db.Ping()
			if err == nil {
				fmt.Println("Successfully connected to MySQL.")
				return db, nil
			}
		}
		fmt.Printf("MySQL not ready, retrying in %s...\n", delay)
		time.Sleep(delay)
	}
	return nil, fmt.Errorf("could not connect to MySQL after %d attempts: %v", retries, err)
}

func InitMaster(Master map[string]*config.ServerConfig, MasterID string) {

	fmt.Println("Initiation Started...")

	master := Master[MasterID]
	username := "root"
	password := master.Environment.MYSQL_ROOT_PASSWORD
	host := "localhost"
	port := master.Ports[0][0:4]

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/", username, password, host, port)
	db, err := waitForMySQL(dsn, 10, 5*time.Second)
	// db, err = sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("Error opening database: %v", err)
	}
	defer db.Close()

	log.Println("Connected to Master database")

	sqlQuery := "DROP USER IF EXISTS 'replicator'@'%';"
	_, err = db.Exec(sqlQuery)
	sqlQuery = "CREATE USER 'replicator'@'%' IDENTIFIED WITH 'mysql_native_password' BY 'replicatorpassword';"
	_, err = db.Exec(sqlQuery)
	sqlQuery = "GRANT REPLICATION SLAVE ON *.* TO 'replicator'@'%';"
	_, err = db.Exec(sqlQuery)
	sqlQuery = "FLUSH PRIVILEGES;"
	_, err = db.Exec(sqlQuery)

	sqlQuery = "SHOW MASTER STATUS;"
	rows, err := db.Query(sqlQuery)

	// Print the result
	var file string
	var position string
	var binlogDoDB string
	var binlogIgnoreDB string
	var executedGtidSet string

	fmt.Println("Master Status:")
	for rows.Next() {
		err := rows.Scan(&file, &position, &binlogDoDB, &binlogIgnoreDB, &executedGtidSet)
		if err != nil {
			log.Fatalf("Error scanning row: %v\n", err)
		}

		fmt.Printf("File: %s\nPosition: %s\nBinlog_Do_DB: %s\nBinlog_Ignore_DB: %s\nExecuted_Gtid_Set: %s\n",
			file, position, binlogDoDB, binlogIgnoreDB, executedGtidSet)
	}

	if err := rows.Err(); err != nil {
		log.Fatalf("Error iterating rows: %v\n", err)
	}

	fmt.Println("Initiation Completed #")
}

func GetDSN(serverconfig *config.ServerConfig) string {

	master := serverconfig
	host := "localhost"
	port := master.Ports[0][0:4]
	dsn := fmt.Sprintf("%s:%s", host, port)
	return dsn
}

func getMasterStatus(Master map[string]*config.ServerConfig, MasterID string) (string, string, error) {

	master := Master[MasterID]
	username := "root"
	password := master.Environment.MYSQL_ROOT_PASSWORD
	host := "localhost"
	port := master.Ports[0][0:4]

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/", username, password, host, port)
	db, err := waitForMySQL(dsn, 10, 5*time.Second)
	// db, err = sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("Error opening database: %v", err)
	}
	defer db.Close()
	log.Println("Connected to Master database")

	sqlQuery := "SHOW MASTER STATUS;"
	rows, err := db.Query(sqlQuery)

	// Print the result
	var file string
	var position string
	var binlogDoDB string
	var binlogIgnoreDB string
	var executedGtidSet string

	fmt.Println("Master Status:")
	for rows.Next() {
		err := rows.Scan(&file, &position, &binlogDoDB, &binlogIgnoreDB, &executedGtidSet)
		if err != nil {
			log.Fatalf("Error scanning row: %v\n", err)
		}

		fmt.Printf("File: %s\nPosition: %s\nBinlog_Do_DB: %s\nBinlog_Ignore_DB: %s\nExecuted_Gtid_Set: %s\n",
			file, position, binlogDoDB, binlogIgnoreDB, executedGtidSet)
		return file, position, nil
	}

	if err := rows.Err(); err != nil {
		log.Fatalf("Error iterating rows: %v\n", err)
	}
	return "", "", err
}
func InitSlave(Master map[string]*config.ServerConfig, MasterID string, slaveconfig *config.ServerConfig, SlaveCntr int, SlaveDSNs *[]string) {

	fmt.Println("Initiation Started...")
	slave := slaveconfig
	username := "root"
	password := slave.Environment.MYSQL_ROOT_PASSWORD
	host := "localhost"
	port := slave.Ports[0][0:4]

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/", username, password, host, port)
	dsn_ := fmt.Sprintf("%s:%s", host, port)
	(*SlaveDSNs) = append((*SlaveDSNs), dsn_)
	db, err := waitForMySQL(dsn, 10, 5*time.Second)
	if err != nil {
		log.Fatalf("Error opening database: %v", err)
	}
	defer db.Close()

	MasterLogFile, MasterLogPos, err := getMasterStatus(Master, MasterID)
	if err != nil {
		log.Println("Error connecting to master...")
		return
	}

	commands := []string{
		fmt.Sprintf("SET GLOBAL server_id = %v;", SlaveCntr),
		"STOP SLAVE;",
		"CHANGE MASTER TO " +
			fmt.Sprintf("MASTER_HOST='%s-server', ", MasterID) +
			"MASTER_USER='replicator', " +
			"MASTER_PASSWORD='replicatorpassword'," +
			fmt.Sprintf("MASTER_LOG_FILE='%s', ", MasterLogFile) +
			fmt.Sprintf("MASTER_LOG_POS=%v;", MasterLogPos),
		"START SLAVE;",
	}

	for _, cmd := range commands {
		_, err := db.Exec(cmd)
		if err != nil {
			log.Fatalf("Error executing command [%s]: %v\n", cmd, err)
		}
		fmt.Printf("Successfully executed: %s\n", cmd)
	}

	// Check slave status
	rows, err := db.Query("SHOW SLAVE STATUS;")

	columns, _ := rows.Columns()
	values := make([]sql.RawBytes, len(columns))
	scanArgs := make([]interface{}, len(values))

	for i := range values {
		scanArgs[i] = &values[i]
	}

	for rows.Next() {
		err = rows.Scan(scanArgs...)
		if err != nil {
			log.Fatalf("Error scanning row: %v\n", err)
		}

		fmt.Println("Slave Status:")
		for i, col := range columns {
			fmt.Printf("%s: %s\n", col, values[i])
		}
	}
}
