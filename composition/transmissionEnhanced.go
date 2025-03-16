package main

import "fmt"

type ETransmission struct{}

func (t ETransmission) ShiftDown(name string) {
	fmt.Printf("%v E Shift Down...\n", name)
}

func (t ETransmission) ShiftUp(name string) {
	fmt.Printf("%v E Shift Up...\n", name)
}
