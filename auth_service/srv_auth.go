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
	envAuthDir     string = "curr/.env_service"
	envMainDir     string = "curr/../.env_main"

	// service environment map
	envServiceMap map[string]string

	// main environment environment map
	envMainMap map[string]string

	// database environment file
	dbEnv string

	// authentication service main port
	mainPort string

	// json format schemes
	responseScheme *jin.Scheme

	// errors
	authFail       error = errors.New("Wrong password.")
	recordNotExist error = errors.New("Record Not Exists.")
	missingEnvFile error = errors.New("Missing environment file or wrong file directory.")

	// log strings
	srvStart    string = ">> Authentication Service Started."
	srvEnd      string = ">> Authentication Service Shutdown Unexpectedly. Error:"
	reqArrived  string = ">> Request Arrived At"
	reqBody     string = ">> Request Body:"
	statError   string = ">> Status method not allowed"
	authFailed  string = ">> Authentication Request Failed:"
	authGranted string = ">> Authentication Request Granted."
)

func init() {
	var err error
	// read main env. file
	envMainMap, err = seecool.GetEnv(envMainDir)
	if err != nil {
		panic(err)
	}
	mainPort = envMainMap["auth_service_port"]
	// read env_service file
	envServiceMap, err = seecool.GetEnv(envAuthDir)
	if err != nil {
		panic(err)
	}
	// read env_database file
	dbEnv = penman.SRead(envDatabaseDir)
	if dbEnv == "" {
		panic(dbEnv)
	}
	db, err := dbConn()
	if err != nil {
		panic(err)
	}
	defer db.Close()
	// response scheme
	responseScheme = jin.MakeScheme("status", envServiceMap["returnKey"], "error")
}

func main() {
	log.Println(srvStart, "port:", mainPort)
	http.HandleFunc("/", authHandle)
	err := http.ListenAndServe(":"+mainPort, nil)
	// handle later
	log.Println(srvEnd, err)
}

func authHandle(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Origin, cache-control")
	w.Header().Set("Content-Type", "application/json")
	// request log
	log.Println(reqArrived, r.RemoteAddr)

	// method check
	if string(r.Method) != http.MethodPost {
		failHandle(w, statError, http.StatusMethodNotAllowed)
		return
	}
	// body read for json parse.
	json, err := ioutil.ReadAll(r.Body)
	if err != nil {
		failHandle(w, authFailed+err.Error(), http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()
	// request body log.
	log.Println(reqBody, string(json))

	// record check core function.
	key, err := checkRecord(envServiceMap["userTable"], json)
	if err != nil {
		failHandle(w, err.Error(), http.StatusUnauthorized)
		return
	}
	doneHandle(w, key)
}

func dbConn() (*sql.DB, error) {
	db, err := sql.Open(envServiceMap["user"], dbEnv)
	if err != nil {
		return nil, err
	}
	return db, nil
}

func checkRecord(table string, json []byte) (string, error) {
	db, err := dbConn()
	if err != nil {
		return "", err
	}
	defer db.Close()

	jsonMap, err := jin.GetMap(json)
	if err != nil {
		return "", err
	}
	primaryKey := envServiceMap["primKey"]
	query := seecool.Select(table).Equal(primaryKey, jsonMap[primaryKey])
	result, err := seecool.QueryJson(db, query)
	if err != nil {
		return "", err
	}

	if string(result) == "[]" {
		return "", recordNotExist
	}

	resultMap, err := jin.GetMap(result, "0")
	if err != nil {
		return "", err
	}
	passwordKey := envServiceMap["passKey"]
	if resultMap[passwordKey] != jsonMap[passwordKey] {
		return "", authFail
	}

	return resultMap[envServiceMap["returnKey"]], nil
}

func statusFailed(err string) []byte {
	return responseScheme.MakeJson("Failed", "null", seecool.EscapeQuote(err))
}

func statusGranted(id string) []byte {
	return responseScheme.MakeJson("OK", id, "null")
}

func failHandle(w http.ResponseWriter, err string, status int) {
	log.Println(authFailed, err)
	w.WriteHeader(status)
	w.Write(statusFailed(err))
}

func doneHandle(w http.ResponseWriter, key string) {
	log.Println(authGranted)
	w.WriteHeader(http.StatusOK)
	w.Write(statusGranted(key))
}
