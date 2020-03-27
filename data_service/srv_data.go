package main

import (
	"database/sql"
	"fmt"
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
	var err error
	query := seecool.Select("users", "username", "type")
	resp, err := seecool.QueryJson(base, query)
	fmt.Println(err)
	fmt.Println(string(resp))
	defer base.Close()
}

func dbConn() {
	var err error
	base, err = sql.Open(envServiceMap["user"], dbEnv)
	if err != nil {
		panic(err)
	}
}
