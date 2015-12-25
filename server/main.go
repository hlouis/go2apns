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

func (srv *Server) Run(args []string) {
	// TODO: only test code here
	writer := &writer.Writer{writer.DEVELOPMENT_ENV}
	err := writer.Connect()

	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
	fmt.Fprintf(os.Stdout, "\n")
}
