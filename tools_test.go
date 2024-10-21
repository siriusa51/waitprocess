package waitprocess

import (
	"context"
	"sync/atomic"
	"time"
)

type teststate struct {
	state int32
}

func (t *teststate) add() {
	atomic.AddInt32(&t.state, 1)
}

func (t *teststate) getstate() int {
	return int(atomic.LoadInt32(&t.state))
}

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

func (tp *testprocess) Run() error {
	atomic.AddInt32(&tp.runCount, 1)
	<-tp.ch
	return nil
}

func (tp *testprocess) Stop() {
	atomic.AddInt32(&tp.stopCount, 1)
	if ok := atomic.CompareAndSwapInt32(&tp.stopped, 0, 1); ok {
		close(tp.ch)
	}
}

func (tp *testprocess) SetContext(_ context.Context) {
}

type errprocess struct {
	err error
}

func withErrprocess(err error) *errprocess {
	return &errprocess{err}
}

func (ep *errprocess) Run() error {
	return ep.err
}

func (ep *errprocess) Stop() {
}

func (ep *errprocess) SetContext(_ context.Context) {
}

type sleepprocess struct {
	duration time.Duration
}

func withSleepprocess(duration time.Duration) *sleepprocess {
	return &sleepprocess{duration}
}

func (tp *sleepprocess) Run() error {
	time.Sleep(tp.duration)
	return nil
}

func (tp *sleepprocess) Stop() {
}

func (tp *sleepprocess) SetContext(_ context.Context) {
}
