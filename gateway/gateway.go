package main

import (
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

var (
	envMainDir        string = "curr/../.env_main"
	secretDir         string = "../.secret"
	secret            string
	sessStore         *sessions.CookieStore
	auth_service_port string

	portNotExist *errorx.Error = errorx.New("Fatal Error", "Main service port does not exist in the main environment file.", 6)
	// main environment environment map
	envMainMap map[string]string
	// gateway service main port
	mainPort string
)

func init() {
	var err error
	// read main env. file
	envMainMap, err = seecool.GetEnv(envMainDir)
	if err != nil {
		panic(err)
	}
	mainPort = envMainMap["gate_service_port"]
	if mainPort == "" {
		panic(portNotExist)
	}
	secret = penman.SRead(secretDir)
	if secret == "" {
		panic("No secret found.")
	}
	sessStore = sessions.NewCookieStore([]byte(secret))
	sessStore.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   30 * 24 * 60 * 60, // 1 month
		HttpOnly: true,
	}
}

func main() {

	fmt.Println("Gateway Service Started at:", mainPort)
	http.HandleFunc("/", rootHandle)
	http.ListenAndServe(":"+mainPort, nil)
	// auth()
	// permission()
	// sess_cookie()
	// handle_request()
}

func rootHandle(w http.ResponseWriter, r *http.Request) {
	// session.Values["json"] = "bar"
	// session.Values[42] = 43
	// session.Values = nil
	// err := session.Save(r, w)
	// if err != nil {
	// 	http.Error(w, err.Error(), http.StatusInternalServerError)
	// 	return
	// }
	sess, exists, err := cookieCheck("login", r)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
	}
	if exists {
		fmt.Println("cookie exists")
		fmt.Println(sess.Values["user_id"])
	} else {
		fmt.Println("cookie not exists")
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
		}
		resp, err := authenticationControl(r)
		stat, err := jin.GetString(resp, "status")
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
		}
		if stat == "OK" {
			respMap, err := jin.GetMap(resp, "response")
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(err.Error()))
				return
			}
			fmt.Println(respMap)
			fmt.Println("DONE")
			// sess.Values["user_id"] = respMap["user_id"]
			// sess.Values["type"] = "user"
			// err = sess.Save(r, w)
			// if err != nil {
			// 	http.Error(w, err.Error(), http.StatusInternalServerError)
			// 	return
			// }
		} else {

		}
		// breakx.Printif(string(resp), err)
		// session, err := sessStore.Get(r, "login")
	}
	return
}

func cookieCheck(name string, r *http.Request) (*sessions.Session, bool, error) {
	session, err := sessStore.Get(r, "login")
	fmt.Printf("%T\n", session)
	if err != nil {
		return nil, false, err
	}
	if len(session.Values) == 0 {
		return nil, false, nil
	}
	return session, true, nil
}

func authenticationControl(r *http.Request) ([]byte, error) {
	json, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	fmt.Println("arrived", string(json))
	resp, err := http.Post("http://localhost:"+envMainMap["auth_service_port"], "application/json", bytes.NewBuffer(json))
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
