package world

import (
	"sync"

	"github.com/qbradq/sharduo/pkg/uo"
)

// An instance object manages all of the objects and behaviors of a game
// instance. All exported methods are thread-safe.
type instance struct {
	ID       uo.Serial
	requests chan interface{}
	wg       *sync.WaitGroup
}

// SendRequest sends a game request to the instance's goroutine
func (i *instance) SendRequest(r interface{}) {
	i.requests <- r
}

// Stop requests the instance to stop executing
func (i *instance) Stop() {
	close(i.requests)
}

// Map of running instances
var runningInstances = make(map[uo.Serial]*instance)

// Create a new instance
func newInstance(id uo.Serial, wg *sync.WaitGroup) *instance {
	i := &instance{
		ID:       id,
		requests: make(chan interface{}, 1000),
		wg:       wg,
	}
	runningInstances[i.ID] = i
	return i
}

// Starts internal instance and all shared instances
func startInstances(wg *sync.WaitGroup) {
	// The internal instance
	newInstance(1, wg)
}

// Stop all running instances
func stopInstances() {
	for _, i := range runningInstances {
		i.Stop()
	}
}
