package main

import (
	"fmt"
	"net/http"
	"io"
	_ "github.com/lib/pq"
	"database/sql"
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
	http.HandleFunc("/alreadyout",alreadyOut)
	http.HandleFunc("/login",login)
	http.HandleFunc("/already",already)
	http.HandleFunc("/",redirec) 
	http.HandleFunc("/logout",logout)
	http.HandleFunc("/secret",secret)
	http.ListenAndServe(":8080", nil)
}

func logout(w http.ResponseWriter, r *http.Request){
	//remember to change as post
	if r.Method == http.MethodGet{
		_, err := r.Cookie("session")
		if err != nil{
			http.Redirect(w, r, "/alreadyout", http.StatusSeeOther)
			return
		}

		c := &http.Cookie{
			Name:   "session",
			Value:  "",
			MaxAge: -1,
		}
		http.SetCookie(w, c)

		_, err = io.WriteString(w,"you have log out successfully")
		if err != nil {
			panic(err)
			return
		}
		return
	}

	_, err := io.WriteString(w,"log out only accept POST request")
		if err != nil {
			panic(err)
			return
		}
}

func alreadyOut(w http.ResponseWriter, r *http.Request){
	_, err := io.WriteString(w,"you have already log out, please do not resubmit the logout request")
	if err != nil {
		panic(err)
		return
	}
} 


func redirec(w http.ResponseWriter, r *http.Request){
	_, err := io.WriteString(w,"you need to login first")
		if err != nil {
			panic(err)
			return
		}
} 


//没啥用
func secret(w http.ResponseWriter, r *http.Request){
	_, err := r.Cookie("session")
	if err != nil{
		io.WriteString(w,"you can not see secret unless you login")
		return
	}
	io.WriteString(w,"you can see secret since you have logged in")
}


func already(w http.ResponseWriter, r *http.Request){
	cookiee, err := r.Cookie("session")
	if err != nil{
		panic(err)
		return
	}

	if cookiee.Value == ""{
		io.WriteString(w,"Please clean your empty cookies by hand")
		return
	}

	// from cookie check which client have logged in
	row := db.QueryRow("SELECT name from client where jwt = $1;",cookiee.Value)
	var good string
	_ = row.Scan(&good)
	fmt.Println(good)

	_, err = io.WriteString(w,"you already loggedi in as " + good +", \nif you want to login as other user, please logout first")
	if err != nil {
		panic(err)
		return
	}
} 



func login(w http.ResponseWriter, r *http.Request){
	_, err := r.Cookie("session")
	if err == nil{
		http.Redirect(w, r, "/already", http.StatusSeeOther)
		return
	}
	//remember make it as post
	if r.Method == http.MethodGet{
		jwt  := r.FormValue("jwt")
		if jwt == ""{
			http.Error(w, "you have enter wrong username or password", 403)
			return
		}
		var good string
		row := db.QueryRow("SELECT name from client where jwt = $1;",jwt)
		err := row.Scan(&good)
		if err != nil{
			http.Error(w, "you have enter wrong username or password", 403)
			return
		}
		fmt.Println(good)


		c := &http.Cookie{
			Name:  "session",
			Value: jwt,
		}

		http.SetCookie(w, c)

		io.WriteString(w,"you have login as " + good)
	}
}





