package main

import (
	"fmt"
	"sync"
)

func main() {
	var mu sync.Mutex
	var wg sync.WaitGroup
	var counter int

	for i := 0; i < 1000; i++ {
		wg.Add(1)

		go func() {
			defer wg.Done()

			mu.Lock()
			counter++
			mu.Unlock()
		}()
	}

	wg.Wait() // ✅ wait for all goroutines

	fmt.Println(counter) // return consistent 1000 values
}

// WITH RACE CONDITIONS
// import (
// 	"fmt"
// )

// func main() {
// 	var counter int

// 	for i := 0; i < 1000; i++ {
// 		go func() {
// 			counter++
// 		}()
// 	}

// 	fmt.Println(counter) // return inconsistent values
// }
