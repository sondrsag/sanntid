package main

import (
	//."driver"
	//."fmt"	
	//."time"
	"control"
	)
	
func main() {
	
	
	asd := make(chan int)

	go control.InitElevator()
	

	<- asd
	
	return
}
