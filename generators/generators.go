package main

import (
	"fmt"
	"math/rand"
	"runtime"
	"sync"
)

func repeatFunc[T any, K any](done <-chan K, fn func() T) <-chan T {
	stream := make(chan T)

	go func() {
		defer close(stream)
		for {
			select {
			case <-done:
				return
			case stream <- fn():
			}
		}
	}()

	return stream
}

func take[T any, K any](done <-chan K, stream <-chan T, n int) <-chan T {
	taken := make(chan T) // unbuffered channel makes channel communication sync

	go func() {
		defer close(taken)
		for i := 0; i < n; i++ {
			select {
			case <-done:
				return
			case taken <- <-stream:
			}
		}
	}()

	return taken
}

func primeFidner(done <-chan int, randIntStream <-chan int) <-chan int {
	isPrime := func(randomInt int) bool {
		for i := randomInt - 1; i > 1; i-- {
			if randomInt%i == 0 {
				return false
			}
		}
		return true
	}

	primes := make(chan int)
	go func() {
		defer close(primes)
		for {
			select {
			case <-done:
				return
			case randomInt := <-randIntStream:
				if isPrime(randomInt) {
					primes <- randomInt
				}
			}
		}
	}()

	return primes
}

func fanIn[T any](done <-chan int, channels ...<-chan T) <-chan T {
	var wg sync.WaitGroup
	fannedInStream := make(chan T)

	transfer := func(c <-chan T) {
		defer wg.Done() // resolve the wait
		for i := range c {
			select {
			case <-done:
				return
			case fannedInStream <- i:
			}
		}
	}

	for _, c := range channels {
		wg.Add(1) // add a wait flag to wait for the response
		go transfer(c)
	}

	go func() {
		wg.Wait()
		close(fannedInStream)
	}()

	return fannedInStream
}

func main() {
	done := make(chan int)
	defer close(done)

	randNumFetcher := func() int { return rand.Intn(500000000) }
	randNumStream := repeatFunc(done, randNumFetcher) // data generator

	// naive
	// primeStream := primeFidner(done, randNumStream)
	// for rando := range take(done, primeStream, 10) { //repeatFunc(done, randNumFetcher) {
	// 	fmt.Println(rando)
	// }

	// fan out
	CPUCount := runtime.NumCPU()
	primeFunderChannels := make([]<-chan int, CPUCount)
	for i := 0; i < CPUCount; i++ {
		primeFunderChannels[i] = primeFidner(done, randNumStream)
	}

	// fan in
	fannedInStream := fanIn(done, primeFunderChannels...)
	for rando := range take(done, fannedInStream, 10) { //repeatFunc(done, randNumFetcher) {
		fmt.Println(rando)
	}
}
