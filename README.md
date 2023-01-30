# WAITPROCESS

This is the library that can manage multiple infinite loops of methods, making them start together or close together, and also allowing to register signal to close.

## Example

### simple

```go
package main

import (
	"context"
	"time"
	"github.com/siriusa51/waitprocess"
)

type sleep struct {
	tag int
}

func (s *sleep) sleep() {
	println("call sleep()", time.Now().String())
	for s.tag == 0 {
		time.Sleep(time.Second)
		println("sleep", time.Now().String())
	}
	println("stop sleep()", time.Now().String())
}

func (s *sleep) stop(ctx context.Context) {
	println("call stop()", time.Now().String())
	time.Sleep(time.Second)
	s.tag = 1
}

func main() {
	s := sleep{}
	err := waitprocess.RegisterProcess(waitprocess.Process{
		ServeForver:        s.sleep,
		Stop:               s.stop,
		ServeForverWithCtx: nil,
	})

	if err != nil {
		panic(err)
	}

	println("call waitprocess.Start()", time.Now().String())
	_ = waitprocess.Start()

	time.Sleep(time.Second * time.Duration(5))

	println("call waitprocess.Stop()", time.Now().String())
	_ = waitprocess.Stop()

	println("call waitprocess.Wait()", time.Now().String())
	waitprocess.Wait()
	println("finish", time.Now().String())
}

```

### Gin

```go
package main

import (
	egin "github.com/siriusa51/waitprocess/extension/gin"
	"github.com/gin-gonic/gin"
	"net/http"
	"time"
	"io/ioutil"
	"github.com/siriusa51/waitprocess"
)

func main() {
	router := gin.New()
	router.GET("/", func(context *gin.Context) {
		context.String(200, "okokok")
	})

	// register router
	if err := egin.RegisterEngine(":8080", router); err != nil {
		panic(err)
	}

	// start router
	if err := waitprocess.Start(); err != nil {
		panic(err)
	}

	time.Sleep(time.Second)
	// call router api
	resp, err := http.Get("http://localhost:8080/")
	if err != nil {
		panic(err)
	}

	buff, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	println("response:", string(buff))
	defer resp.Body.Close()

	// stop router
	err = waitprocess.StopAndWait()
	if err != nil {
		panic(err)
	}
}

```

