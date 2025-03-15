package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
	"time"
)

func (u *Targets) fetchRestData(ctx context.Context) (Response, error) {
	resp, err := http.Get(u.endpoint)
	if err != nil {
		log.Printf("ERROR: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("ERROR: %v", err)
	}

	// Declare a variable to store the result
	var r Response

	// Unmarshal the JSON data into the Person struct
	err2 := json.Unmarshal([]byte(string(body)), &r)
	if err2 != nil {
		log.Println("Error unmarshalling JSON:", err2)
		return r, err2
	}

	select {
	case <-ctx.Done(): // in the main context deadline reached tell the wait group to close
		return r, errors.New("context deadline exceeded")
	default:
		return r, nil
	}

	// if <-ctx.Done(); true {
	// 	return r, errors.New("context deadline exceeded")
	// }
}

func (u *Targets) fetchRestDataAsync(ctx context.Context, ch chan Response, wg *sync.WaitGroup) {
	// Declare a variable to store the result
	var r Response

	resp, err := http.Get(u.endpoint)
	if err != nil {
		r.Error = fmt.Sprintf("ERROR: %v", err)
		wg.Done()
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		r.Error = fmt.Sprintf("ERROR: %v", err)
		wg.Done()
	}

	// Unmarshal the JSON data into the Person struct
	err2 := json.Unmarshal([]byte(string(body)), &r)
	if err2 != nil {
		r.Error = fmt.Sprintf("Error unmarshalling JSON: %v", err2)
		wg.Done()
	}

	select {
	case <-ctx.Done(): // in the main context deadline reached tell the wait group to close
		wg.Done()
	case ch <- r:
		wg.Done()
	}
}

type Targets struct {
	name     string
	endpoint string
}

type SignedIntegers interface {
	int | int8 | int16 | int32 | int64
}

type FloatingPointNumbers interface {
	float32 | float64
}

type Response struct {
	Error string    `json: "error"`
	Id    string    `json: "id"`
	Name  string    `json: "name"`
	Data  Data[int] `json: "data,omitempty"`
}

type Data[T SignedIntegers | FloatingPointNumbers] struct {
	year         T      `json: "year,omitempty"`
	price        T      `json: "price,omitempty"`
	CPUModel     string `json: "CPU model,omitempty"`
	HardDiskSize string `json: "Hard disk size,omitempty"`
}

func main() {
	start := time.Now()
	targets := []Targets{
		{name: "1", endpoint: "https://api.restful-api.dev/objects/1"},
		{name: "2", endpoint: "https://api.restful-api.dev/objects/2"},
		{name: "3", endpoint: "https://api.restful-api.dev/objects/3"},
	}

	/* context */
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel() // Ensure the cancel function is called

	/* async */
	ch := make(chan Response, 3)
	wg := &sync.WaitGroup{}
	wg.Add(len(targets))
	for _, v := range targets {
		go v.fetchRestDataAsync(ctx, ch, wg)
	}
	wg.Wait()
	close(ch)

	for v := range ch { /* capture set outside */
		if v.Error != "" {
			log.Printf("Error: %v\n", v.Error)
			continue
		}
		log.Printf("Device: %v\n", v.Name)
	}

	/* sync */
	// for _, v := range targets {
	// 	v, err := v.fetchRestData(ctx)
	// 	if err != nil {
	// 		log.Printf("Error: %v", err)
	// 		continue
	// 	}
	// 	log.Printf("Device: %v\n", v.Name)
	// }

	log.Println("Took:", time.Since(start))
}
