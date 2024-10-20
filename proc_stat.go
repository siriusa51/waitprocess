package waitprocess

import (
	"context"
)

type procstat struct {
	ctx    context.Context
	cancel context.CancelFunc
	name   string
	proc   Process
}

func newProcstat(name string, proc Process) *procstat {
	return &procstat{
		name: name,
		proc: proc,
	}
}

func (p *procstat) setContext(ctx context.Context) {
	p.ctx, p.cancel = context.WithCancel(ctx)
	p.proc.SetContext(p.ctx)
}

func (p *procstat) run() {
	p.proc.Run()
}

func (p *procstat) stop() {
	defer p.cancel()
	p.proc.Stop()
}
