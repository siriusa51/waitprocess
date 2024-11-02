package http_srv

import (
	"context"
	"github.com/siriusa51/waitprocess/v2"
	"github.com/sirupsen/logrus"
	"net/http"
	"time"
)

type httpServerOption struct {
	wp            *waitprocess.WaitProcess
	timeout       time.Duration
	name          string
	log           *logrus.Entry
	afterStopHook func()
}

type HttpServerOptionFunc func(*httpServerOption)

func newHTTPServerOption(opts ...HttpServerOptionFunc) *httpServerOption {
	opt := &httpServerOption{
		name:    "http_srv",
		timeout: time.Second * 15,
		wp:      waitprocess.Default(),
		log:     logrus.WithField("pkg", "waitprocess/http_srv"),
	}

	for _, o := range opts {
		o(opt)
	}
	return opt
}

func WithName(name string) HttpServerOptionFunc {
	return func(opt *httpServerOption) {
		opt.name = name
	}
}

func WithTimeout(timeout time.Duration) HttpServerOptionFunc {
	return func(opt *httpServerOption) {
		opt.timeout = timeout
	}
}

func WithWaitProcess(wp *waitprocess.WaitProcess) HttpServerOptionFunc {
	return func(opt *httpServerOption) {
		opt.wp = wp
	}
}

func WithAfterStopHook(f func()) HttpServerOptionFunc {
	return func(opt *httpServerOption) {
		opt.afterStopHook = f
	}
}

func RegisterHttpSrv(addr string, handler http.Handler, fs ...HttpServerOptionFunc) *waitprocess.WaitProcess {
	opt := newHTTPServerOption(fs...)

	srv := http.Server{
		Addr:    addr,
		Handler: handler,
	}

	return opt.wp.RegisterProcess(opt.name, waitprocess.RunWithStopFunc(
		func() error {
			err := srv.ListenAndServe()
			if err != nil && err != http.ErrServerClosed {
				return err
			}

			return nil
		},
		func() {
			ctx, cancel := context.WithTimeout(context.Background(), opt.timeout)
			defer cancel()

			if err := srv.Shutdown(ctx); err != nil {
				opt.log.WithError(err).Error("http.Server.Shutdown() error")
			}

			if opt.afterStopHook != nil {
				opt.afterStopHook()
			}
		},
	))
}
