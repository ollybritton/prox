package prox

import (
	"sync"
	"time"
)

// waitTimeout waits for the waitgroup for the specified max timeout.
// Returns true if waiting timed out. Code provided courtesy of stack overflow.
func waitTimeout(wg *sync.WaitGroup, timeout time.Duration) bool {
	c := make(chan struct{})

	go func() {
		defer close(c)
		wg.Wait()
	}()

	select {
	case <-c:
		return false // completed normally
	case <-time.After(timeout):
		return true // timed out
	}
}
