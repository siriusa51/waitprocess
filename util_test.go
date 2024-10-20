package waitprocess

import (
	"sync/atomic"
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
