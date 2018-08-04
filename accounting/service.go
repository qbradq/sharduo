package accounting

// Start kicks off the accounting goroutine
//func Start(wg *sync.WaitGroup) {
func Start() {
	go func() {
		//wg.Add(1)
		mainLoop()
		//wg.Done()
	}()
}

func mainLoop() {
	select {
	case r := <-loginChannel:
		doLogin(r)
	}
}
