package main

import (
	"fmt"
	"sync"
)

// Mutext Locking of critical section
var lock sync.Mutex

func processData(wg *sync.WaitGroup, result *[]int, data int) {
	defer wg.Done()
	processedData := data * 2
	lock.Lock() // locking critical section | caution with high performance impact
	*result = append(*result, processedData)
	lock.Unlock() // locking critical section
}

func main() {
	var wg sync.WaitGroup

	input := []int{1, 2, 3, 4, 5}
	result := []int{}

	for _, data := range input {
		wg.Add(1)
		processData(&wg, &result, data)
	}

	wg.Wait()

	fmt.Println(result)
}

// confinement
// func processData(wg *sync.WaitGroup, resultDest *int, data int) {
// 	defer wg.Done()
// 	processedData := data * 2
// 	*resultDest = processedData
// }

// func main() {
// 	var wg sync.WaitGroup

// 	input := []int{1, 2, 3, 4, 5}
// 	result := make([]int, len(input))

// 	for i, data := range input {
// 		wg.Add(1)
// 		processData(&wg, &result[i], data)
// 	}

// 	wg.Wait()
// 	fmt.Println(result)
// }
