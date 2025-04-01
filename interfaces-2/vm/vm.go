package vm

import "fmt"

type VendingMachine struct{}

func New() *VendingMachine {
	return &VendingMachine{}
}

func (this VendingMachine) GetDrink(money uint64, brand string) string {
	return fmt.Sprintf("Ice Cold %s", brand)
}
