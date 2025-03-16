package main

import "fmt"

type truck struct {
	Engine
	Transmission
	Wheel
}

func (t *truck) fourWheelDrive(name string) {
	fmt.Printf("%v 4WD ...\n", name)
}

type sedan struct {
	Engine
	GenericTransmission
	Wheel
}

func (s *sedan) AutoPark(name string) {
	fmt.Printf("%v Auto Parking ...\n", name)
}

type startable interface {
	Start(n string)
}

func startEngines(cars ...startable) {
	for i, v := range cars {
		v.Start(fmt.Sprintf("Car %v", i))
	}
}

type GenericTransmission interface {
	ShiftDown(s string)
	ShiftUp(s string)
}

func main() {
	truck1 := truck{Engine{}, Transmission{}, Wheel{}}
	truck1.ShiftDown("Truck 1")
	truck1.fourWheelDrive("Truck 1")

	sedan1 := sedan{Engine{}, ETransmission{}, Wheel{}}
	sedan1.ShiftDown("Sedan 1")
	sedan1.AutoPark("Sedan 1")

	startEngines(truck1, sedan1)
}
