package gin

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/siriusa51/waitprocess"
	"github.com/sirupsen/logrus"
	"net/http"
	"time"
)

func RegisterEngine(addr string, engine *gin.Engine) error {
	server := http.Server{
		Addr:    addr,
		Handler: engine,
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
		Stop: func(ctx context.Context) {
			t, c := context.WithTimeout(ctx, time.Second*10)
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
