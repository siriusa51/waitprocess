package waitprocess

import (
	"context"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"sync/atomic"
	"syscall"
	"testing"
	"time"
)

type testprocess struct {
	ch        chan struct{}
	runCount  int32
	stopCount int32
	stopped   int32
}

func withTestprocess() *testprocess {
	return &testprocess{
		ch: make(chan struct{}),
	}
}

func (tp *testprocess) getRunCount() int {
	return int(atomic.LoadInt32(&tp.runCount))
}

func (tp *testprocess) getStopCount() int {
	return int(atomic.LoadInt32(&tp.stopCount))
}

func (tp *testprocess) Run() {
	atomic.AddInt32(&tp.runCount, 1)
	<-tp.ch
}

func (tp *testprocess) Stop() {
	atomic.AddInt32(&tp.stopCount, 1)
	if ok := atomic.CompareAndSwapInt32(&tp.stopped, 0, 1); ok {
		close(tp.ch)
	}
}

func (tp *testprocess) SetContext(_ context.Context) {
}

func TestRegisterProcess(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		wp := NewWaitProcess()
		wp.RegisterProcess("test1", withTestprocess())
		assert.Equal(t, 1, wp.ProcessCount(), "process count should be 1")

		wp.RegisterProcess("test2", withTestprocess())
		assert.Equal(t, 2, wp.ProcessCount(), "process count should be 2")
	})

	t.Run("wp-started-and-register-new-process", func(t *testing.T) {
		wp := NewWaitProcess()

		wp.RegisterProcess("test1", withTestprocess())

		wp.Start()

		assert.Panics(t, func() {
			wp.RegisterProcess("test2", withTestprocess())
		})
	})

	t.Run("wp-stopped-and-register-new-process", func(t *testing.T) {
		wp := NewWaitProcess()

		wp.RegisterProcess("test1", withTestprocess())

		wp.Start()
		wp.Shutdown()

		assert.Panics(t, func() {
			wp.RegisterProcess("test2", withTestprocess())
		})
	})
}

func TestRegisterSignal(t *testing.T) {
	t.Run("signal", func(t *testing.T) {
		wp := NewWaitProcess()
		tp := withTestprocess()
		wp.RegisterProcess("test", tp)
		wp.RegisterSignal(syscall.SIGUSR1)

		wp.Start()
		// mock signal
		syscall.Kill(syscall.Getpid(), syscall.SIGUSR1)
		wp.Wait()

		assert.Equal(t, 1, tp.getRunCount(), "run count should be 1")
		assert.Equal(t, 1, tp.getStopCount(), "stop count should be 1")
	})

	t.Run("signal-unregister", func(t *testing.T) {
		wp := NewWaitProcess()
		tp := withTestprocess()
		wp.RegisterProcess("test", tp)
		wp.RegisterSignal(syscall.SIGUSR1)

		wp.Start()
		defer wp.Shutdown()

		syscall.Kill(syscall.Getpid(), syscall.SIGUSR2)
		time.Sleep(time.Second)

		assert.False(t, wp.Stopped(), "wp should not be stopped")
	})

	t.Run("signal-after-started", func(t *testing.T) {
		wp := NewWaitProcess()
		wp.RegisterProcess("test", withTestprocess())
		wp.Start()
		defer wp.Shutdown()

		assert.Panics(t, func() {
			wp.RegisterSignal(syscall.SIGUSR1)
		})
	})
}

func TestStart(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		wp := NewWaitProcess()
		tp1 := withTestprocess()
		tp2 := withTestprocess()
		wp.RegisterProcess("test1", tp1).RegisterProcess("test2", tp2)
		wp.Start()
		defer wp.Shutdown()

		assert.Equal(t, 1, tp1.getRunCount(), "run count should be 1")
		assert.Equal(t, 0, tp1.getStopCount(), "stop count should be 0")
		assert.Equal(t, 1, tp2.getRunCount(), "run count should be 1")
		assert.Equal(t, 0, tp2.getStopCount(), "stop count should be 0")
	})

	t.Run("start-after-stopped", func(t *testing.T) {
		wp := NewWaitProcess()
		wp.RegisterProcess("test", withTestprocess())
		wp.Start()
		wp.Shutdown()

		assert.Panics(t, func() {
			wp.Start()
		})
	})

	t.Run("empty-process", func(t *testing.T) {
		wp := NewWaitProcess()
		assert.Panics(t, func() {
			wp.Start()
		})
	})

}

