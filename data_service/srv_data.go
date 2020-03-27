package main

import (
	"database/sql"
	"jin"
	"log"
	"net/http"
	"penman"
	"seecool"

	_ "github.com/lib/pq"
)

var (
	// environment directories,
	// 'curr' keyword is a wild card for 'currentDirectory'
	// valid wild card with 'github.com/ecoshub/penman' package
	envDatabaseDir string = "curr/.env_database"
	envServiceDir  string = "curr/.env_service"
	envMainDir     string = "curr/../.env_main"

	// service environment map
	envServiceMap map[string]string

	// main environment environment map
	envMainMap map[string]string

	// database environment file
	dbEnv string

	// data service main port
	mainPort string

	// main database pointer
	base *sql.DB

	// log strings
	srvStart    string = ">> Data Service Started."
	srvEnd      string = ">> Data Service Shutdown Unexpectedly. Error:"
	reqArrived  string = ">> Request Arrived At"
	reqBody     string = ">> Request Body:"
	statError   string = ">> Status method not allowed"
	dataFailed  string = ">> Data Failed:"
	dataGranted string = ">> Data Granted."
)

func init() {
	var err error
	// read main env. file
	envMainMap, err = seecool.GetEnv(envMainDir)
	if err != nil {
		panic(err)
	}
	mainPort = envMainMap["data_service_port"]
	// read env_service file
	envServiceMap, err = seecool.GetEnv(envServiceDir)
	if err != nil {
		panic(err)
	}
	// read env_database file
	dbEnv = penman.SRead(envDatabaseDir)
	if dbEnv == "" {
		panic(dbEnv)
	}
}

func main() {
	dbConn()
	log.Println(srvStart, "port:", mainPort)
	http.HandleFunc("/", dataHandle)
	err := http.ListenAndServe(":"+mainPort, nil)
	// handle later
	log.Println(srvEnd, err)
	defer base.Close()
}

func dataHandle(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Origin, cache-control")
	w.Header().Set("Content-Type", "application/json")
	// request log
	log.Println(reqArrived, r.RemoteAddr)

}

// json := []byte(`{"username":"ozcanocak","password":"ozcanocak","email":"ozcanocak@gmail.com"}`)
// err := insertRecord(json)
// fmt.Println(err)

func dbConn() {
	var err error
	base, err = sql.Open(envServiceMap["user"], dbEnv)
	if err != nil {
		panic(err)
	}
}

func insertRecord(json []byte) error {
	keys, values, err := jin.GetKeysValues(json)
	if err != nil {
		return err
	}
	query := seecool.Insert(envServiceMap["table"]).
		Keys(keys...).
		Values(values...)
	_, err = base.Query(query.String())
	if err != nil {
		return err
	}
	return nil
}
