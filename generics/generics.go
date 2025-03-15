package main

import (
	"fmt"

	"golang.org/x/exp/constraints"
)

type Profile[T constraints.Ordered] struct {
	ID     int
	Name   string
	Age    int
	Gender string
	Gpa    T
}

func (p *Profile[T]) get() {
	fmt.Printf("Name: %v with Age %v %v has GPA of %v", p.Name, p.Age, p.Gender, p.Gpa)
}

func main() {
	profile := Profile[int]{
		ID:     1,
		Name:   "Edcel",
		Age:    29,
		Gender: "Male",
		Gpa:    5,
	}

	profile.get()
}
