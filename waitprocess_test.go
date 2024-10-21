package waitprocess

import (
	"context"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"syscall"
	"testing"
	"time"
)

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
		err := wp.Shutdown()
		assert.Nil(t, err)

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
		err := wp.Wait()
		assert.Nil(t, err)

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

func TestPanic(t *testing.T) {
	t.Run("panic", func(t *testing.T) {
		wp := NewWaitProcess()
		tp := withTestprocess()
		wp.RegisterProcess("test", tp)
		wp.RegisterProcess("panic", RunWithCtx(func(ctx context.Context) error {
			panic("panic")
			return nil
		}))

		assert.Panics(t, func() {
			wp.Run()
		})

		assert.True(t, wp.Stopped(), "wp should be stopped")
		assert.Equal(t, 1, tp.getRunCount(), "run count should be 1")
		assert.Equal(t, 1, tp.getStopCount(), "stop count should be 1")
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

		time.Sleep(time.Second)

		assert.Equal(t, 1, tp1.getRunCount(), "run count should be 1")
		assert.Equal(t, 0, tp1.getStopCount(), "stop count should be 0")
		assert.Equal(t, 1, tp2.getRunCount(), "run count should be 1")
		assert.Equal(t, 0, tp2.getStopCount(), "stop count should be 0")
	})

	t.Run("start-after-stopped", func(t *testing.T) {
		wp := NewWaitProcess()
		wp.RegisterProcess("test", withTestprocess())
		wp.Start()
		err := wp.Shutdown()
		assert.Nil(t, err)

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
		err := wp.Shutdown()
		assert.Nil(t, err)

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
		go func() {
			err := wp.Run()
			assert.Nil(t, err)
		}()

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

		go func() {
			err := wp.Run()
			assert.Nil(t, err)
		}()

		time.Sleep(time.Second)
		tp.Stop()

		err := wp.Wait()
		assert.Nil(t, err)

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

	t.Run("run-return-error", func(t *testing.T) {
		wp := NewWaitProcess()
		wp.RegisterProcess("loop", withTestprocess())
		wp.RegisterProcess("error", withErrprocess(assert.AnError))

		err := wp.Run()
		assert.ErrorAs(t, err, &assert.AnError)
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
		err := wp.Wait()
		assert.Nil(t, err)

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

		err := wp.Wait(time.Second)
		assert.ErrorAs(t, err, &WaitTimeout)

		assert.Equal(t, 1, tp1.getRunCount(), "run count should be 1")
		assert.Equal(t, 0, tp1.getStopCount(), "stop count should be 0")

		assert.Equal(t, 1, tp2.getRunCount(), "run count should be 1")
		assert.Equal(t, 0, tp2.getStopCount(), "stop count should be 0")

		wp.Stop()
		err = wp.Wait()
		assert.Nil(t, err)

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
		err := wp.Shutdown()
		assert.Nil(t, err)
		err = wp.Wait(time.Second)
		assert.Nil(t, err)
		assert.True(t, wp.Stopped(), "wp should be stopped")
	})

	t.Run("wait-error", func(t *testing.T) {
		wp := NewWaitProcess()
		wp.RegisterProcess("loop", withTestprocess())
		wp.RegisterProcess("error", withErrprocess(assert.AnError))

		wp.Start()
		err := wp.Wait()
		assert.ErrorAs(t, err, &assert.AnError)
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

		err := wp.Shutdown()
		assert.Nil(t, err)

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

	t.Run("shutdown-stopped", func(t *testing.T) {
		wp := NewWaitProcess()
		wp.RegisterProcess("test", withTestprocess())
		wp.Start()
		wp.Stop()
		err := wp.Shutdown()
		assert.Nil(t, err)
		assert.True(t, wp.Stopped(), "wp should be stopped")
	})

	t.Run("shutown-error", func(t *testing.T) {
		wp := NewWaitProcess()
		wp.RegisterProcess("loop", withTestprocess())
		wp.RegisterProcess("error", withErrprocess(assert.AnError))

		wp.Start()
		err := wp.Shutdown()
		assert.ErrorAs(t, err, &assert.AnError)
	})

	t.Run("shutdown-timeout", func(t *testing.T) {
		wp := NewWaitProcess()
		wp.RegisterProcess("loop", withTestprocess())
		wp.RegisterProcess("sleep", withSleepprocess(time.Second*2))
		wp.Start()
		defer wp.Shutdown()

		err := wp.Shutdown(time.Second)
		assert.ErrorAs(t, err, &WaitTimeout)
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

func TestIsWaitTimeout(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		assert.True(t, IsWaitTimeout(WaitTimeout))
		assert.False(t, IsWaitTimeout(assert.AnError))
	})
}

func TestError(t *testing.T) {
	t.Run("call-error-before-start", func(t *testing.T) {
		wp := NewWaitProcess()
		assert.Panics(t, func() {
			wp.Error()
		})
	})
}
