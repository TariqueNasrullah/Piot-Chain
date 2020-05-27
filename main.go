package main

import "fmt"

// Add adds two number
func Add(a, b int) int {
	return a + b
}

func main() {
	fmt.Println(Add(10, 20))
}
