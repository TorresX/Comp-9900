package main

import (
	"database/sql"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	_ "github.com/lib/pq"
)

var db *sql.DB

func connectdb() {
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
	http.HandleFunc("/getprofile", getProfile)
	http.HandleFunc("/updateprofile", updateProfile)
	http.HandleFunc("/updatepw", updatePw)
	http.HandleFunc("/updateimage", updateImage)
	http.ListenAndServe(":8080", nil)
}

type Profile struct {
	name  string
	email string
	phone string
	typo  string
	image string
}

func getProfile(w http.ResponseWriter, r *http.Request) {
	cookiee, err := r.Cookie("session")
	if err != nil {
		http.Error(w, "You have to login first", http.StatusForbidden)
		return
	}
	row := db.QueryRow("SELECT name,email,type,image,phone from client where jwt = $1;", cookiee.Value)
	pro := Profile{}
	err = row.Scan(&pro.name, &pro.email, &pro.typo, &pro.image, &pro.phone)
	if err != nil {
		fmt.Println("not found phone")
	}
	s := "Name:  " + pro.name + "\nEmail: " + pro.email + "\nType:  " + pro.typo + "\nimage: " + pro.image + "\nPhone: " + pro.phone
	io.WriteString(w, s)

}

//set to post
func updateProfile(w http.ResponseWriter, r *http.Request) {
	cookiee, err := r.Cookie("session")
	if err != nil {
		http.Error(w, "You have to login first", http.StatusForbidden)
		return
	}
	name := r.FormValue("name")
	email := r.FormValue("email")
	phone := r.FormValue("phone")

	if name == "" || email == "" || phone == "" {
		http.Error(w, "Please enter name, email and phone", http.StatusForbidden)
		return
	}

	_, err = db.Exec("UPDATE client SET name = $1, email = $2, phone=$3 WHERE jwt = $4;", name, email, phone, cookiee.Value)
	if err != nil {
		http.Error(w, http.StatusText(500), http.StatusInternalServerError)
		log.Println(err)
		return
	}
	io.WriteString(w, "you have update your profile")
	fmt.Println(name, " updated his profile")
}

// set to post
func updatePw(w http.ResponseWriter, r *http.Request) {
	cookiee, err := r.Cookie("session")
	if err != nil {
		http.Error(w, "You have to login first", http.StatusForbidden)
		return
	}
	oldpw := r.FormValue("oldpw")
	newpw := r.FormValue("newpw")
	newjwt := r.FormValue("newjwt")
	var checkpw string
	row := db.QueryRow("SELECT code from client where jwt = $1;", cookiee.Value)
	err = row.Scan(&checkpw)
	if err != nil {
		fmt.Println("not found old password")
		return
	}
	// fmt.Println(checkpw)
	if checkpw != oldpw {
		http.Error(w, "You entered wrong old password, please try again", http.StatusForbidden)
		return
	}

	//UUUUUUPdate
	_, err = db.Exec("UPDATE client SET code = $1, jwt = $2 WHERE jwt = $3;", newpw, newjwt, cookiee.Value)
	if err != nil {
		http.Error(w, http.StatusText(500), http.StatusInternalServerError)
		// panic(err)
		return
	}

	// update cookie with new jwt
	c := &http.Cookie{
		Name:  "session",
		Value: newjwt,
	}
	http.SetCookie(w, c)

	fmt.Println(newjwt)
	var name string
	// Check from Updated file
	row2 := db.QueryRow("SELECT name from client where jwt = $1;", newjwt)
	err = row2.Scan(&name)
	if err != nil {
		http.Error(w, "Did not update sucessfully", http.StatusForbidden)
		return
	}

	io.WriteString(w, "you have update your profile")
	fmt.Println(name, " updated his profile")
}

func updateImage(w http.ResponseWriter, r *http.Request) {
	// change into post
	var s string
	var imgnum string
	cookiee, err := r.Cookie("session")
	if err != nil {
		http.Error(w, "You have to login first", http.StatusForbidden)
		return
	}
	if r.Method == http.MethodPost {
		noimage := "null.jpg"
		jwt := cookiee.Value
		row := db.QueryRow("SELECT image from client where jwt = $1;", jwt)
		err = row.Scan(&imgnum)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			log.Println("Line 159", err)
			return
		}
		if imgnum == noimage {
			log.Println("This is the first image of this user")
			var temp int
			row1 := db.QueryRow("select value from number where type = 'owner';")
			err = row1.Scan(&temp)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				log.Println("Line 168", err)
				return
			}
			// log.Println(temp)
			temp++
			tempp := strconv.Itoa(temp)
			tem := ".jpg"
			imgnum = tempp + tem
			// log.Println(tempp)
			_, err = db.Exec("UPDATE number SET value = $1 WHERE type = $2;", temp, "owner")
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				log.Println("Line 179", err)
				return
			}
			_, _ = db.Exec("UPDATE client SET image = $1 WHERE jwt =  $2;", imgnum, jwt)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				log.Println("Line 184", err)
				return
			}

		}
		log.Println(imgnum)

		// open
		f, _, err := r.FormFile("p")
		if err != nil {
			// http.Error(w, err.Error(), http.StatusInternalServerError)
			http.Error(w, "no file founded", http.StatusForbidden)
			return
		}
		defer f.Close()

		// for pic information
		// fmt.Println("\nfile:", f, "\nheader:", h, "\nerr", err)

		// read
		bs, err := ioutil.ReadAll(f)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// store on server
		dst, err := os.Create(filepath.Join("./head/", imgnum))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer dst.Close()

		_, err = dst.Write(bs)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		s = "you have update your image"

	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	io.WriteString(w, `
	<form method="POST" enctype="multipart/form-data">
	<input type="file" name="p">
	<input type="submit">
	</form>
	<br>`+s)

}
