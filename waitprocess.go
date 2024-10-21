package waitprocess

import (
	"context"
	"fmt"
	"github.com/sirupsen/logrus"
	"os"
	"os/signal"
	"sync"
	"sync/atomic"
	"time"
	"unsafe"
)

const (
	stateReady = iota
	stateStarted
)

var WaitTimeout = fmt.Errorf("Wait timeout")

// IsWaitTimeout returns true if the error is a WaitTimeout error
func IsWaitTimeout(err error) bool {
	return err == WaitTimeout
}

type WaitProcess struct {
	ctx        context.Context
	cancel     context.CancelFunc
	state      int
	lock       sync.Mutex
	log        *logrus.Entry
	signalChan chan os.Signal
	procs      []*procstat
	timer      *time.Timer
	stopChan   chan struct{}
	panicked   unsafe.Pointer
	error      unsafe.Pointer
}

// NewWaitProcess creates a new waitprocess
func NewWaitProcess(opts ...WaitProcessOption) *WaitProcess {
	opt := newWaitProcessOption(opts...)

	ctx, cancel := context.WithCancel(opt.ctx)
	return &WaitProcess{
		timer:      opt.timer,
		ctx:        ctx,
		cancel:     cancel,
		log:        opt.log,
		signalChan: make(chan os.Signal, 1),
		state:      stateReady,
		stopChan:   make(chan struct{}),
	}
}

func (wp *WaitProcess) ProcessCount() int {
	return len(wp.procs)
}

// RegisterProcess registers processes to be run by the waitprocess
func (wp *WaitProcess) RegisterProcess(name string, procs Process) *WaitProcess {
	wp.lock.Lock()
	defer wp.lock.Unlock()

	if wp.state != stateReady {
		wp.log.Panic("Cannot call RegisterProcess() after WaitProcess has already started")
	}

	wp.procs = append(wp.procs, newProcstat(name, procs))
	return wp
}

// RegisterSignal registers signals to be caught by the waitprocess
func (wp *WaitProcess) RegisterSignal(sigs ...os.Signal) *WaitProcess {
	wp.lock.Lock()
	defer wp.lock.Unlock()

	if wp.state != stateReady {
		wp.log.Panic("Cannot call RegisterSignal() after WaitProcess has already started")
	}

	signal.Notify(wp.signalChan, sigs...)
	return wp
}

// Start starts the waitprocess
func (wp *WaitProcess) Start() {
	wp.lock.Lock()
	defer wp.lock.Unlock()
	wp.start()
}

// Run starts the waitprocess and waits for it to stop
func (wp *WaitProcess) Run() error {
	wp.lock.Lock()
	wp.lock.Unlock()
	wp.start()
	return wp.wait()
}

// Stop stops the waitprocess
func (wp *WaitProcess) Stop() {
	wp.lock.Lock()
	defer wp.lock.Unlock()
	wp.stop()
}

func (wp *WaitProcess) Stopped() bool {
	select {
	case <-wp.stopChan:
		return true
	default:
		return false
	}
}

// Wait waits for the waitprocess to stop
func (wp *WaitProcess) Wait(timeout ...time.Duration) error {
	wp.lock.Lock()
	defer wp.lock.Unlock()
	return wp.wait(timeout...)
}

// Shutdown stops the waitprocess and waits for it to stop
func (wp *WaitProcess) Shutdown(timeout ...time.Duration) error {
	wp.lock.Lock()
	defer wp.lock.Unlock()
	wp.stop()
	return wp.wait(timeout...)
}

// Error returns the error of the waitprocess, only return the first error
func (wp *WaitProcess) Error() error {
	if wp.state != stateStarted {
		wp.log.Panic("Cannot call Error() before WaitProcess has started")
	}

	if err := atomic.LoadPointer(&wp.error); err != nil {
		return *(*error)(err)
	}
	return nil
}

func (wp *WaitProcess) start() {
	if wp.state != stateReady {
		wp.log.Panic("Cannot call Start() after WaitProcess has already started")
	}

	if len(wp.procs) == 0 {
		wp.log.Panic("Cannot start WaitProcess without any processes")
	}

	wg := sync.WaitGroup{}
	wg.Add(len(wp.procs))

	for i := range wp.procs {
		proc := wp.procs[i]
		log := wp.log.WithField("proc", proc)
		log.Debug("Starting process")
		proc.setContext(wp.ctx)

		go func() {
			defer func() {
				if panicked := proc.getPanicked(); panicked != nil {
					atomic.CompareAndSwapPointer(&wp.panicked, nil, panicked)
					log.WithField("panic", panicked).Error("Process panicked")
				}
				wg.Done()
				wp.cancel()
			}()

			if err := proc.run(); err != nil {
				atomic.CompareAndSwapPointer(&wp.error, nil, unsafe.Pointer(&err))
				log.WithField("error", err).Error("Process error")
			}

			log.Debug("Process stopped")
		}()
	}

	go func() {
		var timer <-chan time.Time

		if wp.timer != nil {
			timer = wp.timer.C
		} else {
			tch := make(chan time.Time)
			defer close(tch)
			timer = tch
		}

		select {
		case <-wp.signalChan:
			wp.log.Debug("Received signal, stopping WaitProcess")
		case <-wp.ctx.Done():
			wp.log.Debug("Context done, stopping WaitProcess")
		case <-timer:
			wp.log.Debug("Timer done, stopping WaitProcess")
		}

		for i := range wp.procs {
			proc := wp.procs[i]
			wp.log.WithField("proc", proc).Debug("Stopping process")
			proc.stop()
		}

		wg.Wait()
		close(wp.stopChan)
	}()

	wp.state = stateStarted
	wp.log.Info("WaitProcess started")
}

func (wp *WaitProcess) stop() {
	if wp.state != stateStarted {
		wp.log.Panic("Cannot call Stop() before WaitProcess has started")
	}

	wp.cancel()
}

func (wp *WaitProcess) wait(timeout ...time.Duration) error {
	if wp.state != stateStarted {
		wp.log.Panic("Cannot call Wait() before WaitProcess has started")
	}

	if len(timeout) > 0 {
		select {
		case <-wp.stopChan:
		case <-time.After(timeout[0]):
			return WaitTimeout
		}
	} else {
		<-wp.stopChan
	}

	if panicked := atomic.LoadPointer(&wp.panicked); panicked != nil {
		panic(*(*any)(panicked))
	}

	return wp.Error()
}
