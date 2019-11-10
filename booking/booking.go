package main

import (
	"database/sql"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
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
	http.HandleFunc("/makeorder", makeorder)
	http.HandleFunc("/checkorder", checkorder)
	http.HandleFunc("/operateorder", operateorder)
	http.HandleFunc("/cancel", cancel)
	http.ListenAndServe(":8080", nil)

}

func makeorder(w http.ResponseWriter, r *http.Request){
	cookiee, err := r.Cookie("session")
	if err != nil {
		http.Error(w, "You have to login first", http.StatusForbidden)
		return
	}

	if r.Method == http.MethodGet{
		http.Error(w, "You have to send booking message by post message", http.StatusForbidden)
		return
	}

	var typo string
	row := db.QueryRow("SELECT type from client where jwt = $1;", cookiee.Value)
	err = row.Scan(&typo)
	if err != nil{
		http.Error(w, "could not find your user type in database", http.StatusForbidden)
		return
	}
	landlord := "host"
	if typo == landlord{
		http.Error(w, "Only tenant can make booking order, please login as tenant", http.StatusForbidden)
		return
	} 

	address   := r.FormValue("address")
	startdate := r.FormValue("startdate")
	enddate   := r.FormValue("enddate")

	// check integrality
	if address == "" || startdate == "" || enddate == "" {
		http.Error(w, "Please enter address, startdate and enddate", http.StatusForbidden)
		return
	}

	var tenant string
	row1 := db.QueryRow("SELECT name from client where jwt = $1;", cookiee.Value)
	err = row1.Scan(&tenant)
	if err != nil{
		http.Error(w, "could not find your user in database, please relog in", http.StatusForbidden)
		return
	}

	var host string
	row2 := db.QueryRow("SELECT owner from house where address = $1;", address)
	err = row2.Scan(&host)
	if err != nil{
		http.Error(w, "could not find your user type in database", http.StatusForbidden)
		return
	}

	var unavaliable string
	row3 := db.QueryRow("SELECT unavaliabledates from house where address = $1;", address)
	err = row3.Scan(&unavaliable)
	if err != nil{
		http.Error(w, "could not find your user type in database", http.StatusForbidden)
		return
	}


	if !judge(startdate, enddate, unavaliable){
		http.Error(w, "You choosen time is not avaliable on this house", http.StatusForbidden)
		return
	}

	// 到这里
	_, err = db.Exec("insert into orders (tenant, owner, address, startdate, enddate) values ($1, $2, $3, $4, $5);", tenant, host, address, startdate, enddate)
	if err != nil {
		http.Error(w, "fail to write into database, line 106", http.StatusInternalServerError)
		log.Println(err)
		return
	}

	newdate := datesupdate(startdate, enddate, unavaliable)
	fmt.Println(newdate)
	_, err = db.Exec("update house set unavaliabledates=$1 where address=$2;", newdate,address)
	if err != nil {
		http.Error(w, "fail to write into database, line 112", http.StatusInternalServerError)
		log.Println(err)
		return
	}


	_,_ = io.WriteString(w, "you have sent the booking request, please wait for host comfirming.")

}

func datesupdate(start string, end string, storage string) (string){
	arr1 := strings.Split(start,"/")
	arr2 := strings.Split(end,"/")
	startday,_ := strconv.Atoi(arr1[2])
	endday,_   := strconv.Atoi(arr2[2])
	gap := endday - startday
	update := storage

	
	var dates []string
	for i := 0; i < gap ; i++{
		dates = append(dates, strconv.Itoa(startday))
		startday++
	}
	
	var date []string
	for i := 0 ; i<gap; i++{
		a := arr1[0] + "/" + arr2[1] + "/" + dates[i]
		date = append(date, a)
	}
	for i := 0; i<gap; i++{
		update = update + date[i] + ","
	}
	return update
}



func judge(start string, end string, storage string) (bool){

	unavaliable := strings.Split(storage,",")
	arr1 := strings.Split(start,"/")
	arr2 := strings.Split(end,"/")
	startday,_ := strconv.Atoi(arr1[2])
	endday,_   := strconv.Atoi(arr2[2])
	gap := endday - startday
	if gap <= 0{
		return false
	}

	
	var dates []string
	for i := 0; i < gap ; i++{
		dates = append(dates, strconv.Itoa(startday))
		startday++
	}
	
	var date []string
	for i := 0 ; i<gap; i++{
		a := arr1[0] + "/" + arr2[1] + "/" + dates[i]
		date = append(date, a)
	}
	
	for i := 0; i < len(unavaliable); i++{
		for j := 0; j < len(date); j++{
				if unavaliable[i] == date[j]{
					fmt.Println("unable to book")
					return false
				}
			}
		}
	return true

}


type Orde struct{
	number  string
	tenant  string
	owner   string
	address string
	status  string

}

