package main

import _ "github.com/go-sql-driver/mysql"

import (
	"bufio"
	"database/sql"
	"errors"
	"flag"
	"fmt"
	"github.com/vaughan0/go-ini"
	"io/ioutil"
	"os"
	"regexp"
	"strconv"
	"strings"
	"syscall"
)

var db *sql.DB

func main() {

	configFileName := "config.ini"

	initPtr := flag.Bool("init", false, "-init")
	newPtr := flag.Bool("new", false, "-new")
	updatePtr := flag.Bool("up", false, "-up")

	flag.Parse()
	//	fmt.Println("value :", *newPtr)

	if *initPtr == true {
		initAction()
		return
	} else if *newPtr == true {
		createNewMigration()
	} else if *updatePtr == true {
		err := updateMigrations(configFileName)
		if err != nil {
			fmt.Println(err)
		}
	}
	fmt.Println("\n")
}

func initAction() bool {
	folderExist, _ := exists("./sqls")
	confFileExist, _ := exists("./migrater.conf")

	if folderExist == false && confFileExist == false { //folder, conf does not exist
		fmt.Println("Initializing migrater folder...")
		createConfFile()
		createSqlsFolder()
	} else if folderExist == true && confFileExist == false { //folder exist, conf not
		fmt.Println("Initializing migrater folder...")
		createConfFile()
	} else if folderExist == false && confFileExist == true {
		createSqlsFolder()
		fmt.Println("Initializing migrater folder...")
	} else { //both exist
		fmt.Println("Already an migrater directory")
	}
	return true
}

func getDBConnString(dbConfig map[string]string) (string, error) {
	if dbConfig["dbtype"] == "mysql" {
		//		root:password@tcp(localhost:3306)/temp
		dbConnString := dbConfig["username"] + ":" + dbConfig["password"] + "@tcp(" + dbConfig["hostname"] + ":" + dbConfig["port"] + ")/" + dbConfig["dbname"]
		return dbConnString, nil
	}
	return "", errors.New("Invalid datatype")
}

func existMigrationTable(dbConn *sql.DB, dbConfig map[string]string, migrTableName string) {
	query := "SELECT COUNT(*) as tableExist FROM information_schema.tables WHERE table_schema = '" + dbConfig["dbname"] + "'  AND table_name = '" + migrTableName + "'"
	rows, err := dbConn.Query(query)
	if err != nil {
		fmt.Println(err)
	}
	var tableExist int

	next := rows.Next()
	if next == false {
		fmt.Println("Some error occurred while checking migrations table exist in db")
		os.Exit(1)
	}

	exist := rows.Scan(&tableExist)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if exist == nil && tableExist == 0 { //migraion table not exist, create it
		var createTableQuery = "CREATE TABLE IF NOT EXISTS " + migrTableName + "(version int(10), description varchar(200) NOT NULL, sql_file varchar(200) NOT NULL,  created_on datetime NOT NULL,  PRIMARY KEY (version))"
		_, err := dbConn.Query(createTableQuery)
		if err == nil {
			return
		}
		fmt.Println("Error occurred while creating migrations table:", err)
		os.Exit(1)
	}
}

func exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return true, err
}

