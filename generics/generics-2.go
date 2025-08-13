package main

import "fmt"

func sumNumbers32(numbers []int32) (sum int32) {
	for _, number := range numbers {
		sum += number
	}
	return
}

func sumNumbers64(numbers []int64) (sum int64) {
	for _, number := range numbers {
		sum += number
	}
	return
}

func main() {
	totalInt32 := sumNumbers32([]int32{1, 2, 3, 4, 5})
	fmt.Printf("The sum of the numbers is: %d\n", totalInt32)

	totalInt64 := sumNumbers64([]int64{15555, 2555555, 3555555, 455555, 555555})
	fmt.Printf("The sum of the numbers is: %d\n", totalInt64)

	total := sumNumbers([]int64{1, 2, 3, 4, 5, 15555, 2555555, 3555555, 455555, 555555})
	fmt.Printf("The sum of the numbers is: %d\n", total)

}

// OR GENERICS!

type Number interface {
	int32 | int64
}

func sumNumbers[T Number](numbers []T) (sum T) {
	for _, number := range numbers {
		sum += number
	}
	return
}
