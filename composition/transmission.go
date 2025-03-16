package main

import "fmt"

type Transmission struct{}

func (t Transmission) ShiftDown(name string) {
	fmt.Printf("%v Shift Down...\n", name)
}

func (t Transmission) ShiftUp(name string) {
	fmt.Printf("%v Shift Up...\n", name)
}
