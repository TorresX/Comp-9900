package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
)

func main() {
	http.HandleFunc("/", foo)
	http.Handle("/favicon.ico", http.NotFoundHandler())
	http.ListenAndServe(":8080", nil)
}

func foo(w http.ResponseWriter, r *http.Request) {

	var s string
	var N int

	if r.Method == http.MethodPost {
		pics := [8]string{"p1", "p2", "p3", "p4", "p5", "p6", "p7", "p8"}
		// open
		for i := 0; i < 8; i++ {
			f, h, err := r.FormFile(pics[i])
			if err != nil {
				continue
			}
			defer f.Close()

			// for your information
			// fmt.Println("\nfile:", f, "\nheader:", h, "\nerr", err)
			fmt.Println("\nfile:",  f, "\nerr", err)

			// read
			bs, err := ioutil.ReadAll(f)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			s = string(bs)
			log.Println(s)

			// store on server
			dst, err := os.Create(filepath.Join("./user/", h.Filename))
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			defer dst.Close()
			N++
			fmt.Println("\nThis is the " + strconv.Itoa(N) + " file")

			_, err = dst.Write(bs)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}
	}
	n := strconv.Itoa(N)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	io.WriteString(w, `
<form action="http://localhost:8080/" method="post" enctype="multipart/form-data">
  <p><input type="file" name="p1">
  	<br>
  <p><input type="file" name="p2">
  	<br>
  <p><input type="file" name="p3">
  	<br>
  <p><input type="file" name="p4">
  	<br>
  <p><input type="file" name="p5">
  	<br>
  <p><input type="file" name="p6">
  	<br>
  <p><input type="file" name="p7">
  	<br>
  <p><input type="file" name="p8">
  	<br>
  <p><button type="submit">Submit</button>
</form>
<br>`+" you have uoload " + n + " files")
}
