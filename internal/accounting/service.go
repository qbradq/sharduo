package accounting

import (
	"sync"
)

var waitGroup sync.WaitGroup
var serviceRequests chan interface{}

// Start is the main goroutine of this service. Start it with "go accounting.Start()".
func Start() {
	serviceRequests = make(chan interface{}, 100)

	waitGroup.Add(1)
	defer waitGroup.Done()

	for req := range serviceRequests {
		switch r := req.(type) {
		case *LoginRequest:
			doLogin(r)
		case *SelectServerRequest:
			doSelectServer(r)
		case *GameServerLoginRequest:
			doGameServerLogin(r)
		}
	}
}

// Stop blocks until all of the service's goroutines have exited
func Stop() {
	close(serviceRequests)
	waitGroup.Wait()
}

// SendRequest sends a request to the service
func SendRequest(r interface{}) bool {
	select {
	case serviceRequests <- r:
		return true
	default:
		return false
	}
}
