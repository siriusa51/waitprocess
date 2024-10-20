package waitprocess

import (
	"context"
	"github.com/sirupsen/logrus"
	"time"
)

type waitProcessOption struct {
	ctx   context.Context
	log   *logrus.Entry
	timer *time.Timer
}

type WaitProcessOption func(*waitProcessOption)

func newWaitProcessOption(opts ...WaitProcessOption) waitProcessOption {
	opt := waitProcessOption{
		ctx: context.Background(),
		log: logrus.WithField("pkg", "waitprocess"),
	}

	for _, o := range opts {
		o(&opt)
	}

	return opt
}

// WithLog sets the logger for the waitprocess
func WithLog(log *logrus.Entry) WaitProcessOption {
	return func(opt *waitProcessOption) {
		opt.log = log
	}
}

// WithContext sets the context for the waitprocess
func WithContext(ctx context.Context) WaitProcessOption {
	return func(opt *waitProcessOption) {
		opt.ctx = ctx
	}
}

// WithTimer sets the timer for the waitprocess
func WithTimer(timer time.Duration) WaitProcessOption {
	return func(opt *waitProcessOption) {
		opt.timer = time.NewTimer(timer)
	}
}
