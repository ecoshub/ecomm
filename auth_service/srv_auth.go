package main

import (
	"database/sql"
	"errorx"
	"io/ioutil"
	"jin"
	"log"
	"net/http"
	"penman"
	"seecool"

	_ "github.com/lib/pq"
)

const (
	// environment directories,
	// 'curr' keyword is a wild card for 'currentDirectory'
	// valid wildcard can be user with 'github.com/ecoshub/penman' package
	envDatabaseDir string = "curr/.env_database"
	envAuthDir     string = "curr/.env_service"
	envMainDir     string = "curr/../.env_main"

	// log strings
	srvStart    string = ">> Authentication Service Started."
	srvEnd      string = ">> Authentication Service Shutdown Unexpectedly. Error:"
	reqArrived  string = ">> Request Arrived At"
	reqBody     string = ">> Request Body:"
	authGranted string = "Authentication Request Granted"
)

var (
	// service environment map
	envServiceMap map[string]string

	// main environment environment map
	envMainMap map[string]string

	// database environment file
	dbEnv string

	// authentication service main port
	mainPort string

	// return columns
	retColumns []string = []string{"user_id", "type", "email", "password"}

	// json format schemes
	responseScheme *jin.Scheme

	// errors
	missingEnvFile *errorx.Error = errorx.New("Fatal Error", "Missing environment file or wrong file directory", 0)
	portNotExist   *errorx.Error = errorx.New("Fatal Error", "Main service port does not exist in the main environment file.", 1)
	retrunNotExist *errorx.Error = errorx.New("Fatal Error", "return array does not exist in the main environment file.", 2)
	authFail       *errorx.Error = errorx.New("Database", "Wrong password", 3)
	recordNotExist *errorx.Error = errorx.New("Database", "Record Not Exists", 4)
	moreExist      *errorx.Error = errorx.New("Database", "More then one record exists with your primary key value", 5)
	statError      *errorx.Error = errorx.New("Service", "Status method not allowed", 6)
	authFailed     *errorx.Error = errorx.New("Service", "Authentication Request Failed", 7)
)

func init() {
	var err error
	// read main env. file
	envMainMap, err = seecool.GetEnv(envMainDir)
	if err != nil {
		panic(err)
	}
	mainPort = envMainMap["auth_service_port"]
	if mainPort == "" {
		panic(err)
	}
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
	// response scheme
	responseScheme = jin.MakeScheme("status", "response", "error")
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
		failHandle(w, err, http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()
	// request body log.
	log.Println(reqBody, string(json))

	// record check core function.
	key, status, err := checkRecord(envServiceMap["userTable"], json)
	if err != nil {
		failHandle(w, err, status)
		return
	}
	doneHandle(w, key)
}

func checkRecord(table string, json []byte) ([]byte, int, error) {
	db, err := sql.Open(envServiceMap["user"], dbEnv)
	if err != nil {
		return nil, -1, err
	}
	defer db.Close()
	// get control keys
	primaryKey := envServiceMap["primKey"]
	passKey := envServiceMap["passKey"]

	// get received primary key from request
	// get received primary password key from request
	jsonMap, err := jin.GetAllMap(json, []string{primaryKey, passKey})
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}
	primKeyReceive := jsonMap[primaryKey]
	passKeyReceive := jsonMap[passKey]

	// create a query from primary key search
	query := seecool.Select(table, retColumns...).Equal(primaryKey, primKeyReceive)
	result, err := seecool.QueryJson(db, query)
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}
	lenr, err := jin.Length(result)
	if err != nil {
		return nil, http.StatusBadRequest, recordNotExist
	}
	if lenr > 1 {
		return nil, http.StatusInternalServerError, moreExist
	}
	if lenr == 0 {
		return nil, http.StatusInternalServerError, recordNotExist
	}
	// get correct password from database response
	correctPass, err := jin.GetString(result, "0", passKey)
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}
	// password check
	if passKeyReceive != correctPass {
		return nil, http.StatusOK, authFail
	}
	//  get response body for return
	response, err := jin.Get(result, "0")
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}
	return response, http.StatusOK, nil
}

func statusFailed(err error) []byte {
	return responseScheme.MakeJson("Failed", "null", seecool.EscapeQuote(err.Error()))
}

func statusGranted(response []byte) []byte {
	return responseScheme.MakeJson("OK", string(response), "null")
}

func failHandle(w http.ResponseWriter, err error, status int) {
	log.Println(authFailed.Link(err))
	w.WriteHeader(status)
	w.Write(statusFailed(authFailed))
	authFailed.ClearLink()
}

func doneHandle(w http.ResponseWriter, response []byte) {
	log.Println(authGranted)
	w.WriteHeader(http.StatusOK)
	w.Write(statusGranted(response))
}
