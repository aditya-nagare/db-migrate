package main

import (
	"fmt"
	"flag"
	"os"
	"io/ioutil"
)

func main() {
	
	initPtr := flag.Bool("init", false, "-init")
	newPtr  := flag.Bool("new", false, "-new")
	
	flag.Parse()	
	fmt.Println("value :", *newPtr)
	
	if (*initPtr == true) {
		initAction()
		return	
	} else if(*newPtr == true) {
		fmt.Println("New command called")
	}
}

func initAction() bool {
	folderExist,   _ := exists("./sqls")
	confFileExist, _ := exists("./migrater.conf")
	
	if(folderExist == false && confFileExist == false) { //folder, conf does not exist
		fmt.Println("Initializing migrater folder...")
		createConfFile()
		createSqlsFolder()
	} else if(folderExist == true && confFileExist == false) {//folder exist, conf not
		fmt.Println("Initializing migrater folder...")
		createConfFile()
	} else if (folderExist == false && confFileExist == true) {
		createSqlsFolder()
		fmt.Println("Initializing migrater folder...")
	} else { //both exist
		fmt.Println("Already an migrater directory")
	}
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

func createNewMigration() {
	//check sqls directory exist
	//check is there any file with pattern dddd_name exist 
	//take use of file from user
	//create mock file with temp data in it
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
/**
go build migrater.go && ./migrater -init

migrater -init : will create "sqls" directory, migrater.conf file it will have connection info for databas
migrater -new : should be run from valid migrater directory, having "sqls" folder, it will create .sql 
	file in sqls directory, will take first 50 chars of desc and append with name like 
	0002_desc_of_file.sql, it will also write desc in file itself at start of file commented.
migrater -update : will run all the migration against database
migrater -v 5 : it will up or down the migrations from sqls folder
how am I going to use conf 
**/