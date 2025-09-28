package main

import (
	"fmt"
)

func add(a, b float64) float64 {
	return a + b
}

func removeMain() {
	var a, b float64
	fmt.Println("Enter the values of a and b:")
	fmt.Scanf("%f %f", &a, &b)
	sum := add(a, b)
	fmt.Println("The sum is:", sum)
}