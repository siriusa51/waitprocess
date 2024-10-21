package waitprocess

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestRunWithChan(t *testing.T) {
	t.Run("case-single-proc", func(t *testing.T) {
		wp := NewWaitProcess()

		stat := &teststate{}

		wp.RegisterProcess("test", RunWithChan(func(cancel <-chan struct{}) error {
			<-cancel
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

		wp.RegisterProcess("test1", RunWithChan(func(cancel <-chan struct{}) error {
			<-cancel
			stat1.add()
			return nil
		})).RegisterProcess("test2", RunWithChan(func(cancel <-chan struct{}) error {
			<-cancel
			stat2.add()
			return nil
		}))

		wp.Start()
		time.Sleep(time.Second)
		err := wp.Shutdown()
		assert.Nil(t, err, "error should be nil")

		assert.Equal(t, 1, stat1.getstate(), "state should be 1")
		assert.Equal(t, 1, stat2.getstate(), "state should be 1")
	})

	// test stop one process
	t.Run("case-multi-and-stop-one", func(t *testing.T) {
		wp := NewWaitProcess()

		stat := &teststate{}

		wp.RegisterProcess("noloop", RunWithChan(func(cancel <-chan struct{}) error {
			time.Sleep(time.Second)
			return nil
		})).RegisterProcess("loop", RunWithChan(func(cancel <-chan struct{}) error {
			<-cancel
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
		wp.RegisterProcess("error", RunWithChan(func(cancel <-chan struct{}) error {
			stat.add()
			return assert.AnError
		}))
		wp.RegisterProcess("loop", withTestprocess())

		err := wp.Run()
		assert.NotNil(t, err, "error should be not nil")
		assert.Equal(t, 1, stat.getstate(), "state should be 1")
	})
}
