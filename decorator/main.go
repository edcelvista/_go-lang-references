package main

import "fmt"

type StringFunc func(string) string

func greet(name string) string {
	return fmt.Sprintf("Hello, %s", name)
}

func middleware(f StringFunc) StringFunc {
	return func(s string) string {
		res := f(s)
		return res
	}
}

func main() {
	decoratedGreet := middleware(greet)
	res := decoratedGreet("Edcel")
	fmt.Println(res)
}
