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

		wp.RegisterProcess("test", RunWithChan(func(cancel <-chan struct{}) {
			<-cancel
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

		wp.RegisterProcess("test1", RunWithChan(func(cancel <-chan struct{}) {
			<-cancel
			stat1.add()
		})).RegisterProcess("test2", RunWithChan(func(cancel <-chan struct{}) {
			<-cancel
			stat2.add()
		}))

		wp.Start()
		time.Sleep(time.Second)
		wp.Shutdown()

		assert.Equal(t, 1, stat1.getstate(), "state should be 1")
		assert.Equal(t, 1, stat2.getstate(), "state should be 1")
	})

	// test stop one process
	t.Run("case-multi-and-stop-one", func(t *testing.T) {
		wp := NewWaitProcess()

		stat := &teststate{}

		wp.RegisterProcess("noloop", RunWithChan(func(cancel <-chan struct{}) {
			time.Sleep(time.Second)
		})).RegisterProcess("loop", RunWithChan(func(cancel <-chan struct{}) {
			<-cancel
			stat.add()
		}))

		wp.Run()

		assert.Equal(t, 1, stat.getstate(), "state should be 1")
	})
}
