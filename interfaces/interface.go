package main

import "fmt"

type cat struct {
	name string
}

func (c *cat) say() {
	fmt.Printf("Cat named %v says: Meow!\n", c.name)
}

type dog struct {
	name string
}

func (d *dog) say() {
	fmt.Printf("Dog named %v says: Woff!\n", d.name)
}

type cow struct {
	name string
}

func (c *cow) say() {
	fmt.Printf("Cow named %v says: Mooo!\n", c.name)
}

type IAnimal interface {
	say()
}

func main() {
	animals := []IAnimal{
		&cat{
			name: "Cato",
		},
		&dog{
			name: "Dogo",
		},
		&cow{
			name: "Cowo",
		},
	}

	for _, k := range animals {
		k.say()
	}
}
