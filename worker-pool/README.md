## Worker Pool in Golang
This project demonstrates worker pool in Golang. The implementation use sync.WaitGroup to run worker and use buffered channels to limit the number of concurrent workers.


```golang
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
```