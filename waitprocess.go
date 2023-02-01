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
)

const (
	defaultTimeout = 10

	stateReady   = uint32(0)
	stateRunning = uint32(1)
	stateStop    = uint32(2)
)

type eventType int

const (
	typeRegisterProcess = iota
	typeRegisterSignal
	typeStart
	typeStop
)

type event struct {
	types  eventType
	option interface{}
	errCh  chan error
}

type WaitProcess struct {
	// waitprocess state
	state uint32
	// set timeout when stopping
	timeout int
	// singleton pattern initialisation event handlers
	once   sync.Once
	stopWG sync.WaitGroup
	// receive process stop signal
	stopCh chan struct{}
	// registered processes
	processes Processes
	// registered signal
	signals []os.Signal
	// event handlers channel
	events chan event
	// notify process register by process.ServeForverWithCtx
	cancelFunc context.CancelFunc
	// notify process register by process.ServeForverWithChan
	notifyStopChan chan struct{}
	finishCh       chan struct{}
}

var (
	once       sync.Once
	defaultObj *WaitProcess
)

func Default() *WaitProcess {
	once.Do(func() {
		defaultObj = New()
	})

	return defaultObj
}

type Process struct {
	ServeForver func()
	StopForCtx  func(ctx context.Context)
	StopForChan func(<-chan struct{})

	ServeForverWithCtx func(ctx context.Context)
	ServeForverWitChan func(<-chan struct{})
}

type Processes []Process

func (fps Processes) run(wp *WaitProcess) {
	ctx, fun := context.WithCancel(context.Background())
	wp.cancelFunc = fun

	for _, sp := range fps {
		go func(sp Process) {
			if sp.ServeForverWithCtx != nil {
				sp.ServeForverWithCtx(ctx)
			} else if sp.ServeForver != nil {
				sp.ServeForver()
			} else if sp.ServeForverWitChan != nil {
				sp.ServeForverWitChan(wp.notifyStopChan)
			}

			_ = wp.Stop()
		}(sp)
	}
}

func (fps Processes) stop(wp *WaitProcess, ctx context.Context) {
	for _, sp := range fps {
		go func(sp Process) {
			if sp.StopForCtx != nil {
				sp.StopForCtx(ctx)
			}

			if sp.StopForChan != nil {
				sp.StopForChan(wp.notifyStopChan)
			}

			wp.stopWG.Done()
		}(sp)
	}
}

func New() *WaitProcess {
	wp := &WaitProcess{
		state:          stateReady,
		stopCh:         make(chan struct{}),
		events:         make(chan event),
		finishCh:       make(chan struct{}),
		notifyStopChan: make(chan struct{}),
	}

	return wp
}

func (wp *WaitProcess) getState() uint32 {
	return atomic.LoadUint32(&wp.state)
}

func (wp *WaitProcess) setState(s uint32) {
	atomic.StoreUint32(&wp.state, s)
}

func (wp *WaitProcess) eventLoop() {
	for {
		select {
		case e := <-wp.events:
			wp.processEvent(e)
		case <-wp.stopCh:
			// stop signal
			wp.doStop()
			break
		}
	}
}

func (wp *WaitProcess) processEvent(e event) {
	switch e.types {
	case typeRegisterProcess:
		e.errCh <- wp.registerProcess(e.option.(Processes)...)
	case typeRegisterSignal:
		e.errCh <- wp.registerSignal(e.option.([]os.Signal)...)
	case typeStart:
		e.errCh <- wp.start(e.option.([]int)...)
	case typeStop:
		e.errCh <- wp.stop()
	default:
		e.errCh <- fmt.Errorf("unknown event type")
	}
}

func (wp *WaitProcess) pushEvent(etype eventType, option interface{}) error {
	wp.once.Do(func() {
		go wp.eventLoop()
	})

	e := event{
		types:  etype,
		option: option,
		errCh:  make(chan error),
	}

	select {
	case wp.events <- e:
	case <-wp.finishCh:
		if etype == typeStop {
			return nil
		}

		return fmt.Errorf("waitprocess is stopped")
	}

	select {
	case err := <-e.errCh:
		return err
	case <-wp.finishCh:
		if etype == typeStop {
			return nil
		}

		return fmt.Errorf("waitprocess is stopped")
	}
}

