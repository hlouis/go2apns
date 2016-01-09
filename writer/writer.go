// Package writer provides methods to write notifications to APNS
// and keep the http/2 connections with APNS
package writer

import (
	"bytes"
	"fmt"

	"github.com/hlouis/go2apns"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/hpack"
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

	// keypair path for this cert and key pem file
	// TODO: use this field
	KeyPairPath string

	// for internal use
	conn *connection
	henc *hpack.Encoder
	hbuf bytes.Buffer
}

// Write push notification to APNS
func (w *Writer) Write(n *go2apns.Notification, out chan string) error {
	if w.conn == nil {
		con, err := connect(hosts[w.ApnsEnv], "test/push.crt.pem", "test/push.key.pem")
		if err != nil {
			return err
		}
		w.conn = con
	}

	if w.henc == nil {
		w.henc = hpack.NewEncoder(&w.hbuf)
	}

	w.henc.WriteField(hpack.HeaderField{Name: ":method", Value: "POST"})
	w.henc.WriteField(hpack.HeaderField{Name: ":path", Value: "/3/device/a98c857e079bbe143a6a48a4e671b2a480af826de2ef3d9fb0172922fbd7b15f", Sensitive: true})
	w.henc.WriteField(hpack.HeaderField{Name: "apns-expiration", Value: "0"})
	w.henc.WriteField(hpack.HeaderField{Name: "apns-priority", Value: "10"})
	w.henc.WriteField(hpack.HeaderField{Name: "content-length", Value: "33"})
	w.henc.WriteField(hpack.HeaderField{Name: "apns-topic", Value: "mobi.xy3d.Go2ApnsTest"})

	hbf := w.hbuf.Bytes()
	header := http2.HeadersFrameParam{
		BlockFragment: hbf,
		EndHeaders:    true,
	}

	s, e := w.conn.openStream(header)
	if e != nil {
		return fmt.Errorf("Open stream error: %v", e)
	}

	payload := "{ \"aps\" : { \"alert\" : \"H1llo\" } }"
	e = w.conn.framer.WriteData(s.id, true, []byte(payload))
	if e != nil {
		return fmt.Errorf("Write data error: %v", e)
	}

	return nil
}

// // ////////////////////
//   Helper functions  //
// //////////////////////
