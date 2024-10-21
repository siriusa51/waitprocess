# WaitProcess

This is the library that can manage multiple infinite loops of methods, making them start together or close together, and also allowing to register signal to close.

## Installation

Use go get.

```shell
go get -u github.com/siriusa51/waitprocess/v2
```

Then import the package into your own code.

```
import "github.com/siriusa51/waitprocess/v2"
```



## Example

```go
package main

import (
	"context"
	"github.com/siriusa51/waitprocess/v2"
	"github.com/siriusa51/waitprocess/v2/ext/http_srv"
	"net/http"
	"os"
)

func main() {
	waitprocess.RegisterProcess("app1", waitprocess.RunWithStopFunc(
		func() error {
			// do something and run forever
			return nil
		},
		func() {
			// do something and stop
		},
	))

	waitprocess.RegisterProcess("app2", waitprocess.RunWithCtx(
		func(ctx context.Context) error {
			// do something and run forever, stop by context
			return nil
		},
	))

	waitprocess.RegisterProcess("app3", waitprocess.RunWithChan(
		func(i <-chan struct{}) error {
			// do something and run forever, stop by channel
			return nil
		},
	))

	// http srv example
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
		writer.Write([]byte("hello"))
	})
	http_srv.RegisterHttpSrv(":5050", mux)

	// register signal to waitprocess
	waitprocess.RegisterSignal(os.Interrupt, os.Kill)

	// start waitprocess, app1, app2, app3, http server will start together
	// if any app stop or receive signal, all will stop
	if err := waitprocess.Run(); err != nil {
		panic(err)
	}
}

```

