package main

import (
	"fmt"
	"net/http"
	"sync"
	"time"
)

func makeRequest(url string, priorityValue int, resultChan chan<- int) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered in f", r)
		}
	}()

	// Create an HTTP client
	client := &http.Client{}

	// Create an HTTP request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Println("Error creating request:", err)
		return
	}

	// Set the 'X-Priority' header with the given value
	req.Header.Set("X-Priority", fmt.Sprintf("%d", priorityValue))

	// Make the HTTP request
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error making request:", err)
		return
	}

	// Ensure the response body is closed
	defer resp.Body.Close()

	// Check if the response code is 200 (success)
	if resp.StatusCode == http.StatusOK {
		resultChan <- priorityValue
	}
}

func main() {
	url := "http://localhost:10000"
	numRequests := 1000
	numWorkers := 5
	resultChan := make(chan int, numRequests*numWorkers)

	var wg sync.WaitGroup

	// Run workers to make requests in parallel
	for k := 0; k < 5; k++ {
		for i := 1; i <= numWorkers; i++ {
			wg.Add(1)
			go func(workerID int) {
				defer wg.Done()
				for j := 1; j <= numRequests; j++ {
					makeRequest(url, workerID, resultChan)
					time.Sleep(10 * time.Millisecond)
				}
			}(i)
		}
	}

	// Wait for fixed amount of time to count responses
	time.Sleep(5 * time.Second)

	// Close the result channel to signal that we're done
	close(resultChan)

	// Collect and count success rates for each X-Priority value
	successCounts := make(map[int]int)
	for success := range resultChan {
		successCounts[success]++
	}

	// Print success count for each X-Priority value
	for i := 1; i <= 5; i++ {
		successCount, exists := successCounts[i]
		if !exists {
			successCount = 0
		}
		fmt.Printf("X-Priority: %d, Successes: %d\n", i, successCount)
	}
}
