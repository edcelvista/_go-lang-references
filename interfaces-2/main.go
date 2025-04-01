package main

import (
	"fmt"

	"edcelvista.com/go/interfaces-2/vm"
)

type vendingMachine interface {
	GetDrink(money uint64, brand string) string
}

type Application struct {
	vm vendingMachine
}

func (this Application) Run() {
	myDrink := this.vm.GetDrink(100, "Cola")
	fmt.Println(myDrink)
}

func newApplication(vm vendingMachine) *Application {
	return &Application{vm: vm}
}

func main() {
	vendingMachine := vm.New()
	app := newApplication(vendingMachine)
	app.Run()
}
