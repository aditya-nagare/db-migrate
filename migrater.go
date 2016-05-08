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

func main() {

	configFileName := "config.ini"
	migrTableName := "db_migrations"
	config, err := ini.LoadFile(configFileName)
	if err != nil {
		fmt.Println(err)
		return
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
		return
	}

	if dbConfig["dbtype"] != "mysql" {
		printMsgLine("Unsupported database, Currently migration tool only support 'mysql' database.", "error")
		return
	}

	dbConnString, _ := getDBConnString(dbConfig)
	db, err := sql.Open(dbConfig["dbtype"], dbConnString)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		panic(err.Error()) // proper error handling instead of panic in your app
	}

	result := existMigrationTable(db, dbConfig, migrTableName)
	fmt.Println(result)
	return
	rows, err := db.Query("SELECT id, name FROM test")
	if err != nil {
		panic(err.Error()) // proper error handling instead of panic in your app
	}

	var (
		id   int
		name string
	)

	for rows.Next() {
		err := rows.Scan(&id, &name)
		if err != nil {
			//			fmt.Fatal(err)
		}
		fmt.Println(id, name)
	}
	err = rows.Err()
	fmt.Println(err)

	//	columns1, _ := rows.Columns()
	//    fmt.Printf("value of id : %d, Name : %s  ", 1,  columns1.name)
	return

	initPtr := flag.Bool("init", false, "-init")
	newPtr := flag.Bool("new", false, "-new")

	flag.Parse()
	//	fmt.Println("value :", *newPtr)

	if *initPtr == true {
		initAction()
		return
	} else if *newPtr == true {
		createNewMigration()
	}
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

//func getDBConnection(dbConfig map[string]string) (bool, error) {
//	fmt.Println("in func")
//	fmt.Println(dbConfig)
//	return true, nil
//	db, err := sql.Open("mysql", "root:password@/temp")
//	if err != nil {
//	    panic(err.Error()) // Just for example purpose. You should use proper error handling instead of panic
//	}
//}

func getDBConnString(dbConfig map[string]string) (string, error) {
	if dbConfig["dbtype"] == "mysql" {
		//		root:password@tcp(localhost:3306)/temp
		dbConnString := dbConfig["username"] + ":" + dbConfig["password"] + "@tcp(" + dbConfig["hostname"] + ":" + dbConfig["port"] + ")/" + dbConfig["dbname"]
		return dbConnString, nil
	}
	return "", errors.New("Invalid datatype")
}

func existMigrationTable(dbConn *sql.DB, dbConfig map[string]string, migrTableName string) bool {
	query := "SELECT COUNT(*) as tableExist FROM information_schema.tables WHERE table_schema = '" + dbConfig["dbname"] + "'  AND table_name = '" + migrTableName + "'"
	rows, err := dbConn.Query(query)
	if err != nil {
		fmt.Println(err)
	}
	var tableExist int

	next := rows.Next()
	if next == false {
		return false
	}

	exist := rows.Scan(&tableExist)
	if err != nil {
		fmt.Println(err)
	}
	if tableExist == 0 {
		//migraions table not found in db, create new with appropriate columnsable
		return false
	}
	return true
	fmt.Println(exist)
	return true
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
	fmt.Println("New file created at ./sqls/" + newFileName + ", Write your new SQL up and down statements in it.")
	fmt.Println("--------------------------------------------------------------------------------------------\n")

	return true, nil
}

func createConfFile() (bool, error) {
	var content = []byte("content configuration")
	var err = ioutil.WriteFile("./settings.conf", content, 0755)
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
		return
	}
	fmt.Println(msg)
}

/**
export GOPATH=`pwd`
go build migrater.go && ./migrater -init

migrater -init : will create "sqls" directory, migrater.conf file it will have connection info for databas
migrater -new : should be run from valid migrater directory, having "sqls" folder, it will create .sql
	file in sqls directory, will take first 50 chars of desc and append with name like
	0002_desc_of_file.sql, it will also write desc in file itself at start of file commented.
migrater -update : will run all the migration against database
migrater -v 5 : it will up or down the migrations from sqls folder
how am I going to use conf

mysql drivers used : https://github.com/go-sql-driver/mysql
**/