// single thead
func (wp *WaitProcess) RegisterProcess(processes ...Process) error {
	return wp.pushEvent(typeRegisterProcess, Processes(processes))
}

// single thead
// add loop forerver process functions
func (wp *WaitProcess) registerProcess(processes ...Process) error {
	if wp.getState() != stateReady {
		return fmt.Errorf("waitprocess is started")
	}
	for _, p := range processes {
		if p.ServeForver != nil && p.StopForChan == nil && p.StopForCtx == nil {
			return fmt.Errorf("if you want to set Process.ServeForver, then you must also set Process.StopForChan or Process.StopForCtx")
		}
	}
	wp.processes = append(wp.processes, processes...)
	return nil
}

// single thead
func (wp *WaitProcess) RegisterSignal(sigs ...os.Signal) error {
	return wp.pushEvent(typeRegisterSignal, sigs)
}

// single thead
// add os signal
func (wp *WaitProcess) registerSignal(sigs ...os.Signal) error {
	if wp.getState() != stateReady {
		return fmt.Errorf("waitprocess is started")
	}
	if len(sigs) == 0 {
		return nil
	}

	wp.signals = append(wp.signals, sigs...)
	return nil
}

// single thead
func (wp *WaitProcess) Stop() error {
	return wp.pushEvent(typeStop, nil)
}

// single thead
func (wp *WaitProcess) stop() error {
	if wp.getState() == stateReady {
		return fmt.Errorf("waitprocess not started")
	}

	if wp.getState() == stateStop {
		return nil
	}

	// set state
	wp.setState(stateStop)

	// notify stop
	wp.stopCh <- struct{}{}
	return nil
}

// single thead
func (wp *WaitProcess) Start(timeout ...int) error {
	return wp.pushEvent(typeStart, timeout)
}

// single thead
func (wp *WaitProcess) start(timeout ...int) error {
	if wp.getState() != stateReady {
		return fmt.Errorf("waitprocess is started")
	}

	wp.setState(stateRunning)

	if len(timeout) > 1 {
		wp.timeout = timeout[0]
	}

	if len(wp.signals) != 0 {
		wp.processes = append(wp.processes, Process{
			ServeForver: func() {
				ch := make(chan os.Signal)
				signal.Notify(ch, wp.signals...)
				select {
				case sig := <-ch:
					logrus.WithFields(logrus.Fields{
						"signal": sig.String(),
					}).Info("wrap stop signal")
				}
			},
		})
	}

	wp.stopWG.Add(len(wp.processes))
	wp.stopCh = make(chan struct{}, len(wp.processes)+1)
	wp.processes.run(wp)

	return nil
}

func (wp *WaitProcess) doStop() {
	if wp.timeout <= 0 {
		wp.timeout = defaultTimeout
	}

	signal.Reset(wp.signals...)

	ctx, timeoutFunc := context.WithTimeout(context.Background(), time.Second*time.Duration(wp.timeout))
	defer timeoutFunc()

	close(wp.notifyStopChan)

	wp.cancelFunc()

	wp.processes.stop(wp, ctx)

	waitGroupCh := make(chan struct{})

	go func() {
		wp.stopWG.Wait()
		waitGroupCh <- struct{}{}
	}()

	select {
	case <-ctx.Done():
		// timeout
	case <-waitGroupCh:
		// stop function done
	}

	// notify wait
	close(wp.finishCh)
}

// wait process stop
func (wp *WaitProcess) Wait() {
	select {
	case <-wp.finishCh:
	}
}

// process will start and block to the end
func (wp *WaitProcess) Run(timeout ...int) error {
	if err := wp.Start(timeout...); err != nil {
		return err
	}
	wp.Wait()
	return nil
}

func (wp *WaitProcess) StopAndWait() error {
	if err := wp.Stop(); err != nil {
		return err
	}
	wp.Wait()
	return nil
}

func RegisterProcess(processes ...Process) error {
	return Default().RegisterProcess(processes...)
}

func RegisterSignal(sigs ...os.Signal) error {
	return Default().registerSignal(sigs...)
}

func Stop() error {
	return Default().Stop()
}

func Start(timeout ...int) error {
	return Default().Start(timeout...)
}

func Wait() {
	Default().Wait()
}

func StopAndWait() error {
	return Default().StopAndWait()
}

func Run() error {
	return Default().Run()
}
