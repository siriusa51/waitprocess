package main

import (
	"context"
	"github.com/siriusa51/waitprocess"
	"time"
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
