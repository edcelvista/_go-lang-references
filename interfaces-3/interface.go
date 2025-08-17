package main

import (
	"fmt"
)

// Define an interface
type Shape interface {
	Area() float64
	Perimeter() float64
}

// Define a struct: Rectangle
type Rectangle struct {
	Width, Height float64
}

// Implement Shape interface for Rectangle
func (r Rectangle) Area() float64 {
	return r.Width * r.Height
}

func (r Rectangle) Perimeter() float64 {
	return 2 * (r.Width + r.Height)
}

// Define a struct: Circle
type Circle struct {
	Radius float64
}

// Implement Shape interface for Circle
func (c Circle) Area() float64 {
	return 3.14 * c.Radius * c.Radius
}

func (c Circle) Perimeter() float64 {
	return 2 * 3.14 * c.Radius
}

// Function that takes an interface
func printShapeInfo(s Shape) {
	fmt.Printf("Area: %.2f\n", s.Area())
	fmt.Printf("Perimeter: %.2f\n", s.Perimeter())
}

func main() {
	r := Rectangle{Width: 10, Height: 5}
	c := Circle{Radius: 7}

	// Both Rectangle and Circle satisfy Shape interface
	printShapeInfo(r)
	printShapeInfo(c)
}
