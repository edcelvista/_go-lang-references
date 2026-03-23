package main

import (
	"context"
	"fmt"
	"math/rand"
	"time"
)

func openConnection(done chan bool) {
	fmt.Println("Attempting Connection....")

	if rand.Intn(100) > 50 {
		fmt.Println("Connection Failure!")
		time.Sleep(100000 * time.Hour)
	} else {
		time.Sleep(2 * time.Second)
		fmt.Println("Connection Established!")
	}

	done <- true
}

func openConnectionWithTimeout() {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	done := make(chan bool)
	go openConnection(done)

	select {
	case <-done:
		fmt.Println("Connection Succesfully!")
	case <-ctx.Done():
		fmt.Println("Connection Timeout")
	}
}

func main() {
	openConnectionWithTimeout()
}
