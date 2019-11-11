package main

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"net/http"
	"io"

)

var db *sql.DB

func connectdb(){
	var err error
	db, err = sql.Open("postgres", "postgres://admin:admin@localhost/rookie?sslmode=disable")
	if err != nil {
		panic(err)
	}
	if err = db.Ping(); err != nil {
		panic(err)
	}
	fmt.Println("You connected to your database.")
}

func main() {
	connectdb()
	http.HandleFunc("/regist", regist)
	err := http.ListenAndServe(":8080", nil)
	if err != nil{
		panic(err)
	}
}


func regist(w http.ResponseWriter, r *http.Request){
	if r.Method == http.MethodPost{
		name := r.FormValue("name")
		pw   := r.FormValue("pw")
		jwt  := r.FormValue("jwt")
		email:= r.FormValue("email")
		tye  := r.FormValue("type")

		if name == "" ||  email == "" || jwt == ""|| tye == ""|| pw==""{
			http.Error(w, "Please enter name, password, email and choose type", http.StatusForbidden)
			return
		}

		// check if name already used
		row := db.QueryRow("SELECT type from client where name = $1;",name)
		var good string
		er := row.Scan(&good)
		if er != nil{
			// http.Error(w, http.StatusText(500), http.StatusInternalServerError)
			fmt.Println("not found")
		}

		// check if name already used
		if good == "host" || good == "tenant" {
			http.Error(w, "Username already used, please try another", http.StatusForbidden)
			return
		}

		//insert into DB
		_, err := db.Exec("INSERT INTO client (name, email, code, jwt, type) VALUES ($1, $2, $3, $4, $5)", name, email, pw, jwt, tye)
		if err != nil {
			http.Error(w, http.StatusText(500), http.StatusInternalServerError)
			panic(err)
			return
		}
		_, err = io.WriteString(w,"you have created new user")
		if err != nil {
			panic(err)
			return
		}

	}


}













