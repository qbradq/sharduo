package common

// A Service manages one or more goroutines that process requests.
type Service interface {
	// Start is the main goroutine of this service. Start it with "go s.Start()".
	Start()
	// Stop blocks until all of the service's goroutines have exited
	Stop()
	// SendRequest sends a request to the service and returns false only if the
	// underlying service is out of resources. SendRequest is thread-safe.
	SendRequest(r interface{}) bool
}
