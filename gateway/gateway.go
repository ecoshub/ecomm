package main

import (
	"breakx"
	"bytes"
	"errorx"
	"fmt"
	"io/ioutil"
	"jin"
	"log"
	"net/http"
	"penman"
	"seecool"

	"github.com/gorilla/sessions"
)

const (
	// environment directories,
	// 'curr' keyword is a wild card for 'currentDirectory'
	// valid wildcard can be user with 'ecoshub/penman' and 'ecoshub/seecool' GetEnv() func.
	envMainDir string = "curr/../.env_main"
	secretDir  string = "../.secret"

	// log strings
	srvStart   string = ">> Gateway Service Started."
	srvEnd     string = ">> Gateway Service Shutdown Unexpectedly. Error:"
	reqArrived string = ">> Request Arrived At"
	reqBody    string = ">> Request Body:"
)

var (

	// main session store
	store *sessions.CookieStore

	// session crypto string
	secret string

	// auth service port
	auth_service_port string

	// my service name from env file
	myServiceName string = "gate_service_port"

	// gateway service main port
	mainPort string

	// main environment environment map
	envMainMap map[string]string

	// errors
	portNotExist   *errorx.Error = errorx.New("Fatal Error", "Main service port does not exist in the main environment file.", 0)
	secretNotExist *errorx.Error = errorx.New("Fatal Error", "secret not exist in the main environment file.", 1)
)

func init() {
	var err error
	// read main env. file
	envMainMap, err = seecool.GetEnv(envMainDir)
	if err != nil {
		panic(err)
	}
	mainPort = envMainMap[myServiceName]
	if mainPort == "" {
		panic(portNotExist)
	}
	secret = penman.SRead(secretDir)
	if secret == "" {
		panic(secretNotExist)
	}
	store = sessions.NewCookieStore([]byte(secret))
	store.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   30 * 24 * 60 * 60, // 1 month
		HttpOnly: true,
	}
}

func main() {
	log.Println(srvStart, "port:", mainPort)
	http.HandleFunc("/login", loginHandle)
	http.HandleFunc("/logout", logoutHandle)
	err := http.ListenAndServe(":"+mainPort, nil)
	log.Println(srvEnd, err)
}

func loginHandle(w http.ResponseWriter, r *http.Request) {
	// band json error is wrong?
	loginSession, auth, err := cookieHandle(w, r)
	if err != nil {
		failHandle(w, err, http.StatusInternalServerError)
		return
	}
	fmt.Println(auth)
	fmt.Println(loginSession.Values)
}

func logoutHandle(w http.ResponseWriter, r *http.Request) {
	// later
}

func cookieHandle(w http.ResponseWriter, r *http.Request) (*sessions.Session, bool, error) {
	// get cookie named 'login'
	loginSession, err := store.Get(r, "login")
	if err != nil {
		return nil, false, err
	}
	// request read
	json, err := ioutil.ReadAll(r.Body)
	if err != nil {
		breakx.Point()
		return loginSession, false, err
	}
	action, err := jin.GetString(json, "action")
	if err != nil {
		lene := len(err.Error())
		errCode := err.Error()[lene-3 : lene-1]
		if errCode != "08" {
			breakx.Point()
			return loginSession, false, err
		}
	}
	// shortcut auth
	if loginSession.Values["auth"] == "true" && action != "login" {
		// keep going
		return loginSession, true, nil
	}

	// mthod check for login action.
	if string(r.Method) == "POST" {
		// wants to login?
		if action == "login" {
			resp, auth, err := authenticationControl(json)
			if err != nil {
				breakx.Point()
				return loginSession, false, err
			}
			if auth {
				respMap, err := jin.GetMap(resp, "response")
				if err != nil {
					breakx.Point()
					return loginSession, false, err
				}
				for k, v := range respMap {
					loginSession.Values[k] = v
				}
				loginSession.Values["auth"] = "true"
				err = loginSession.Save(r, w)
				if err != nil {
					breakx.Point()
					return loginSession, false, err
				}
				breakx.Point()
				return loginSession, true, nil
			}
		}
	}
	loginSession.Values["auth"] = "false"
	err = loginSession.Save(r, w)
	if err != nil {
		return loginSession, false, err
	}
	return loginSession, false, nil
}

func authenticationControl(json []byte) ([]byte, bool, error) {
	resp, err := http.Post("http://localhost:"+envMainMap["auth_service_port"], "application/json", bytes.NewBuffer(json))
	if err != nil {
		return nil, false, err
	}
	defer resp.Body.Close()
	json, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, false, err
	}
	return json, true, nil
}

func failHandle(w http.ResponseWriter, err error, status int) {
	log.Println(err)
	w.WriteHeader(status)
	w.Write([]byte(err.Error()))
}
