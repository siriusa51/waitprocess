package waitprocess

import "context"

type stopFuncProcess struct {
	runFunc  func() error
	stopFunc func()
}

// RunWithStopFunc create a Process with runFunc and stopFunc, call stop to stop runFunc
func RunWithStopFunc(runFunc func() error, stopFunc func()) Process {
	return &stopFuncProcess{
		runFunc:  runFunc,
		stopFunc: stopFunc,
	}
}

func (p *stopFuncProcess) SetContext(_ context.Context) {
}

func (p *stopFuncProcess) Run() error {
	return p.runFunc()
}

func (p *stopFuncProcess) Stop() {
	p.stopFunc()
}
