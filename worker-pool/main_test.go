package main

import (
	"strconv"
	"testing"
)

// Test that all 100 results are returned.
func TestWorkerPool_ReturnsAllResults(t *testing.T) {
	results := runWorkerPool(100, 5)
	if len(results) != 100 {
		t.Fatalf("expected 100 results, got %d", len(results))
	}
}

// Test that every result is the correct square.
func TestWorkerPool_CorrectSquares(t *testing.T) {
	results := runWorkerPool(100, 5)
	for _, r := range results {
		expected := r.ID * r.ID
		if r.Value != expected {
			t.Errorf("task %d: expected %d, got %d", r.ID, expected, r.Value)
		}
	}
}

// Test that results are sorted by ID.
func TestWorkerPool_ResultsInOrder(t *testing.T) {
	results := runWorkerPool(100, 5)
	for i := 1; i < len(results); i++ {
		if results[i].ID <= results[i-1].ID {
			t.Errorf("results not sorted: ID %d at index %d comes after ID %d",
				results[i].ID, i, results[i-1].ID)
		}
	}
}

// Test that IDs cover exactly 1..numTasks with no duplicates.
func TestWorkerPool_NoDuplicateIDs(t *testing.T) {
	const numTasks = 100
	results := runWorkerPool(numTasks, 5)

	seen := make(map[int]bool)
	for _, r := range results {
		if seen[r.ID] {
			t.Errorf("duplicate result for task ID %d", r.ID)
		}
		seen[r.ID] = true
	}
	for i := 1; i <= numTasks; i++ {
		if !seen[i] {
			t.Errorf("missing result for task ID %d", i)
		}
	}
}

// Test with different worker counts to prove concurrency level doesn't affect correctness.
func TestWorkerPool_VariousWorkerCounts(t *testing.T) {
	for _, workers := range []int{1, 2, 5, 10, 50} {
		t.Run(
			"workers="+strconv.Itoa(workers),
			func(t *testing.T) {
				results := runWorkerPool(100, workers)
				if len(results) != 100 {
					t.Fatalf("workers=%d: expected 100 results, got %d", workers, len(results))
				}
				for _, r := range results {
					if r.Value != r.ID*r.ID {
						t.Errorf("workers=%d, task %d: expected %d, got %d",
							workers, r.ID, r.ID*r.ID, r.Value)
					}
				}
			},
		)
	}
}
