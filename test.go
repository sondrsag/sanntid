package main

import(
	"fmt"
	)


	
func main(){
	type Matrix [4][3]int
	var matr Matrix
	matr[1][1] = 5
	fmt.Println(matr)

	for i := 0; i <= 5; i++ {
		fmt.Println(i)
	}
}


