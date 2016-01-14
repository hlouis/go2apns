// Package server provides Go2apns entry point
package main

import (
	"fmt"
	"os"
	"time"

	"github.com/hlouis/go2apns"
	"github.com/hlouis/go2apns/reader"
	"github.com/hlouis/go2apns/writer"
)

const version = "0.1.0"

type server struct {
	writer *writer.Writer
}

func myPanic(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}

func (srv *server) init(args []string) {
	// TODO: only test code here
	writer := &writer.Writer{
		ApnsEnv:     writer.DEVELOPMENT_ENV,
		KeyPairPath: "test/push",
	}

	srv.writer = writer

	//err := writer.Write(go2apns.Notification{})
	//myPanic(err)
	//err := writer.Connect()
	//defer writer.Close()
	//myPanic(err)

	//err = writer.Ping("my ping")
	//myPanic(err)
	//fmt.Fprintf(os.Stdout, "\n")
}

func (srv *server) doPush(req *go2apns.Notification) {
	out := make(chan go2apns.NotiResult)
	srv.writer.Write(req, out)
	// TODO: use some constant to define the timeout time
	timeout := time.After(5 * time.Second)

	go func() {
		select {
		case result := <-out:
			req.Result <- result
			return
		case <-timeout:
			req.Result <- go2apns.NotiResult{500, `{"reason":"Timeout"}`}
			srv.writer.Reconnect()
			return
		}
	}()
}

func (srv *server) run() {
	r := reader.Reader{
		Host: ":9090",
	}

	reqs := r.Start()
	for req := range reqs {
		// TODO: We should limit the max concurrent doPush goroutine
		srv.doPush(req)
	}
}

func main() {
	fmt.Print("Hello world!\n")

	srv := &server{}
	srv.init(os.Args[1:])
	srv.run()
}
