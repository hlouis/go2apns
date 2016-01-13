// Package writer provides methods to write notifications to APNS
// and keep the http/2 connections with APNS
package writer

import (
	"bytes"
	"crypto/rand"
	"fmt"
	"log"
	"strconv"

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
	conn       *connection
	henc       *hpack.Encoder
	hbuf       bytes.Buffer
	hdec       *hpack.Decoder
	lastHeader map[string]string

	pushResults chan pushResult
	monitors    chan monitor
}

type pushResult struct {
	StreamID uint32
	Status   string
	Msg      string
}

type monitor struct {
	StreamID uint32
	out      chan string
}

func watch(prs <-chan pushResult, monitors <-chan monitor) {
	cbs := make(map[uint32]monitor)

	go func() {
		for {
			select {
			case pr := <-prs:
				m := cbs[pr.StreamID]
				m.out <- pr.Msg
			case m := <-monitors:
				cbs[m.StreamID] = m
			}
		}
	}()
}

// Write push notification to APNS
func (w *Writer) Write(n *go2apns.Notification, out chan string) error {
	if w.conn == nil {
		con, err := connect(hosts[w.ApnsEnv], "test/push.crt.pem", "test/push.key.pem")
		if err != nil {
			return err
		}
		w.conn = con

		w.henc = hpack.NewEncoder(&w.hbuf)
		w.lastHeader = make(map[string]string)
		w.pushResults = make(chan pushResult)
		w.monitors = make(chan monitor)
		watch(w.pushResults, w.monitors)
		go w.readFrames()
	}

	w.hbuf.Reset()
	w.henc.WriteField(hpack.HeaderField{Name: ":method", Value: "POST"})
	w.henc.WriteField(hpack.HeaderField{Name: ":path", Value: fmt.Sprintf("/3/device/%s", n.Token), Sensitive: true})
	log.Println(fmt.Sprintf("/3/device/%s", n.Token))
	if w.conn.streamID < 3 {
		//w.henc.WriteField(hpack.HeaderField{Name: "apns-id", Value: pseudo_uuid()})
		w.henc.WriteField(hpack.HeaderField{Name: "apns-expiration", Value: n.Expiration})
	} else {
		//w.henc.WriteField(hpack.HeaderField{Name: "apns-id", Value: pseudo_uuid(), Sensitive: true})
		w.henc.WriteField(hpack.HeaderField{Name: "apns-expiration", Value: n.Expiration, Sensitive: true})
	}
	w.henc.WriteField(hpack.HeaderField{Name: "apns-priority", Value: n.Priority})
	w.henc.WriteField(hpack.HeaderField{Name: "content-length", Value: strconv.Itoa(len(n.Payload))})
	w.henc.WriteField(hpack.HeaderField{Name: "apns-topic", Value: n.Topic})

	hbf := w.hbuf.Bytes()
	log.Printf("\nhbf len: %v\n\n", len(hbf))
	header := http2.HeadersFrameParam{
		BlockFragment: hbf,
		EndHeaders:    true,
	}

	s, e := w.conn.openStream(header)
	w.monitors <- monitor{s.id, out}
	if e != nil {
		return fmt.Errorf("Open stream error: %v", e)
	}

	e = w.conn.framer.WriteData(s.id, true, []byte(n.Payload))
	if e != nil {
		return fmt.Errorf("Write data error: %v", e)
	}

	return nil
}

func (w *Writer) Reconnect() {
	if w.conn != nil {
		w.conn.close()
		w.conn = nil
	}
}

// // ////////////////////
//   Helper functions  //
// //////////////////////

func (w *Writer) readFrames() error {
	fmt.Print("Start to read frames to dead!\n")
	for {
		err := w.readFrame()
		if err != nil {
			return err
		}
	}
	return nil
}

func (w *Writer) readFrame() error {
	f, err := w.conn.framer.ReadFrame()
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
		if f.StreamEnded() {
			w.pushResults <- pushResult{
				f.StreamID, w.lastHeader[":status"], string(f.Data()),
			}
		}
	case *http2.HeadersFrame:
		w.lastHeader = make(map[string]string)
		if f.HasPriority() {
			log.Printf("  PRIORITY = %v", f.Priority)
		}
		if w.hdec == nil {
			// TODO: if the user need to send a SETTINGS frame advertising
			// something larger, we'll need to respect SETTINGS_HEADER_TABLE_SIZE
			// and stuff here instead of using the 4k default. But for now:
			tableSize := uint32(4 << 10)
			w.hdec = hpack.NewDecoder(tableSize, w.onNewHeaderField)
		}
		w.hdec.Write(f.HeaderBlockFragment())
		if w.lastHeader[":status"] == "200" && f.StreamEnded() {
			w.pushResults <- pushResult{
				f.StreamID, w.lastHeader[":status"], "success",
			}
		}
	}

	return nil
}

func (w *Writer) onNewHeaderField(f hpack.HeaderField) {
	if f.Sensitive {
		log.Printf("  %s = %q (SENSITIVE)", f.Name, f.Value)
	}
	log.Printf("  %s = %q", f.Name, f.Value)
	w.lastHeader[f.Name] = f.Value
}

func pseudo_uuid() (uuid string) {

	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		fmt.Println("Error: ", err)
		return
	}

	uuid = fmt.Sprintf("%X-%X-%X-%X-%X", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])

	return
}
