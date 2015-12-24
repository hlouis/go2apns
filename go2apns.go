// Package main provides whole program entry point
package main

import (
	"fmt"
	"os"

	"github.com/hlouis/go2apns/server"
)

func main() {
	fmt.Print("Hello world!\n")

	server = &server.Server{}
	server.Run(os.Args[1:])
}
