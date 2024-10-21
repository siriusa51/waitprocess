package waitprocess

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestRunWithStopFunc(t *testing.T) {
	t.Run("case-single-proc", func(t *testing.T) {
		wp := NewWaitProcess()

		stat := &teststate{}

		ch := make(chan struct{})
		wp.RegisterProcess("test", RunWithStopFunc(
			func() error {
				<-ch
				stat.add()
				return nil
			},
			func() {
				close(ch)
				stat.add()
			},
		))

		wp.Start()
		err := wp.Shutdown()
		assert.Nil(t, err, "shutdown should be success")

		assert.Equal(t, 2, stat.getstate(), "state should be 1")
	})

	t.Run("case-multi-process", func(t *testing.T) {
		wp := NewWaitProcess()

		stat1 := &teststate{}
		ch1 := make(chan struct{})

		stat2 := &teststate{}
		ch2 := make(chan struct{})

		wp.RegisterProcess("test1", RunWithStopFunc(
			func() error {
				<-ch1
				stat1.add()
				return nil
			},
			func() {
				close(ch1)
				stat1.add()
			},
		)).RegisterProcess("test2", RunWithStopFunc(
			func() error {
				<-ch2
				stat2.add()
				return nil
			},
			func() {
				close(ch2)
				stat2.add()
			},
		))

		wp.Start()
		time.Sleep(time.Second)
		err := wp.Shutdown()
		assert.Nil(t, err, "shutdown should be success")

		assert.Equal(t, 2, stat1.getstate(), "state should be 1")
		assert.Equal(t, 2, stat2.getstate(), "state should be 1")
	})

	// test stop one process
	t.Run("case-multi-and-stop-one", func(t *testing.T) {
		wp := NewWaitProcess()

		stat1 := &teststate{}
		stat2 := &teststate{}
		ch := make(chan struct{})
		wp.RegisterProcess("noloop", RunWithStopFunc(
			func() error {
				time.Sleep(time.Second * 1)
				stat1.add()
				return nil
			},
			func() {
				stat1.add()
			},
		)).RegisterProcess("loop", RunWithStopFunc(
			func() error {
				<-ch
				stat2.add()
				return nil
			},
			func() {
				close(ch)
				stat2.add()
			},
		))

		err := wp.Run()
		assert.Nil(t, err, "run should be success")

		assert.Equal(t, 2, stat1.getstate(), "state should be 1")
		assert.Equal(t, 2, stat2.getstate(), "state should be 1")
	})

	// test error process
	t.Run("case-error-process", func(t *testing.T) {
		wp := NewWaitProcess()
		wp.RegisterProcess("loop", withTestprocess())
		wp.RegisterProcess("error", RunWithStopFunc(
			func() error {
				return assert.AnError
			},
			func() {
			},
		))

		err := wp.Run()
		assert.NotNil(t, err, "run should be error")
	})
}
