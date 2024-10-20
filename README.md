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
	"net/http"
	"os"

	"github.com/siriusa51/waitprocess/v2"
	"github.com/siriusa51/waitprocess/v2/ext/http_srv"
)

func main() {
	waitprocess.RegisterProcess("app1", waitprocess.RunWithStopFunc(
		func() {
			// do something and run forever
		},
		func() {
			// do something and stop
		},
	))

	waitprocess.RegisterProcess("app2", waitprocess.RunWithCtx(
		func(ctx context.Context) {
			// do something and run forever, stop by context
		},
	))

	waitprocess.RegisterProcess("app3", waitprocess.RunWithChan(
		func(ch <-chan struct{}) {
			// do something and run forever, stop by channel
		},
	))

	// http srv example
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
		writer.Write([]byte("hello"))
	})
	http_srv.RunWithHttpSrv(":5050", mux)

	// register signal to waitprocess
	waitprocess.RegisterSignal(os.Interrupt, os.Kill)

	// start waitprocess
	if err := waitprocess.Run(); err != nil {
		panic(err)
	}
}
```

