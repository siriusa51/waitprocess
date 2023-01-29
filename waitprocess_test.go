package waitprocess

import (
	"context"
	"github.com/stretchr/testify/assert"
	"os"
	"sync/atomic"
	"syscall"
	"testing"
	"time"
)

func TestWaitProcess(t *testing.T) {
	obj := New()
	count := uint32(0)
	ch := make(chan struct{})
	_ = obj.RegisterProcess(
		Process{
			ServeForverWithCtx: func(ctx context.Context) {
				select {
				case <-ctx.Done():
					atomic.AddUint32(&count, 1)
				}
			},
		},
		Process{
			ServeForver: func() {
				select {
				case <-ch:
					atomic.AddUint32(&count, 1)
				}
			},
			Stop: func(ctx context.Context) {
				ch <- struct{}{}
			},
		},
	)

	_ = obj.Start(1)

	time.Sleep(time.Second * time.Duration(1))
	_ = obj.Stop()

	time.Sleep(time.Second * time.Duration(1))
	assert.Equal(t, uint32(2), count)
}

func TestSignal(t *testing.T) {
	obj := New()
	count := uint32(0)
	_ = obj.registerSignal(syscall.SIGUSR1)

	ch := make(chan struct{})

	_ = obj.RegisterProcess(
		Process{
			ServeForverWithCtx: func(ctx context.Context) {
				select {
				case <-ctx.Done():
					atomic.AddUint32(&count, 1)
				}
			},
		},
		Process{
			ServeForver: func() {
				select {
				case <-ch:
					atomic.AddUint32(&count, 1)
				}
			},
			Stop: func(ctx context.Context) {
				ch <- struct{}{}
			},
		},
	)

	_ = obj.Start(1)

	time.Sleep(time.Second * time.Duration(1))

	_ = syscall.Kill(os.Getpid(), syscall.SIGUSR1)

	time.Sleep(time.Second * time.Duration(1))
	assert.Equal(t, uint32(2), count)
}

func TestNoBlock(t *testing.T) {
	obj := New()
	count := uint32(0)

	_ = obj.RegisterProcess(
		Process{
			ServeForverWithCtx: func(ctx context.Context) {
				atomic.AddUint32(&count, 1)
				time.Sleep(time.Second * time.Duration(1))
			},
		},
		Process{
			ServeForverWithCtx: func(ctx context.Context) {
				select {
				case <-ctx.Done():
					atomic.AddUint32(&count, 1)
				}
			},
		},
	)

	_ = obj.Start(1)

	time.Sleep(time.Second * time.Duration(2))

	assert.Equal(t, uint32(2), count)
}

func TestTimeout(t *testing.T) {
	obj := New()
	count := uint32(0)

	_ = obj.RegisterProcess(
		Process{
			ServeForverWithCtx: func(ctx context.Context) {
				select {
				case <-ctx.Done():
					atomic.AddUint32(&count, 1)
				}
			},
		},
		Process{
			ServeForverWithCtx: func(ctx context.Context) {
				select {
				case <-ctx.Done():
					// timeout
					time.Sleep(time.Second * time.Duration(10))
					atomic.AddUint32(&count, 1)
				}
			},
		},
	)

	_ = obj.Start(1)

	time.Sleep(time.Second * time.Duration(1))
	_ = obj.Stop()
	time.Sleep(time.Second * time.Duration(1))
	assert.Equal(t, uint32(1), count)
}
