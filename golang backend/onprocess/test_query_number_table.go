package main

import (
	"database/sql"
	"fmt"
	"log"
	"strconv"

	_ "github.com/lib/pq"
)

var db *sql.DB

func init() {
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

func notmain() {
	var imgnum string
	noimage := "null.jpg"
	jwt := "bb"
	row := db.QueryRow("SELECT image from client where jwt = $1;", jwt)
	err := row.Scan(&imgnum)
	if err != nil {
		log.Println("Did not find your profile")
		return
	}
	if imgnum == noimage {
		log.Println("yes")
		var temp int
		row1 := db.QueryRow("select value from number where type = 'owner';")
		_ = row1.Scan(&temp)
		// log.Println(temp)
		temp++
		tempp := strconv.Itoa(temp)
		tem := ".jpg"
		imgnum = tempp + tem
		log.Println(tempp)
		_, _ = db.Exec("UPDATE number SET value = $1 WHERE type = $2;", temp, "owner")
		_, _ = db.Exec("UPDATE client SET image = $1 WHERE jwt =  $2;", imgnum, jwt)

	}
	log.Println(imgnum)
}