func checkorder(w http.ResponseWriter, r *http.Request){
	cookiee, err := r.Cookie("session")
	if err != nil {
		http.Error(w, "You have to login first", http.StatusForbidden)
		return
	}

	var typo string
	typohost   := "host"
	typotenant := "tenant"
	jwt := cookiee.Value

	row := db.QueryRow("SELECT type from client where jwt = $1;", jwt)
	err = row.Scan(&typo)
	if err != nil{
		http.Error(w, "could not find your user in database, please relog in", http.StatusForbidden)
		return
	}

	// query by host
	if typo == typohost{
		rows, err := db.Query("select orders.number, orders.tenant, orders.owner, orders.address, orders.status from orders inner join client on orders.owner = client.name where client.jwt = $1;", jwt)
		if err != nil {
			http.Error(w, "could not find your orders in database.", http.StatusForbidden)
			return
		}
 
		orders := make([]Orde, 0)
		
		for rows.Next() {
			ord := Orde{}
			err := rows.Scan(&ord.number, &ord.tenant, &ord.owner, &ord.address, &ord.status) // order matters
			if err != nil {
				http.Error(w, "No orders found, line 192", 500)
				return
			}
			orders = append(orders, ord)
		}
		
	
		if len(orders) == 0{
			fmt.Fprintf(w, "could not found your orders")
		}
		if err = rows.Err(); err != nil {
			http.Error(w, "No orders found, line 199", 500)
			return
		}
	
		for _, ord := range orders {
			fmt.Fprintf(w, "order id:  %s \nTenant:    %s \nHost:      %s \nAddress:   %s \nStatus:    %s \n\n\n", ord.number, ord.tenant, ord.owner, ord.address, ord.status)
		}
		return 
	}


	//query by tenant 
	if typo == typotenant{
		rows, err := db.Query("select orders.number, orders.tenant, orders.owner, orders.address, orders.status from orders inner join client on orders.tenant = client.name where client.jwt = $1;", jwt)
		if err != nil {
			http.Error(w, "could not find your orders in database.", http.StatusForbidden)
			return
		}
 
		orders := make([]Orde, 0)
		
		for rows.Next() {
			ord := Orde{}
			err := rows.Scan(&ord.number, &ord.tenant, &ord.owner, &ord.address, &ord.status) // order matters
			if err != nil {
				http.Error(w, "No orders found, line 192", 500)
				return
			}
			orders = append(orders, ord)
		}
		
	
		if len(orders) == 0{
			fmt.Fprintf(w, "could not found your orders")
		}
		if err = rows.Err(); err != nil {
			http.Error(w, "No orders found, line 199", 500)
			return
		}
	
		for _, ord := range orders {
			fmt.Fprintf(w, "order id:  %s \nTenant:    %s \nHost:      %s \nAddress:   %s \nStatus:    %s \n\n\n", ord.number, ord.tenant, ord.owner, ord.address, ord.status)
		}
		return 
	}
}


// Host comfirm orders
func operateorder(w http.ResponseWriter, r *http.Request){
	cookiee, err := r.Cookie("session")
	if err != nil {
		http.Error(w, "You have to login first", http.StatusForbidden)
		return
	}

	if r.Method == http.MethodGet{
		http.Error(w, "You have to send booking message by post message", http.StatusForbidden)
		return
	}

	
	var checkjwt    string
	var checkstatus string
	jwt          := cookiee.Value
	orderid      := r.FormValue("id")
	newstatus    := r.FormValue("newstatus")

	if !(newstatus == "comfirmed" || newstatus == "rejected"){
		_,_ = io.WriteString(w, "your new order status must be \"comfirmed\" or \"rejected\".")
		return
	}

	if orderid == "" || newstatus == ""{
		http.Error(w, "Please choose order id and new status", http.StatusForbidden)
		return
	}

	row := db.QueryRow("SELECT jwt from client inner join orders on client.name = orders.owner where orders.number = $1;", orderid)
	err  = row.Scan(&checkjwt)
	if err != nil{
		http.Error(w, "Sorry could not find your house details in database", 404)
		return
	}

	if jwt != checkjwt{
		http.Error(w, "You cannot change other's orders' status", http.StatusForbidden)
		return
	}

	row1 := db.QueryRow("SELECT status from orders where number = $1;", orderid)
	_     = row1.Scan(&checkstatus)
	if checkstatus != "wait Host comfirming"{
		http.Error(w, "you alreay updated the status of this order, cannot modify again", http.StatusForbidden)
		return
	}
	if checkstatus == "canceled"{
		http.Error(w, "This order already been canceled.", http.StatusForbidden)
		return
	}


	_, err = db.Exec("update orders set status=$1,ownerchecked='done' ,tenantchecked='not'  where number=$2;", newstatus, orderid)
	if err != nil{
		http.Error(w, "Failed to update your order", http.StatusInternalServerError)
		return
	}

	// if not using host comfirm system, commit this if
	if newstatus == "rejected"{
		var startdate string
		var enddate   string
		var olddates  string
		var address   string
		row := db.QueryRow("SELECT address from orders where number = $1;", orderid)
		_ = row.Scan(&address)
		row2 := db.QueryRow("SELECT startdate from orders where number = $1;", orderid)
		_ = row2.Scan(&startdate)
		row3 := db.QueryRow("SELECT enddate from orders where number = $1;", orderid)
		_ = row3.Scan(&enddate)
		row4 := db.QueryRow("SELECT unavaliabledates from house where address = $1;", address)
		_ = row4.Scan(&olddates)
		//fmt.Println(startdate,enddate,olddates)

		newdates := deletedates(startdate, enddate, olddates) 
		_, err = db.Exec("update house set unavaliabledates=$1 where address=$2;", newdates, address)
		//fmt.Println(newdates,address)
		if err != nil{
			http.Error(w, "Failed to update your order", http.StatusInternalServerError)
			return
		}
	}
	_,_ = io.WriteString(w, "you have successfully update the order status.")
}

