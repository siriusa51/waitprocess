package waitprocess

import (
	"os"
	"sync"
	"time"
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
func Run() error {
	return Default().Run()
}

// Error returns the error of the WaitProcess.
func Error() error {
	return Default().Error()
}

// Stop stops the WaitProcess.
func Stop() {
	Default().Stop()
}

// Wait waits for the WaitProcess to stop.
func Wait(timeout ...time.Duration) error {
	return Default().Wait(timeout...)
}

// Shutdown stops the WaitProcess.
func Shutdown(timeout ...time.Duration) error {
	return Default().Shutdown(timeout...)
}

// PreStartHook adds a hook to be run before the waitprocess starts
func PreStartHook(name string, h hookFunc) *WaitProcess {
	return Default().PreStartHook(name, h)
}

// AfterStopHook adds a hook to be run after the waitprocess stops
func AfterStopHook(name string, h hookFunc) *WaitProcess {
	return Default().AfterStopHook(name, h)
}
