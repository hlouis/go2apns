package writer

import (
	"crypto/tls"
	"fmt"
	"log"
	"net"
)

func (writer *Writer) Connect() error {
	// Test code start
	cert, err := tls.LoadX509KeyPair("test/push.crt.pem", "test/push.key.pem")
	if err != nil {
		return fmt.Errorf("Error load cert and key: %v", err)
	}
	// Test code end

	host := hosts[writer.ApnsEnv]
	serverName, _, _ := net.SplitHostPort(host)
	cfg := &tls.Config{
		ServerName:   serverName,
		Certificates: []tls.Certificate{cert},
	}

	log.Printf("Connecting to %s ...", host)
	tc, err := tls.Dial("tcp", host, cfg)
	if err != nil {
		return fmt.Errorf("Error dialing %s: %v", host, err)
	}

	log.Printf("Connected to %v", tc.RemoteAddr())
	tc.Close()

	return nil
}
