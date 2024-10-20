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
			func() {
				<-ch
				stat.add()
			},
			func() {
				close(ch)
				stat.add()
			},
		))

		wp.Start()
		wp.Shutdown()

		assert.Equal(t, 2, stat.getstate(), "state should be 1")
	})

	t.Run("case-multi-process", func(t *testing.T) {
		wp := NewWaitProcess()

		stat1 := &teststate{}
		ch1 := make(chan struct{})

		stat2 := &teststate{}
		ch2 := make(chan struct{})

		wp.RegisterProcess("test1", RunWithStopFunc(
			func() {
				<-ch1
				stat1.add()
			},
			func() {
				close(ch1)
				stat1.add()
			},
		)).RegisterProcess("test2", RunWithStopFunc(
			func() {
				<-ch2
				stat2.add()
			},
			func() {
				close(ch2)
				stat2.add()
			},
		))

		wp.Start()
		time.Sleep(time.Second)
		wp.Shutdown()

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
			func() {
				time.Sleep(time.Second * 1)
				stat1.add()
			},
			func() {
				stat1.add()
			},
		)).RegisterProcess("loop", RunWithStopFunc(
			func() {
				<-ch
				stat2.add()
			},
			func() {
				close(ch)
				stat2.add()
			},
		))

		wp.Run()

		assert.Equal(t, 2, stat1.getstate(), "state should be 1")
		assert.Equal(t, 2, stat2.getstate(), "state should be 1")
	})
}