func TestStop(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		wp := NewWaitProcess()
		tp1 := withTestprocess()
		tp2 := withTestprocess()
		wp.RegisterProcess("test1", tp1).RegisterProcess("test2", tp2)
		wp.Start()
		wp.Shutdown()

		assert.Equal(t, 1, tp1.getRunCount(), "run count should be 1")
		assert.Equal(t, 1, tp1.getStopCount(), "stop count should be 1")
		assert.Equal(t, 1, tp2.getRunCount(), "run count should be 1")
		assert.Equal(t, 1, tp2.getStopCount(), "stop count should be 1")
	})

	t.Run("stop-before-start", func(t *testing.T) {
		wp := NewWaitProcess()

		assert.Panics(t, func() {
			wp.Stop()
		})
	})
}

func TestRun(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		wp := NewWaitProcess()
		tp1 := withTestprocess()
		tp2 := withTestprocess()

		wp.RegisterProcess("test1", tp1).RegisterProcess("test2", tp2)
		go wp.Run()

		time.Sleep(time.Second)

		assert.Equal(t, 1, tp1.getRunCount(), "run count should be 1")
		assert.Equal(t, 0, tp1.getStopCount(), "stop count should be 0")

		assert.Equal(t, 1, tp2.getRunCount(), "run count should be 1")
		assert.Equal(t, 0, tp2.getStopCount(), "stop count should be 0")

		wp.Shutdown()

		assert.Equal(t, 1, tp1.getRunCount(), "run count should be 1")
		assert.Equal(t, 1, tp1.getStopCount(), "stop count should be 1")

		assert.Equal(t, 1, tp2.getRunCount(), "run count should be 1")
		assert.Equal(t, 1, tp2.getStopCount(), "stop count should be 1")
	})

	t.Run("stop-process", func(t *testing.T) {
		wp := NewWaitProcess()
		tp := withTestprocess()

		wp.RegisterProcess("app1", tp)
		wp.RegisterProcess("app2", withTestprocess())

		go wp.Run()
		time.Sleep(time.Second)
		tp.Stop()

		wp.Wait()
		assert.True(t, wp.Stopped(), "wp should be stopped")
		assert.Equal(t, 1, tp.getRunCount(), "run count should be 1")
		// call twice
		assert.Equal(t, 2, tp.getStopCount(), "stop count should be 1")
	})
	t.Run("run-after-stopped", func(t *testing.T) {
		wp := NewWaitProcess()
		wp.RegisterProcess("test", withTestprocess())
		wp.Start()
		wp.Shutdown()

		assert.Panics(t, func() {
			wp.Run()
		})
	})
}

