package waitprocess

import (
	"context"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestRunWithCtx(t *testing.T) {
	t.Run("case-single-proc", func(t *testing.T) {
		wp := NewWaitProcess()

		stat := &teststate{}

		wp.RegisterProcess("test", RunWithCtx(func(ctx context.Context) {
			<-ctx.Done()
			stat.add()
		}))

		wp.Start()
		wp.Stop()
		wp.Wait()

		assert.Equal(t, 1, stat.getstate(), "state should be 1")
	})

	t.Run("case-multi-process", func(t *testing.T) {
		wp := NewWaitProcess()

		stat1 := &teststate{}
		stat2 := &teststate{}

		wp.RegisterProcess("test1", RunWithCtx(func(ctx context.Context) {
			stat1.add()
		})).RegisterProcess("test2", RunWithCtx(func(ctx context.Context) {
			<-ctx.Done()
			stat2.add()
		}))

		wp.Start()
		wp.Shutdown()

		assert.Equal(t, 1, stat1.getstate(), "state should be 1")
		assert.Equal(t, 1, stat2.getstate(), "state should be 1")
	})

	// test stop one process
	t.Run("case-multi-and-stop-one", func(t *testing.T) {
		wp := NewWaitProcess()

		stat := &teststate{}

		wp.RegisterProcess("noloop", RunWithCtx(func(ctx context.Context) {
			time.Sleep(time.Second * 1)
		})).RegisterProcess("loop", RunWithCtx(func(ctx context.Context) {
			<-ctx.Done()
			stat.add()
		}))

		wp.Run()

		assert.Equal(t, 1, stat.getstate(), "state should be 1")
	})
}
