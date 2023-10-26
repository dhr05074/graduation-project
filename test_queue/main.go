package main

import (
	"fmt"
	"time"
)

// Task represents a task with a priority.
type Task struct {
	Priority int
	Message  string
}

func worker(id int, tasks <-chan Task) {
	for task := range tasks {
		fmt.Printf("Worker %d is processing task: %s (Priority: %d)\n", id, task.Message, task.Priority)
		time.Sleep(time.Second) // Simulate some work
		fmt.Printf("Worker %d finished task: %s\n", id, task.Message)
	}
}

func main() {
	numWorkers := 3
	taskQueue := make(chan Task, 10)

	// Start worker goroutines
	for i := 1; i <= numWorkers; i++ {
		go worker(i, taskQueue)
	}

	// Enqueue tasks with different priorities
	tasks := []Task{
		{Priority: 1, Message: "Low priority task 1"},
		{Priority: 2, Message: "Medium priority task 1"},
		{Priority: 1, Message: "Low priority task 2"},
		{Priority: 3, Message: "High priority task 1"},
		{Priority: 2, Message: "Medium priority task 2"},
		{Priority: 3, Message: "High priority task 2"},
	}

	// Enqueue tasks with priorities
	for _, task := range tasks {
		taskQueue <- task
	}

	close(taskQueue)

	// Wait for workers to finish
	for i := 1; i <= numWorkers; i++ {
		<-time.After(time.Second)
	}
}
