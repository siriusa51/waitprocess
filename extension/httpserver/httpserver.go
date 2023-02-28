package httpserver

import (
	"context"
	"github.com/siriusa51/waitprocess"
	"github.com/sirupsen/logrus"
	"net/http"
	"time"
)

func RegisterHTTPServer(addr string, handler http.Handler, timeout ...int) error {
	t := 10
	if len(timeout) > 0 {
		t = timeout[0]
	}

	server := http.Server{
		Addr:    addr,
		Handler: handler,
	}

	err := waitprocess.RegisterProcess(waitprocess.Process{
		ServeForver: func() {
			err := server.ListenAndServe()
			if err != nil && err != http.ErrServerClosed {
				logrus.WithFields(
					logrus.Fields{
						"error": err,
					},
				).Error("http.Server.ListenAndServe() error")
			}
		},
		StopForCtx: func(ctx context.Context) {
			t, c := context.WithTimeout(ctx, time.Second*time.Duration(t))
			defer c()
			if err := server.Shutdown(t); err != nil {
				logrus.WithFields(
					logrus.Fields{
						"error": err,
					},
				).Error("http.Server.Shutdown() error")
			}
		},
	})
	return err
}
