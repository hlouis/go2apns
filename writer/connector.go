// act like a http/2 client to APNS
package writer

import (
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"net"

	"golang.org/x/net/http2"
)

// connection hold http/2 connection
type connection struct {
	tc       *tls.Conn                  // tc is the basic tls connection
	framer   *http2.Framer              // framer use to write/read frame from tc
	settings map[http2.SettingID]uint32 // setting from server
	streamID uint32                     // stream id for this connection
}

type stream struct {
	id uint32 // this stream id
}

func connect(host string, certPath string, keyPath string) (con *connection, err error) {
	fmt.Printf("Start to connect %v with cert: %v, key: %v\n", host, certPath, keyPath)
	con = &connection{
		settings: make(map[http2.SettingID]uint32),
	}
	// load cert
	cert, err := tls.LoadX509KeyPair(certPath, keyPath)
	if err != nil {
		return nil, fmt.Errorf("Error load cert and key: %v", err)
	}

	// connect the TLS socket
	serverName, _, err := net.SplitHostPort(host)
	if err != nil {
		return nil, fmt.Errorf("Error parse host string: %v", err)
	}

	cfg := &tls.Config{
		ServerName:   serverName,
		Certificates: []tls.Certificate{cert},
	}

	tc, err := tls.Dial("tcp", host, cfg)
	if err != nil {
		return nil, fmt.Errorf("Error dialing %s: %v", host, err)
	}

	if err := tc.Handshake(); err != nil {
		return nil, fmt.Errorf("TLS handshake: %v", err)
	}
	if err := tc.VerifyHostname(serverName); err != nil {
		return nil, fmt.Errorf("VerifyHostname: %v", err)
	}
	con.tc = tc
	con.framer = http2.NewFramer(tc, tc)

	// send http/2 prefix
	if _, err := io.WriteString(tc, http2.ClientPreface); err != nil {
		return nil, err
	}

	// write an empty setting frame
	if err := con.framer.WriteSettings(); err != nil {
		return nil, err
	}
	return con, nil
}

func (con *connection) close() {
	con.tc.Close()
}

func (con *connection) ping(str string) error {
	var data [8]byte
	copy(data[:], str)
	//return writer.framer.WritePing(true, data)

	// TODO: test code
	err := con.framer.WritePing(false, data)

	if err != nil {
		return err
	}

	return readFrame(con.framer)
}

func (con *connection) openStream(header http2.HeadersFrameParam) (s *stream, err error) {
	if con.streamID == 0 {
		con.streamID = 1
	} else {
		con.streamID += 2
	}

	s = &stream{
		id: con.streamID,
	}

	header.StreamID = con.streamID
	err = con.framer.WriteHeaders(header)
	return s, err
}

// // //////////////////
//   Helper methods  //
// ////////////////////

func readFrames(fr *http2.Framer) error {
	fmt.Print("Start to read frames to dead!\n")
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
