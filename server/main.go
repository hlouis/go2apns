// Package server provides Go2apns entry point
package main

import (
	"fmt"
	"os"

	"github.com/hlouis/go2apns"
	"github.com/hlouis/go2apns/writer"
)

const version = "0.1.0"

type server struct {
}

func myPanic(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}

func (srv *server) run(args []string) {
	// TODO: only test code here
	writer := &writer.Writer{
		ApnsEnv:     writer.DEVELOPMENT_ENV,
		KeyPairPath: "test/push",
	}

	err := writer.Write(go2apns.Notification{})
	myPanic(err)
	//err := writer.Connect()
	//defer writer.Close()
	//myPanic(err)

	//err = writer.Ping("my ping")
	//myPanic(err)
	//fmt.Fprintf(os.Stdout, "\n")
}

func main() {
	fmt.Print("Hello world!\n")

	srv := &server{}
	srv.run(os.Args[1:])
}
