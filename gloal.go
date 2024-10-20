package waitprocess

import (
	"os"
	"sync"
)

var (
	instence *WaitProcess
	once     sync.Once
)

// Default returns the default WaitProcess instance.
func Default() *WaitProcess {
	once.Do(func() {
		instence = NewWaitProcess()
	})

	return instence
}

// RegisterProcess registers a process with the given name and Process.
func RegisterProcess(name string, procs Process) *WaitProcess {
	return Default().RegisterProcess(name, procs)
}

// RegisterSignal registers a signal with the given os.Signal.
func RegisterSignal(sigs ...os.Signal) *WaitProcess {
	return Default().RegisterSignal(sigs...)
}

// Start starts the WaitProcess.
func Start() {
	Default().Start()
}

// Run starts the WaitProcess and waits for it to stop.
func Run() {
	Default().Run()
}

// Stop stops the WaitProcess.
func Stop() {
	Default().Stop()
}

// Wait waits for the WaitProcess to stop.
func Wait() {
	Default().Wait()
}

// Shutdown stops the WaitProcess.
func Shutdown() {
	Default().Shutdown()
}
