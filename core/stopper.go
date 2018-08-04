package core

// A Stopper implements part of the Runnable interface
type Stopper struct {
	stop bool
}

// Stop requests the goroutine to exit
func (s *Stopper) Stop() {
	s.stop = true
}

// Stopping returns true when the object is trying to stop
func (s *Stopper) Stopping() bool {
	return s.stop
}
