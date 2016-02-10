package main
//	"flag"

import (
	"fmt"
	"os"
	"io/ioutil"
)

func main() {
	 var initPtr = true
//	initPtr := flag.Bool("init", false, "a string")
//	flag.Parse()	
	
	if (initPtr == true) {
		var folderExist = false
		var confFileExist = false
//		folderExist, err   := exists("./sqls")
//		confFileExist, err := exists("./migrater.conf")
//		if(err == nil) {
//			fmt.Println("Some error occured")
//		}
		
		if(folderExist == false && confFileExist == false) { //folder, conf does not exist
			fmt.Println("sqls AND conf does not exist -- creating")
			createConfFile()
			createSqlsFolder()
		} else if(folderExist == true && confFileExist == false) {//folder exist, conf not
			fmt.Println("folder exist, file not")
			createConfFile()
		} else if (folderExist == false && confFileExist == true) {
			createSqlsFolder()
			fmt.Println("SQL folder not exist, conf does exist")
		} else { //both exist
			fmt.Println("Already an migrater directory")
		}
//		fmt.Println("Value:", *initPtr)	
	}
}

func initAction() {
	
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

migrater -init : will create "sqls" directory, migrater.conf file it will have connection info for databas
migrater -new : should be run from valid migrater directory, having "sqls" folder, it will create .sql 
	file in sqls directory, will take first 50 chars of desc and append with name like 
	0002_desc_of_file.sql, it will also write desc in file itself at start of file commented.
migrater -update : will run all the migration against database
migrater -v 5 : it will up or down the migrations from sqls folder
how am I going to use conf 
**/