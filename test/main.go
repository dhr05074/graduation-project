package main

import (
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"
)

const (
	numWorkers        = 500
	requestsPerWorker = 100
	requestsPerSecond = 20
)

func main() {
	reportFile, err := os.Create("report.csv")
	if err != nil {
		panic(err)
	}
	defer reportFile.Close()

	// Define the endpoints to test.
	endpoints := []string{
		"http://localhost:10000/high",
		"http://localhost:10000/mid",
		"http://localhost:10000/low",
	}

	// Create a wait group to synchronize goroutines.
	var wg sync.WaitGroup

	// Create a map to store the success rates for each endpoint.
	var successRates []int

	for i := 0; i < len(endpoints); i++ {
		successRates = append(successRates, 0)
	}

	go func() {
		for range time.Tick(1 * time.Second) {
			fmt.Printf("successRates: %+v\n", successRates)

			line := "\n"

			for i := 0; i < len(successRates); i++ {
				line += fmt.Sprintf("%d,", successRates[i])
			}

			reportFile.WriteString(line)
		}
	}()

	// Create a function to make requests to an endpoint.
	makeRequests := func(endpoint string, idx int) {
		defer wg.Done()

		for i := 0; i < requestsPerWorker; i++ {
			// Create an HTTP client.
			client := &http.Client{}

			// Send an HTTP GET request to the endpoint.
			resp, err := client.Get(endpoint)
			if err != nil {
				//fmt.Printf("Error making request to %s: %v\n", endpoint, err)
				continue
			}

			if resp.StatusCode == http.StatusOK {
				successRates[idx] += 1
			}

			// Close the response body.
			resp.Body.Close()

			time.Sleep(1 * time.Second / requestsPerSecond)
		}
	}

	// Launch goroutines to make requests to each endpoint.
	for k, endpoint := range endpoints {
		for i := 0; i < numWorkers; i++ {
			wg.Add(1)
			go makeRequests(endpoint, k)
		}
	}

	// Wait for all goroutines to finish.
	wg.Wait()

	// Print the success rates for each endpoint.
	fmt.Println("Success rates for each endpoint:")
	for i, endpoint := range endpoints {
		fmt.Printf("%s: %d\n", endpoint, successRates[i])
	}
}
