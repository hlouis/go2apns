// Package server provides Go2apns entry point
package server

import (
	"fmt"
	"os"

	"github.com/hlouis/go2apns/writer"
)

const version = "0.1.0"

type Server struct {
}

func myPanic(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}

func (srv *Server) Run(args []string) {
	// TODO: only test code here
	writer := &writer.Writer{
		ApnsEnv: writer.DEVELOPMENT_ENV,
	}
	err := writer.Connect()
	defer writer.Close()
	myPanic(err)

	err = writer.Ping("my ping")
	myPanic(err)
	fmt.Fprintf(os.Stdout, "\n")
}
