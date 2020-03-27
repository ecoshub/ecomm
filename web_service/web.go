package main

import (
	"log"
	"net/http"
	"penman"

	"github.com/gorilla/mux"
)

func main() {

	r := mux.NewRouter()
	r.HandleFunc("/", rootHandle).Methods("GET")
	r.HandleFunc("/home", rootHandle).Methods("GET")
	r.HandleFunc("/login", loginHandle).Methods("GET", "POST")
	r.HandleFunc("/signup", signupHandle).Methods("GET", "POST")
	r.HandleFunc("/profile", profileHandle).Methods("GET")
	err := http.ListenAndServe(":8080", r)
	log.Fatal(err)
}

func rootHandle(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(`
		<a href="http://localhost:8080/login">login</a><br>
		<a href="http://localhost:8080/profile">profile</a><br>
		<a href="http://localhost:8080/signup">signup</a>`))
}

func loginHandle(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Origin, cache-control")
	w.Header().Set("Content-Type", "text/html")

	w.Write(penman.Read(penman.GetCurrentDir() + penman.Sep() + "login.html"))
}

func signupHandle(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(`<p>signup</p><br><a href="http://localhost:8080">home</a><br>`))
}

func profileHandle(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(`<p>profile</p><br><a href="http://localhost:8080">home</a><br>`))
}