func deletedates(start string, end string, storage string) (string){
	arr1 := strings.Split(start,"/")
	arr2 := strings.Split(end,"/")
	unavaliable := strings.Split(storage,",")
	startday,_ := strconv.Atoi(arr1[2])
	endday,_   := strconv.Atoi(arr2[2])
	gap := endday - startday
	if gap <= 0{
		log.Fatal("end date shold later than start date")
	}
	var dates []string
	for i := 0; i < gap ; i++{
		dates = append(dates, strconv.Itoa(startday))
		startday++
	}
	var date []string
	for i := 0 ; i<gap; i++{
		a := arr1[0] + "/" + arr2[1] + "/" + dates[i]
		date = append(date, a)
	}
	
	var newslice []string
	var fla string
	for i := 0; i < len(unavaliable); i++{
		fla = "good"
		for j := 0; j < len(date); j++{
				if unavaliable[i] == date[j]{
					fla = "bad"
				}	
			}
			if fla=="good"{newslice = append(newslice,unavaliable[i])}
		}
	//fmt.Println(newslice)
	var new string
	for i:=0; i<len(newslice); i++{
		new = new +newslice[i] + ","
	}
	//fmt.Println(new)
	return new
}

func cancel(w http.ResponseWriter, r *http.Request){
	cookiee, err := r.Cookie("session")
	if err != nil {
		http.Error(w, "You have to login first", http.StatusForbidden)
		return
	}

	if r.Method == http.MethodGet{
		http.Error(w, "You have to send cancel message by post message", http.StatusForbidden)
		return
	}
	var check string
	var checkstatus string
	jwt := cookiee.Value
	orderid := r.FormValue("id")
	row := db.QueryRow("select jwt from client inner join orders on client.name=orders.tenant where orders.number=$1",orderid)
	err = row.Scan(&check)
	if err != nil{
		http.Error(w, "Could not find your order details, 434", http.StatusForbidden)
		return
	}
	if check != jwt{
		http.Error(w, "You cannot modify other's order", http.StatusForbidden)
		return
	}
	if orderid == ""{
		http.Error(w, "you must enter orderid when cancel order", http.StatusForbidden)
		return
	}
	row1 := db.QueryRow("SELECT status from orders where number = $1;", orderid)
	_     = row1.Scan(&checkstatus)
	if checkstatus == "canceled"{
		http.Error(w, "you cannot operate canceled orders", http.StatusForbidden)
		return
	}

	var startdate string
	var enddate   string
	var olddates  string
	var address   string
	row = db.QueryRow("SELECT address from orders where number = $1;", orderid)
	_ = row.Scan(&address)
	row2 := db.QueryRow("SELECT startdate from orders where number = $1;", orderid)
	_ = row2.Scan(&startdate)
	row3 := db.QueryRow("SELECT enddate from orders where number = $1;", orderid)
	_ = row3.Scan(&enddate)
	row4 := db.QueryRow("SELECT unavaliabledates from house where address = $1;", address)
	_ = row4.Scan(&olddates)
	//fmt.Println(startdate,enddate,olddates)

	newdates := deletedates(startdate, enddate, olddates) 
	_, err = db.Exec("update house set unavaliabledates=$1 where address=$2;", newdates, address)
	//fmt.Println(newdates,address)
	if err != nil{
		http.Error(w, "Failed to update your order", http.StatusInternalServerError)
		return
	}

	_, err = db.Exec("update orders set status='canceled' where number=$1;", orderid)
	if err != nil{
		http.Error(w, "Failed to update the order status", http.StatusInternalServerError)
		return
	}
	_,_ = io.WriteString(w,"you have successfully cancelled the order.")
}
