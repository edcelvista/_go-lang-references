package main

import "fmt"

type Wheel struct{}

func (w *Wheel) TurnLeft(name string) {
	fmt.Printf("%v Turn Left...\n", name)
}

func (w *Wheel) TurnRight(name string) {
	fmt.Printf("%v Turn Right...\n", name)
}
