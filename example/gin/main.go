package main

import (
	"github.com/gin-gonic/gin"
	"github.com/siriusa51/waitprocess"
	"github.com/siriusa51/waitprocess/extension/httpserver"
	"io/ioutil"
	"net/http"
	"time"
)

func main() {
	router := gin.New()
	router.GET("/", func(context *gin.Context) {
		context.String(200, "okokok")
	})

	// register router
	if err := httpserver.RegisterHTTPServer(":8080", router); err != nil {
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
