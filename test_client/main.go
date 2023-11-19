package main

import (
	"fmt"
	"github.com/go-resty/resty/v2"
	"sync"
	"time"
)

func main() {
	success := make(chan int, 2<<10)
	failure := make(chan int, 2<<10)
	c := make(chan struct{})

	successCount := map[int]int{}
	failureCount := map[int]int{}

	var n = 5

	go func(clear <-chan struct{}) {
		tick := time.Tick(1 * time.Second)
		for {
			select {
			case p := <-success:
				successCount[p] += 1
			case p := <-failure:
				failureCount[p] += 1
			case <-clear:
				successCount = map[int]int{}
				failureCount = map[int]int{}
			case <-tick:
				Stat(n, successCount, failureCount)
			}
		}
	}(c)

	wg := &sync.WaitGroup{}
	wg.Add(1)

	for n = 0; n < 500; n++ {
		go func() {
			cli := resty.New()
			for {
				for i := 1; i <= 5; i++ {
					time.Sleep(1 * time.Millisecond)

					res, err := cli.R().SetHeader("X-Priority", fmt.Sprintf("%d", i)).Get("http://localhost:8080")
					if err != nil {
						failure <- i
						continue
					}

					if res.StatusCode() != 200 {
						failure <- i
						continue
					}

					success <- i
				}
				time.Sleep(1 * time.Millisecond)
			}
		}()
	}
	wg.Wait()
}

func Stat(n int, success map[int]int, failure map[int]int) {
	stat := fmt.Sprintf("%d", n)
	for k, v := range success {
		stat += fmt.Sprintf(",%f", float64(v)/(float64(v)+float64(failure[k])))
	}
	stat += fmt.Sprintf("\n")

	fmt.Printf(stat)
}
