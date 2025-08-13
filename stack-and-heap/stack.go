package main

import "fmt"

func main() {
	greet("John")
}

func greet(name string) {
	message := "Hello, " + name
	fmt.Println(message)
}