func TestWait(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		wp := NewWaitProcess()
		tp1 := withTestprocess()
		tp2 := withTestprocess()

		wp.RegisterProcess("test1", tp1).RegisterProcess("test2", tp2)
		wp.Start()

		time.Sleep(time.Second)

		assert.Equal(t, 1, tp1.getRunCount(), "run count should be 1")
		assert.Equal(t, 0, tp1.getStopCount(), "stop count should be 0")

		assert.Equal(t, 1, tp2.getRunCount(), "run count should be 1")
		assert.Equal(t, 0, tp2.getStopCount(), "stop count should be 0")

		wp.Stop()
		wp.Wait()

		assert.Equal(t, 1, tp1.getRunCount(), "run count should be 1")
		assert.Equal(t, 1, tp1.getStopCount(), "stop count should be 1")

		assert.Equal(t, 1, tp2.getRunCount(), "run count should be 1")
		assert.Equal(t, 1, tp2.getStopCount(), "stop count should be 1")
	})

	t.Run("timeout", func(t *testing.T) {
		wp := NewWaitProcess()
		tp1 := withTestprocess()
		tp2 := withTestprocess()

		wp.RegisterProcess("test1", tp1).RegisterProcess("test2", tp2)
		wp.start()

		time.Sleep(time.Second)

		assert.Equal(t, 1, tp1.getRunCount(), "run count should be 1")
		assert.Equal(t, 0, tp1.getStopCount(), "stop count should be 0")

		assert.Equal(t, 1, tp2.getRunCount(), "run count should be 1")
		assert.Equal(t, 0, tp2.getStopCount(), "stop count should be 0")

		wp.Wait(time.Second)

		assert.Equal(t, 1, tp1.getRunCount(), "run count should be 1")
		assert.Equal(t, 0, tp1.getStopCount(), "stop count should be 0")

		assert.Equal(t, 1, tp2.getRunCount(), "run count should be 1")
		assert.Equal(t, 0, tp2.getStopCount(), "stop count should be 0")

		wp.Stop()
		wp.Wait()

		assert.Equal(t, 1, tp1.getRunCount(), "run count should be 1")
		assert.Equal(t, 1, tp1.getStopCount(), "stop count should be 1")

		assert.Equal(t, 1, tp2.getRunCount(), "run count should be 1")
		assert.Equal(t, 1, tp2.getStopCount(), "stop count should be 1")
	})

	t.Run("wait-before-start", func(t *testing.T) {
		wp := NewWaitProcess()

		assert.Panics(t, func() {
			wp.Wait()
		})
	})

	t.Run("wait-stopped", func(t *testing.T) {
		wp := NewWaitProcess()
		wp.RegisterProcess("test", withTestprocess())
		wp.Start()
		wp.Shutdown()
		wp.Wait(time.Second)
		assert.True(t, wp.Stopped(), "wp should be stopped")
	})
}

func TestShutdown(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		wp := NewWaitProcess()
		tp1 := withTestprocess()
		tp2 := withTestprocess()

		wp.RegisterProcess("test1", tp1).RegisterProcess("test2", tp2)
		wp.Start()

		time.Sleep(time.Second)

		assert.Equal(t, 1, tp1.getRunCount(), "run count should be 1")
		assert.Equal(t, 0, tp1.getStopCount(), "stop count should be 0")

		assert.Equal(t, 1, tp2.getRunCount(), "run count should be 1")
		assert.Equal(t, 0, tp2.getStopCount(), "stop count should be 0")

		wp.Shutdown()

		assert.Equal(t, 1, tp1.getRunCount(), "run count should be 1")
		assert.Equal(t, 1, tp1.getStopCount(), "stop count should be 1")

		assert.Equal(t, 1, tp2.getRunCount(), "run count should be 1")
		assert.Equal(t, 1, tp2.getStopCount(), "stop count should be 1")
	})

	t.Run("shutdown-before-start", func(t *testing.T) {
		wp := NewWaitProcess()

		assert.Panics(t, func() {
			wp.Shutdown()
		})
	})
}

func TestStopped(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		wp := NewWaitProcess()
		tp1 := withTestprocess()
		tp2 := withTestprocess()

		assert.False(t, wp.Stopped(), "wp should not be stopped")
		wp.RegisterProcess("test1", tp1).RegisterProcess("test2", tp2)
		wp.Start()

		assert.False(t, wp.Stopped(), "wp should not be stopped")

		wp.Shutdown()

		assert.True(t, wp.Stopped(), "wp should be stopped")
	})
}

func TestNewProcess(t *testing.T) {
	t.Run("timer", func(t *testing.T) {
		wp := NewWaitProcess(WithTimer(time.Second))
		wp.RegisterProcess("test", withTestprocess())

		wp.Run()
		assert.True(t, wp.Stopped(), "wp should be stopped")
	})

	t.Run("context", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		wp := NewWaitProcess(WithContext(ctx))
		tp := withTestprocess()
		wp.RegisterProcess("test", tp)

		wp.Start()
		cancel()

		wp.Wait()
		assert.True(t, wp.Stopped(), "wp should be stopped")
		assert.Equal(t, 1, tp.getRunCount(), "run count should be 1")
		assert.Equal(t, 1, tp.getStopCount(), "stop count should be 1")
	})

	t.Run("log", func(t *testing.T) {
		log := logrus.WithField("pkg", "waitprocess")
		wp := NewWaitProcess(WithLog(log))
		assert.Equal(t, log, wp.log)
	})
}
