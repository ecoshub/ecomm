package main

import (
	"database/sql"
	"errors"
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
	// valid wildcard can be user with 'github.com/ecoshub/penman' package
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

	// errors
	recordNotExists error = errors.New("Record does not exists.")

	// log strings
	srvStart    string = ">> Data Service Started."
	srvEnd      string = ">> Data Service Shutdown Unexpectedly. Error:"
	reqArrived  string = ">> Request Arrived At"
	reqBody     string = ">> Request Body:"
	statError   string = ">> Status method not allowed"
	dataFailed  string = ">> Data Service Request Failed:"
	dataSuccess string = ">> Data Service Request Done."
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
		failHandle(w, dataFailed+statError, http.StatusInternalServerError)
		return
	}
	// body read for json parse.
	json, err := ioutil.ReadAll(r.Body)
	if err != nil {
		failHandle(w, dataFailed+err.Error(), http.StatusInternalServerError)
		return
	}
	// action value determines the CRUD ection.
	action, err := jin.GetString(json, "action")
	if err != nil {
		failHandle(w, dataFailed, http.StatusInternalServerError)
		return
	}

	switch action {
	case "insert":
		err, status := insertRecord(json)
		if err != nil {
			failHandle(w, dataFailed+err.Error(), status)
			return
		}
	case "update":
		err, status := updateRecord(json)
		if err != nil {
			failHandle(w, dataFailed+err.Error(), status)
			return
		}
	}
	doneHandle(w)
}

func dbConn() {
	var err error
	base, err = sql.Open(envServiceMap["user"], dbEnv)
	if err != nil {
		panic(err)
	}
}

func updateRecord(json []byte) (error, int) {
	jsonMap, err := jin.GetMap(json)
	if err != nil {
		return err, http.StatusInternalServerError
	}
	keys, values, err := jin.GetKeysValues(json, "body")
	if err != nil {
		return err, http.StatusInternalServerError
	}
	// record exists or not
	query := seecool.Select(envServiceMap["table"]).
		Equal(jsonMap["key"], jsonMap["value"])
	result, err := seecool.QueryJson(base, query)
	if err != nil {
		return err, http.StatusInternalServerError
	}
	if string(result) == "[]" {
		return recordNotExists, http.StatusBadRequest
	}

	query = seecool.Update(envServiceMap["table"]).
		Keys(keys...).
		Values(values...).
		Equal(jsonMap["key"], jsonMap["value"])
	_, err = base.Query(query.String())
	if err != nil {
		return err, http.StatusInternalServerError
	}
	return nil, http.StatusOK
}

func insertRecord(json []byte) (error, int) {
	keys, values, err := jin.GetKeysValues(json, "body")
	if err != nil {
		return err, http.StatusInternalServerError
	}
	query := seecool.Insert(envServiceMap["table"]).
		Keys(keys...).
		Values(values...)
	_, err = base.Query(query.String())
	if err != nil {
		return err, http.StatusInternalServerError
	}
	return nil, http.StatusOK
}

func statusFailed(err string) []byte {
	return responseScheme.MakeJson("Failed", seecool.EscapeQuote(err))
}

func statusSuccess() []byte {
	return responseScheme.MakeJson("OK", "null")
}

func failHandle(w http.ResponseWriter, err string, status int) {
	log.Println(dataFailed, err)
	w.WriteHeader(status)
	w.Write(statusFailed(err))
}

func doneHandle(w http.ResponseWriter) {
	log.Println(dataSuccess)
	w.WriteHeader(http.StatusOK)
	w.Write(statusSuccess())
}
