package main

import(
	"strings"
	"fmt"
	"time"
)


func testFunc(ch chan){
	ch <- "fuck"
	return
}

func chanTest() chan {
	ch := make(chan string)
	testFunc(ch)
	return ch
}

func main(){
	ch := chanTest()
	text <- ch
	fmt.println(text)
}