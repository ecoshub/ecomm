package main

import (
	"database/sql"
	"io/ioutil"
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

	// json format schemes
	responseScheme *jin.Scheme

	// log strings
	srvStart    string = ">> Data Service Started."
	srvEnd      string = ">> Data Service Shutdown Unexpectedly. Error:"
	reqArrived  string = ">> Request Arrived At"
	reqBody     string = ">> Request Body:"
	statError   string = ">> Status method not allowed"
	dataFailed  string = ">> Data Request Failed:"
	dataGranted string = ">> Data Request Granted."
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
	// response scheme
	responseScheme = jin.MakeScheme("status", "error")
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

	// method check
	if string(r.Method) != http.MethodPost {
		log.Println(dataFailed, statError)
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write(statusFailed(statError))
		return
	}
	// body read for json parse.
	json, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Println(dataFailed, err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(statusFailed(err.Error()))
		return
	}
	action, err := jin.GetString(json, "action")
	if err != nil {
		log.Println(dataFailed, err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(statusFailed(err.Error()))
		return
	}

	switch action {
	case "insert":
		body, err := jin.Get(json, "body")
		if err != nil {
			log.Println(dataFailed, err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write(statusFailed(err.Error()))
			return
		}
		err = insertRecord(body)
		if err != nil {
			log.Println(dataFailed, err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write(statusFailed(err.Error()))
			return
		}
		log.Println(dataGranted)
		w.WriteHeader(http.StatusOK)
		w.Write(statusGranted())
	}
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

func statusFailed(err string) []byte {
	return responseScheme.MakeJson("Failed", seecool.EscapeQuote(err))
}

func statusGranted() []byte {
	return responseScheme.MakeJson("OK", "null")
}
