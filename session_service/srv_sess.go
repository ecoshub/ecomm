package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"jin"
	"net/http"
	"seecool"

	"github.com/gorilla/sessions"
)

var (
	mainPort string
	// environment directories,
	// 'curr' keyword is a wild card for 'currentDirectory'
	// valid wild card with 'github.com/ecoshub/penman' package
	envSessionDir string = "curr/.env_session"
	envMainDir    string = "curr/../.env_main"

	// environment map
	envSessionMap map[string]string
	// main environment environment map
	envMainMap map[string]string
	sessStore  *sessions.CookieStore
)

func init() {
	var err error
	envMainMap, err = seecool.GetEnv(envSessionDir)
	if err != nil {
		panic(err)
	}
	mainPort = envMainMap["sess_service_port"]
	envSessionMap, err = seecool.GetEnv(envSessionDir)
	if err != nil {
		panic(err)
	}
	sessStore = sessions.NewCookieStore([]byte(envSessionMap["secret"]))
	sessStore.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   86400 * 7,
		HttpOnly: true,
	}
}

func main() {
	fmt.Println("Session service started. port:", mainPort)
	http.HandleFunc("/", MyHandler)
	err := http.ListenAndServe(":"+mainPort, nil)
	panic(err)
}

func MyHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Origin, cache-control")
	w.Header().Set("Content-Type", "application/json")

	jsonBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	authResp, err := authenticationControl(jsonBody)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	authMap, err := jin.GetMap(authResp)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	fmt.Println(authMap)

	// session, _ := sessStore.Get(r, "session-name")
	// // session.Values["json"] = "bar"
	// // session.Values[42] = 43
	// session.Values = nil
	// err := session.Save(r, w)
	// if err != nil {
	// 	http.Error(w, err.Error(), http.StatusInternalServerError)
	// 	return
	// }
	// fmt.Println(session.Values)
}

func authenticationControl(json []byte) ([]byte, error) {
	resp, err := http.Post(envSessionMap["authURL"], "application/json", bytes.NewBuffer(json))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	json, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return json, nil
}

// http.Error(w, err.Error(), http.StatusInternalServerError)