func createNewMigration() (bool, error) {
	folderExist, _ := exists("./sqls")
	if folderExist == false {
		fmt.Println("\"sqls\" folder does not exist please use \"migrater -init\" command to initialize")
		return false, nil
	}

	if isWritable("./sqls") == false {
		fmt.Println("sqls folder is not writable. Please make it writable and try again")
		return false, nil
	}

	fileList, _ := ioutil.ReadDir("./sqls/")
	regx, _ := regexp.Compile("^[0-9]{4}_")

	var counter, preFileNum int64 = 1, 0

	for _, f := range fileList {
		var fileName = f.Name()
		match := regx.FindString(fileName)

		if match == "" {
			continue
		}

		match = strings.Replace(match, "_", "", 1)
		fileNum, _ := strconv.ParseInt(match, 10, 64)

		if preFileNum == fileNum {
			fmt.Printf("%04d_* file has a duplicate entry. Please remove duplicates. \n", fileNum)
			return false, nil
		} else if counter != fileNum {
			fmt.Printf("%04d_*.sql file is missing\n", counter)
			return false, nil
		}
		preFileNum = fileNum
		counter++
	}

	reader := bufio.NewReader(os.Stdin)
	fmt.Println("\nNew .sql file description :")
	fileDesc, _ := reader.ReadString('\n')

	fileDesc = strings.ToLower(strings.Trim(fileDesc, "\n"))
	reg, _ := regexp.Compile("[^A-Za-z0-9]+")
	newFileDesc := reg.ReplaceAllString(fileDesc, "_")
	newFileName := fmt.Sprintf("%04d_%s.sql", counter, newFileDesc)

	var content = []byte("[up]\n;sql up queries here\n\n\n[down]\n sql down queries here")
	var err = ioutil.WriteFile("./sqls/"+newFileName, content, 0755)
	if err != nil {
		panic(err)
	}
	fmt.Println("\n--------------------------------------------------------------------------------------------")
	//	fmt.Println("New file created at ./sqls/" + newFileName + ", Write your new SQL up and down statements in it.")
	fmt.Println("New file created at ./sqls/" + newFileName + ", Write your new SQL statements in it.")
	fmt.Println("--------------------------------------------------------------------------------------------\n")
	return true, nil
}

func createConfFile() (bool, error) {
	var configContent = `[database]
dbtype   = mysql
dbname   = 
hostname = localhost
port     = 3306
username = 
password = `

	var content = []byte(configContent)
	var err = ioutil.WriteFile("./config.ini", content, 0755)
	if err != nil {
		panic(err)
	}
	return true, nil
}

func createSqlsFolder() (bool, error) {
	var err = os.Mkdir("sqls", 0755)
	if err != nil {
		panic(err)
	}
	return true, nil
}

func isWritable(path string) bool {
	return syscall.Access(path, 2) == nil
}

func printMsgLine(msg string, msgType string) {
	if msgType == "error" {
		fmt.Println(msg) //print in red color
		os.Exit(1)
	}
	fmt.Println(msg)
	return
}

func getConfigValues(configFileName string) map[string]string {

	config, err := ini.LoadFile(configFileName)
	if err != nil {
		fmt.Println("Error : ", err)
		os.Exit(1)
	}

	dbConfig := map[string]string{"dbtype": "", "dbname": "", "hostname": "", "port": "", "username": "", "password": ""}

	confCheckError := false
	for confKey, _ := range dbConfig {
		value, ok := config.Get("database", confKey)
		val := strings.Trim(value, " ")

		if !ok {
			fmt.Println(confKey + " entry is missing in " + configFileName)
			confCheckError = true
		} else if val == "" {
			fmt.Println(confKey + " value can not be blank in " + configFileName)
			confCheckError = true
		}
		dbConfig[confKey] = value
	}

	if confCheckError == true {
		os.Exit(1)
	}

	if dbConfig["dbtype"] != "mysql" {
		printMsgLine("Unsupported database, Currently migration tool only support 'mysql' database.", "error")
		os.Exit(1)
	}
	return dbConfig
}

