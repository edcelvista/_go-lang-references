package main

import "fmt"

type Engine struct{}

func (e Engine) Start(name string) {
	fmt.Printf("%v Engine Starts...\n", name)
}
