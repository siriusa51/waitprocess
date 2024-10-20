package waitprocess

import "context"

type stopFuncProcess struct {
	runFunc  func()
	stopFunc func()
}

// RunWithStopFunc create a Process with runFunc and stopFunc, call stop to stop runFunc
func RunWithStopFunc(runFunc func(), stopFunc func()) Process {
	return &stopFuncProcess{
		runFunc:  runFunc,
		stopFunc: stopFunc,
	}
}

func (p *stopFuncProcess) SetContext(_ context.Context) {
}

func (p *stopFuncProcess) Run() {
	p.runFunc()
}

func (p *stopFuncProcess) Stop() {
	p.stopFunc()
}
