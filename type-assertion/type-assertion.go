package main

import (
	"log"
)

func process[T any](value T) { // type switch & generics
	switch v := any(value).(type) {
	case int:
		log.Println("Integer:", v)
	case string:
		log.Println("String:", v)
	default:
		log.Println("Unknown:")
	}
}

func main() {
	var a any = "Hello World"

	b := a

	if v, ok := b.(string); ok {
		log.Println(ok, v)
	}

	c := 3
	process(c)
}
