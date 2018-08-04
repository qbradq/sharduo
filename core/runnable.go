package core

import (
	"sync"
)

// A Runnable manages the lifecycle of a goroutine
type Runnable interface {
	Run(wg *sync.WaitGroup)
	Stop()
	Stopping() bool
}
