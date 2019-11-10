package main

import (
	"fmt"
	"strings"
	"strconv"
	"log"
)

func main() {
	start   := "2019/12/9"
	end     := "2019/12/30"
	storage := "2019/12/5,2019/12/6,2019/12/9,"
	a := judge(start,end,storage)
	fmt.Println(a)
}

func judge(start string, end string, storage string) (bool){

	unavaliable := strings.Split(storage,",")
	arr1 := strings.Split(start,"/")
	arr2 := strings.Split(end,"/")
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
	


	