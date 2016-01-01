// Package writer provides methods to write notifications to APNS
// and keep the http/2 connections with APNS
package writer

import (
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"net"

	"golang.org/x/net/http2"
)

const (
	DEVELOPMENT_ENV = "development"
	PRODUCTION_ENV  = "production"
)

var hosts = map[string]string{
	DEVELOPMENT_ENV: "api.development.push.apple.com:443",
	PRODUCTION_ENV:  "api.push.apple.com:443",
}

type Writer struct {
	// ApnsEnv use to specify which APNS environment to use
	ApnsEnv string

	tc     *tls.Conn
	framer *http2.Framer
}

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
	writer.tc = tc
	writer.framer = http2.NewFramer(tc, tc)

	log.Printf("Connected to %v", tc.RemoteAddr())
	if err := tc.Handshake(); err != nil {
		return fmt.Errorf("TLS handshake: %v", err)
	}
	if err := tc.VerifyHostname(serverName); err != nil {
		return fmt.Errorf("VerifyHostname: %v", err)
	}
	state := tc.ConnectionState()
	log.Printf("Negotiated protocol %q", state.NegotiatedProtocol)
	//if !state.NegotiatedProtocolIsMutual || state.NegotiatedProtocol == "" {
	//return fmt.Errorf("Could not negotiate protocol mutually")
	//}

	// write http2 preface
	if _, err := io.WriteString(tc, http2.ClientPreface); err != nil {
		return err
	}

	// write an empty setting frame
	if err := writer.framer.WriteSettings(); err != nil {
		return err
	}

	//return nil
	// TODO: test code
	return readFrame(writer.framer)
}

func (writer *Writer) Ping(str string) error {
	var data [8]byte
	copy(data[:], str)
	//return writer.framer.WritePing(true, data)

	// TODO: test code
	err := writer.framer.WritePing(false, data)

	if err != nil {
		return err
	}

	return readFrames(writer.framer)
}

func (writer *Writer) Close() error {
	log.Printf("Close connections by request!")
	return writer.tc.Close()
}

// // //////////////////
//   Helper methods  //
// ////////////////////

func readFrames(fr *http2.Framer) error {
	for {
		err := readFrame(fr)
		if err != nil {
			return err
		}
	}
	return nil
}

func readFrame(fr *http2.Framer) error {
	f, err := fr.ReadFrame()
	if err != nil {
		return fmt.Errorf("ReadFrame: %v", err)
	}
	log.Printf("%v", f)
	switch f := f.(type) {
	case *http2.PingFrame:
		log.Printf("  Data = %q", f.Data)
	case *http2.SettingsFrame:
		f.ForeachSetting(func(s http2.Setting) error {
			log.Printf("  %v", s)
			return nil
		})
	case *http2.WindowUpdateFrame:
		log.Printf("  Window-Increment = %v\n", f.Increment)
	case *http2.GoAwayFrame:
		log.Printf("  Last-Stream-ID = %d; Error-Code = %v (%d)\n", f.LastStreamID, f.ErrCode, f.ErrCode)
	case *http2.DataFrame:
		log.Printf("  %q", f.Data())
	case *http2.HeadersFrame:
		if f.HasPriority() {
			log.Printf("  PRIORITY = %v", f.Priority)
		}
	}

	return nil
}
