package main

import (
	"fmt"
)

func main() {
	// var name string = "kantitad"
	var age int = 25

	email := "sisawat_k@su.ac.th"
	gpa := 3.75

	firstname, lastname := "kantitad", "sisawat"

	fmt.Printf("Name %s %s, age %d, email %s, gpa %.2f\n", firstname, lastname, age, email, gpa)
}