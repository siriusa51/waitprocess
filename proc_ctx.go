package waitprocess

import "context"

type ctxProcess struct {
	ctx context.Context
	run func(context.Context) error
}

// RunWithCtx creates a process that runs with a context
func RunWithCtx(run func(context.Context) error) Process {
	return &ctxProcess{run: run}
}

func (p *ctxProcess) SetContext(ctx context.Context) {
	p.ctx = ctx
}

func (p *ctxProcess) Run() error {
	return p.run(p.ctx)
}

func (p *ctxProcess) Stop() {
	// do nothing
}
