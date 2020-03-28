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
	recordNotExists   *errorx.Error = errorx.New("Not Exists Error", "Record does not exists", 0)
	keyValuePairerror *errorx.Error = errorx.New("Wrong Number", "One key&value pair expected", 1)
	wrongAction       *errorx.Error = errorx.New("Wrong Action", "Wrong or missing 'action' key", 2)
	statError         *errorx.Error = errorx.New("Not Allowed", "Status method not allowed", 3)
	dataFailed        *errorx.Error = errorx.New("Request Failed", "Data Service Request Failed", 3)
	dataSuccess       *errorx.Error = errorx.New("Request Done", "Data Service Request Done", 3)

	// log strings
	srvStart   string = ">> Data Service Started"
	srvEnd     string = ">> Data Service Shutdown Unexpectedly"
	reqArrived string = ">> Request Arrived At"
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
		failHandle(w, statError, http.StatusInternalServerError)
		return
	}
	// body read for json parse.
	json, err := ioutil.ReadAll(r.Body)
	if err != nil {
		failHandle(w, err, http.StatusInternalServerError)
		return
	}
	// action value determines the CRUD ection.
	action, err := jin.GetString(json, "action")
	if err != nil {
		lene := len(err.Error())
		errCode := err.Error()[lene-3 : lene-1]
		if errCode == "08" {
			failHandle(w, dataFailed, http.StatusBadRequest)
			return
		}
		failHandle(w, dataFailed, http.StatusInternalServerError)
		return
	}

	switch action {
	case "insert":
		err, status := insertRecord(json)
		if err != nil {
			failHandle(w, err, status)
			return
		}
	case "update":
		err, status := updateRecord(json)
		if err != nil {
			failHandle(w, err, status)
			return
		}
	case "delete":
		err, status := deleteRecord(json)
		if err != nil {
			failHandle(w, err, status)
			return
		}
	default:
		failHandle(w, wrongAction, http.StatusBadRequest)
		return
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

func deleteRecord(json []byte) (error, int) {
	keys, values, err := jin.GetKeysValues(json, "body")
	if err != nil {
		return err, http.StatusInternalServerError
	}
	if len(keys) != 1 || len(values) != 1 {
		return keyValuePairerror, http.StatusBadRequest
	}
	// primary or unique key & value pair
	key := keys[0]
	value := values[0]
	// record exists or not
	query := seecool.Select(envServiceMap["table"]).Equal(key, value)
	result, err := seecool.QueryJson(base, query)
	if err != nil {
		return err, http.StatusInternalServerError
	}
	if string(result) == "[]" {
		return recordNotExists, http.StatusBadRequest
	}
	query = seecool.Delete(envServiceMap["table"]).Equal(key, value)
	_, err = base.Query(query.String())
	if err != nil {
		return err, http.StatusInternalServerError
	}
	return nil, http.StatusOK
}

func updateRecord(json []byte) (error, int) {
	jsonMap, err := jin.GetMap(json)
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
	keys, values, err := jin.GetKeysValues(json, "body")
	if err != nil {
		return err, http.StatusInternalServerError
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

func statusFailed(err error) []byte {
	return responseScheme.MakeJson("Failed", seecool.EscapeQuote(err.Error()))
}

func statusSuccess() []byte {
	return responseScheme.MakeJson("OK", "null")
}

func failHandle(w http.ResponseWriter, err error, status int) {
	log.Println(dataFailed, err)
	w.WriteHeader(status)
	w.Write(statusFailed(err))
}

func doneHandle(w http.ResponseWriter) {
	log.Println(dataSuccess)
	w.WriteHeader(http.StatusOK)
	w.Write(statusSuccess())
}
