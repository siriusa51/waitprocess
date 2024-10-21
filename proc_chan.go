package waitprocess

import "context"

type chanProcess struct {
	ctx context.Context
	run func(<-chan struct{}) error
}

// RunWithChan creates a process that runs a function with a channel to stop the process
func RunWithChan(run func(<-chan struct{}) error) Process {
	return &chanProcess{run: run}
}

func (p *chanProcess) SetContext(ctx context.Context) {
	p.ctx = ctx
}

func (p *chanProcess) Run() error {
	return p.run(p.ctx.Done())
}

func (p *chanProcess) Stop() {
	// do nothing
}
