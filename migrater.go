package main

import (
	"fmt"
	"flag"
	"os"
	"io/ioutil"
	"syscall"
	"regexp"
	"strings"
	"strconv"
	"bufio"
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
		createNewMigration();
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

func createNewMigration() (bool, error) {
	folderExist,   _ := exists("./sqls")
	if(folderExist == false){
		fmt.Println("\"sqls\" folder does not exist please use \"migrater -init\" command to initialize")
		return false, nil;
	}

	if( isWritable("./sqls") == false) {
		fmt.Println("sqls folder is not writable. Please make it writable and try again");
		return false, nil
	}
	
	fileList, _ := ioutil.ReadDir("./sqls/")
	
	regx, _ := regexp.Compile("^[0-9]{4}_")
	
	var counter, preFileNum int64 = 1, 0;
	
	for _, f := range fileList{
		var fileName = f.Name()
		match := regx.FindString(fileName)
		
		if(match == "") {
			continue
		}	

		match = strings.Replace(match, "_", "", 1)
		fileNum, _ := strconv.ParseInt(match, 10, 64)
		
		fmt.Println(fileNum, counter)
		if(preFileNum == fileNum) {
			fmt.Printf("%04d_* file has a duplicate entry. Please remove duplicates. \n", fileNum)
			return false, nil
		} else if(counter != fileNum){
			fmt.Printf("%04d_*.sql file is missing\n", counter);
			return false, nil
		}
		preFileNum = fileNum;
		counter++
	}

	reader := bufio.NewReader(os.Stdin)
	fmt.Println("New .sql file description :")
	fileDesc, _ := reader.ReadString('\n')
	fmt.Println(fileDesc)
	
    fileDesc = strings.ToLower(strings.Trim(fileDesc, ""))
	reg, _ := regexp.Compile("[^A-Za-z0-9]+")
    newFileDesc := reg.ReplaceAllString(fileDesc, "_")
	newFileName := fmt.Sprintf("%04d_%s.sql", counter, newFileDesc)
	
	var content = []byte("[up]\n;sql up queries here\n\n\n[down]\n sql down queries here")
	var err = ioutil.WriteFile("./sqls/"+newFileName, content, 0755)
    if err != nil {
        panic(err)
    }
    fmt.Println("New file created ./sqls/"+newFileName)	
	
	return true, nil
	
	//check sqls directory exist and writable
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

func isWritable(path string) bool {
    return syscall.Access(path, 2) == nil
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