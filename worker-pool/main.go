package main

import (
	"fmt"
	"sort"
	"sync"
)

// Task represents a unit of work: square the number.
type Task struct {
	ID    int
	Value int
}

// Result holds the output of a completed task.
type Result struct {
	ID    int
	Value int
}

// runWorkerPool dispatches numTasks tasks (square 1..numTasks) across numWorkers
// goroutines and returns the results sorted by ID.
func runWorkerPool(numTasks, numWorkers int) []Result {
	tasks := make(chan Task, numTasks)
	results := make(chan Result, numTasks)

	// Start workers.
	var wg sync.WaitGroup
	for w := 0; w < numWorkers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for t := range tasks {
				results <- Result{ID: t.ID, Value: t.Value * t.Value}
			}
		}()
	}

	// Send tasks.
	for i := 1; i <= numTasks; i++ {
		tasks <- Task{ID: i, Value: i}
	}
	close(tasks)

	// Wait for all workers to finish, then close results.
	wg.Wait()
	close(results)

	// Collect and sort results by ID.
	collected := make([]Result, 0, numTasks)
	for r := range results {
		collected = append(collected, r)
	}
	sort.Slice(collected, func(i, j int) bool {
		return collected[i].ID < collected[j].ID
	})

	return collected
}

func main() {
	const (
		numTasks   = 100
		numWorkers = 5
	)

	results := runWorkerPool(numTasks, numWorkers)

	// Print results in order.
	for _, r := range results {
		fmt.Printf("Task %3d: %d² = %d\n", r.ID, r.ID, r.Value)
	}
}
