package gin

import (
	"github.com/gin-gonic/gin"
	"github.com/siriusa51/waitprocess/extension/httpserver"
)

func RegisterEngine(addr string, engine *gin.Engine, timeout ...int) error {
	return httpserver.RegisterHTTPServer(addr, engine, timeout...)
}
