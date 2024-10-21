package waitprocess

import (
	"context"
	"unsafe"
)

type procstat struct {
	ctx      context.Context
	cancel   context.CancelFunc
	name     string
	proc     Process
	panicked unsafe.Pointer
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

func (p *procstat) getPanicked() unsafe.Pointer {
	return p.panicked
}

func (p *procstat) run() {
	defer func() {
		if r := recover(); r != nil {
			p.panicked = unsafe.Pointer(&r)
		}
	}()

	p.proc.Run()
}

func (p *procstat) stop() {
	defer p.cancel()
	p.proc.Stop()
}
