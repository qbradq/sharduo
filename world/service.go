package world

import (
	"sync"
)

// SendRequest recieve structs in the world.*Request class for dispatch
func SendRequest(r interface{}) {

}

// Stop must be called to end the service's goroutine
func Stop() {
	close(serviceRequests)
}

// Service is the world service's main goroutine
func Service(wg *sync.WaitGroup) {
	wg.Add(1)
	defer wg.Done()

	startInstances(wg)
	defer stopInstances()

	for {
		_, open := <-serviceRequests
		if open == false {
			break
		}
		/*
			switch t := r.(type) {
			}
		*/
	}
}

// Channel for service requests (NOT game requests)
var serviceRequests = make(chan interface{}, 100)
