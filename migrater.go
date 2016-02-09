package main

import (
	"fmt"
	"flag"
	"os"
)

func main() {
	initPtr := flag.Bool("init", false, "a string")
	flag.Parse()	
	
	if (*initPtr == true) {
		folderExist, err   := exists("./sqls")
		confFileExist, err := exists("./migrater.conf")
		
		if(folderExist == false && confFileExist == false) { //folder, conf does not exist
			fmt.Println("sqls AND conf does not exist -- creating")
		} else if(folderExist == true && confFileExist == false) {//folder exist, conf not
			fmt.Println("Both sqls and conf exist")
		} else if (folderExist == true && confFileExist == false) {
			fmt.Println("SQL folder exist, conf does not exist create it")
		}
		fmt.Println("Value:", *initPtr)	
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

/**

migrater -init : will create "sqls" directory, migrater.conf file 
migrater -new : should be run from valid migrater directory, having "sqls" and migrater.conf file
migrater -update : will run all the migration against database
migrater -v 5 : it will up or down the migrations from sqls folder
how am I going to use conf 
**/