func updateMigrations(configFileName string) error {
	migrTableName := "db_migrations"
	dbConfig := getConfigValues(configFileName)
	dbConnString, _ := getDBConnString(dbConfig)
	db, err := sql.Open(dbConfig["dbtype"], dbConnString)
	if err != nil {
		return err
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		return errors.New("DB Error : " + err.Error())
	}

	existMigrationTable(db, dbConfig, migrTableName)

	getTopVersionQuery := "SELECT IFNULL(MAX(version), 0) as curVersion FROM " + migrTableName
	rows, err := db.Query(getTopVersionQuery)

	if err != nil {
		return err
	}

	var curVersion int64

	for rows.Next() {
		err := rows.Scan(&curVersion)
		if err != nil {
			return err
		}
		break
	}

	folderExist, _ := exists("./sqls")
	if folderExist == false {
		return errors.New("\"sqls\" folder does not exist please use \"migrater -init\" command to initialize")
	} else if isWritable("./sqls") == false {
		return errors.New("sqls folder is not writable. Please make it writable and try again")
	}

	fileList, _ := ioutil.ReadDir("./sqls/")
	regx, _ := regexp.Compile("^[0-9]{4}_")

	var counter, preFileNum int64 = 1, 0
	var sqlFiles []string

	for _, f := range fileList { //loop for checking duplicate and missing sql files, code is redundant in createMigraion script too, refactor
		var fileName = f.Name()
		match := regx.FindString(fileName)
		if match == "" {
			continue
		}

		match = strings.Replace(match, "_", "", 1)
		fileNum, _ := strconv.ParseInt(match, 10, 64)

		if preFileNum == fileNum {
			var errorMsg = fmt.Sprintf("%04d_* file has a duplicate entry. Please remove duplicates. \n", fileNum)
			return errors.New(errorMsg)
		} else if counter != fileNum {
			var errorMsg = fmt.Sprintf("%04d_*.sql file is missing\n", counter)
			return errors.New(errorMsg)
		}
		preFileNum = fileNum
		counter++
		sqlFiles = append(sqlFiles, fileName)
	}

	counter--
	topMigrationVersion := preFileNum
	nextVersion := curVersion + 1

	if counter == 1 && curVersion == 0 {
		fmt.Println("Running all migrations from start...")
	} else {
		if curVersion == topMigrationVersion {
			fmt.Println("Database is up to date, no new migration to run!")
			fmt.Println("Current database version: ", topMigrationVersion)
			return nil
		} else if nextVersion == topMigrationVersion {
			fmt.Printf("Running migration version %04d \n", nextVersion)
		} else {
			fmt.Printf("Running migrations from version %04d to %04d\n", nextVersion, topMigrationVersion)
		}
	}

	for _, sqlFileName := range sqlFiles {
		match := regx.FindString(sqlFileName)
		match = strings.Replace(match, "_", "", 1)
		fileNum, _ := strconv.ParseInt(match, 10, 64)

		if fileNum < nextVersion {
			continue
		}
		fmt.Println("Running queries from ", sqlFileName, "file")
		file, err := os.Open("./sqls/" + sqlFileName)

		if err != nil {
			fmt.Println(err)
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		if err := scanner.Err(); err != nil {
			return err
		}

		var query, queryLine, queryLineTrimmed string
		queryEndRegEx, _ := regexp.Compile(";$")

		for scanner.Scan() {
			queryLine = scanner.Text()
			queryLineTrimmed = strings.Trim(queryLine, " ")
			match := queryEndRegEx.FindString(queryLineTrimmed)
			if match == "" {
				query += queryLine
			} else { //line ends with semicolon, query
				query += queryLine
				_, err := db.Exec(query)
				if err != nil {
					fmt.Println("\nError in file : ./sqls/"+sqlFileName, ", Unfortunately mysql doesn't support rollback(NO support for Transactional DDL queries) for prevouly exceuted queries in this file.")
					return err
				}
				query = ""
				queryLine = ""
			}
		}

		regxVersion, _ := regexp.Compile("^([0-9]{4})")
		migrVersion := regxVersion.FindString(sqlFileName)

		regxDesc, _ := regexp.Compile("([0-9]{4}_|.sql)")
		migrDescBytes := regxDesc.ReplaceAll([]byte(sqlFileName), []byte(""))
		migrDesc := strings.Replace(string(migrDescBytes), "_", " ", -1)

		updateMigrQuery := "INSERT INTO " + migrTableName + "(version, description, sql_file, created_on) VALUES ('" + migrVersion + "', '" + migrDesc + "', '" + sqlFileName + "', now())"
		_, updateErr := db.Exec(updateMigrQuery)
		if updateErr != nil {
			return err
		}
	}
	fmt.Println("Current database version: ", topMigrationVersion)
	return nil
}

/**


@todo : when no parameters given to command print manual page

export GOPATH=`pwd`
go build migrater.go && ./migrater -init

migrater -init : will create "sqls" directory, migrater.conf file it will have connection info for databas
migrater -new : should be run from valid migrater directory, having "sqls" folder, it will create .sql
	file in sqls directory, will take first 50 chars of desc and append with name like
	0002_desc_of_file.sql, it will also write desc in file itself at start of file commented.
migrater -up : will run all the migration against database
--migrater -v 5 : it will up or down the migrations from sqls folder
how am I going to use conf

mysql drivers used : https://github.com/go-sql-driver/mysql
http://engineroom.teamwork.com/go-learn/
for db http://jinzhu.me/gorm/
**/
