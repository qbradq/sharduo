package world

import (
	"log"
	"sync"
)

var waitGroup sync.WaitGroup

// Stop must be called to end the service's goroutine
func Stop() {
	close(serviceRequests)
}

// Start is the world service's main goroutine. Call it as "go world.Start()"
func Start() {
	waitGroup.Add(1)
	defer waitGroup.Done()

	startInstances()
	defer stopInstances()

	for req := range serviceRequests {
		switch r := req.(type) {
		case *NewCharacterRequest:
			doNewCharacter(r)
		default:
			log.Fatalln("Bad world request", req)
		}
	}
}

// Channel for service requests (NOT game requests)
var serviceRequests = make(chan interface{}, 100)

// ServiceRequest sends top-level service requests
func ServiceRequest(r interface{}) bool {
	select {
	case serviceRequests <- r:
		return true
	default:
		return false
	}
}
