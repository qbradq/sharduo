package accounting

import (
	"sync"
)

// ServiceRequests recieves structs in the accounting.*Request class for
// processing
var ServiceRequests = make(chan interface{}, 100)

// Service is the accounting service's main goroutine
func Service(wg *sync.WaitGroup) {
	wg.Add(1)
	defer wg.Done()

	for r := range ServiceRequests {
		switch t := r.(type) {
		case *LoginRequest:
			doLogin(t)
		}
	}
}

// Stop must be called to end the service's goroutine
func Stop() {
	close(ServiceRequests)
}
