package main

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"net/http"
)

var db *sql.DB

func connection() {
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

type houses struct {
	region  string
	owner   string
	address string
	price   int
}

func main() {
	connection()
	http.HandleFunc("/house", houseList)
	http.HandleFunc("/house/", Queryhouse)
	http.ListenAndServe(":8888", nil)

}

func houseList(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, http.StatusText(405), http.StatusMethodNotAllowed)
		return
	}

	rows, err := db.Query("SELECT region,address,owner,price FROM house;")
	if err != nil {
		http.Error(w, http.StatusText(500), 500)
		return
	}
	defer rows.Close()

	hous := make([]houses, 0)
	for rows.Next() {
		hou := houses{}
		err := rows.Scan(&hou.region, &hou.address, &hou.owner, &hou.price) // order matters
		if err != nil {
			http.Error(w, http.StatusText(500), 500)
			return
		}
		hous = append(hous, hou)
	}
	if err = rows.Err(); err != nil {
		http.Error(w, http.StatusText(500), 500)
		return
	}

	for _, hou := range hous {
		fmt.Fprintf(w, "Address: %s,\n Region: %s,   Owner: %s,    Price: %v$\n\n", hou.address, hou.region, hou.owner, hou.price)
	}
}

func Queryhouse(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, http.StatusText(405), http.StatusMethodNotAllowed)
		return
	}

	area := r.FormValue("area")
	if area == "" {
		http.Error(w, http.StatusText(400), http.StatusBadRequest)
		return
	}

	rows, err := db.Query("SELECT region,address,owner,price FROM house where region=$1;", area)
	if err != nil {
		http.Error(w, http.StatusText(500), 500)
		return
	}
	defer rows.Close()


	hous := make([]houses, 0)
	for rows.Next() {
		hou := houses{}
		err := rows.Scan(&hou.region, &hou.address, &hou.owner, &hou.price) // order matters
		if err != nil {
			http.Error(w, http.StatusText(500), 500)
			return
		}
		hous = append(hous, hou)
	}
	if err = rows.Err(); err != nil {
		http.Error(w, http.StatusText(500), 500)
		return
	}

	for _, hou := range hous {
		fmt.Fprintf(w, "Address: %s,\n Region: %s,   Owner: %s,    Price: %v$\n\n", hou.address, hou.region, hou.owner, hou.price)
	}
}
