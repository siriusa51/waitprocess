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

		wp.RegisterProcess("test", RunWithCtx(func(ctx context.Context) error {
			<-ctx.Done()
			stat.add()
			return nil
		}))

		wp.Start()
		wp.Stop()
		err := wp.Wait()
		assert.Nil(t, err, "error should be nil")

		assert.Equal(t, 1, stat.getstate(), "state should be 1")
	})

	t.Run("case-multi-process", func(t *testing.T) {
		wp := NewWaitProcess()

		stat1 := &teststate{}
		stat2 := &teststate{}

		wp.RegisterProcess("test1", RunWithCtx(func(ctx context.Context) error {
			stat1.add()
			return nil
		})).RegisterProcess("test2", RunWithCtx(func(ctx context.Context) error {
			<-ctx.Done()
			stat2.add()
			return nil
		}))

		wp.Start()
		err := wp.Shutdown()
		assert.Nil(t, err, "error should be nil")

		assert.Equal(t, 1, stat1.getstate(), "state should be 1")
		assert.Equal(t, 1, stat2.getstate(), "state should be 1")
	})

	// test stop one process
	t.Run("case-multi-and-stop-one", func(t *testing.T) {
		wp := NewWaitProcess()

		stat := &teststate{}

		wp.RegisterProcess("noloop", RunWithCtx(func(ctx context.Context) error {
			time.Sleep(time.Second * 1)
			return nil
		})).RegisterProcess("loop", RunWithCtx(func(ctx context.Context) error {
			<-ctx.Done()
			stat.add()
			return nil
		}))

		err := wp.Run()
		assert.Nil(t, err, "error should be nil")

		assert.Equal(t, 1, stat.getstate(), "state should be 1")
	})

	// test error process
	t.Run("case-error-process", func(t *testing.T) {
		wp := NewWaitProcess()

		stat := &teststate{}

		wp.RegisterProcess("error", RunWithCtx(func(ctx context.Context) error {
			stat.add()
			return assert.AnError
		}))

		wp.RegisterProcess("loop", withTestprocess())

		err := wp.Run()
		assert.NotNil(t, err, "error should not be nil")
		assert.Equal(t, 1, stat.getstate(), "state should be 1")
	})
}
