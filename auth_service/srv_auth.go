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
	missingEnvFile *errorx.Error = errorx.New("Fatal Error", "Missing environment file or wrong file directory", 0)
	portNotExist   *errorx.Error = errorx.New("Fatal Error", "Main service port does not exist in the main environment file.", 1)
	authFail       *errorx.Error = errorx.New("Database", "Wrong password", 2)
	recordNotExist *errorx.Error = errorx.New("Database", "Record Not Exists", 3)
	moreExist      *errorx.Error = errorx.New("Database", "More then one record exists with your primary key value", 4)
	statError      *errorx.Error = errorx.New("Service", "Status method not allowed", 5)
	authFailed     *errorx.Error = errorx.New("Service", "Authentication Request Failed", 6)
	authGranted    *errorx.Error = errorx.New("Service", "Authentication Request Granted", 7)

	// log strings
	srvStart   string = ">> Authentication Service Started."
	srvEnd     string = ">> Authentication Service Shutdown Unexpectedly. Error:"
	reqArrived string = ">> Request Arrived At"
	reqBody    string = ">> Request Body:"
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
	db, err := dbConn()
	if err != nil {
		panic(err)
	}
	defer db.Close()
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

func dbConn() (*sql.DB, error) {
	db, err := sql.Open(envServiceMap["user"], dbEnv)
	if err != nil {
		return nil, err
	}
	return db, nil
}

func checkRecord(table string, json []byte) ([]byte, int, error) {
	db, err := dbConn()
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}
	defer db.Close()
	// get control keys
	primaryKey := envServiceMap["primKey"]
	passKey := envServiceMap["passKey"]

	// get received primary key from request
	primKeyReceive, err := jin.GetString(json, primaryKey)
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}
	// get received primary password key from request
	passKeyReceive, err := jin.GetString(json, passKey)
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}
	// create a query from primary key search
	query := seecool.Select(table).Equal(primaryKey, primKeyReceive)
	result, err := seecool.QueryJson(db, query)
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}
	// empty response control
	if string(result) == "[]" {
		return nil, http.StatusBadRequest, recordNotExist
	}
	// parse response json
	prs, err := jin.Parse(result)
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}
	// get as array for length check
	arr, err := prs.GetStringArray()
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}
	// length check for response array
	if len(arr) != 1 {
		return nil, http.StatusInternalServerError, moreExist
	}
	// get correct password from database response
	correctPass, err := prs.GetString("0", passKey)
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}
	// password check
	if passKeyReceive != correctPass {
		return nil, http.StatusOK, authFail
	}
	//  get response body for return
	response, err := prs.Get("0")
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
