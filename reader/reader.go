package reader

import (
	"fmt"
	"log"
	"net/http"

	"github.com/hlouis/go2apns"
)

type Reader struct {
	Host string // server addres:port

	results chan *go2apns.Notification
}

// result is json
// { "error":0 }
// { "error":1, "msg":"wrong bundle id" }
func writeRes(failed bool, msg string, w http.ResponseWriter) {
	if failed {
		fmt.Fprintf(w, "{\"error\":1,\"msg\":\"%s\"}", msg)
	} else {
		fmt.Fprintf(w, "{\"error\":0},\"msg\":\"%s\"}", msg)
	}
}

func handler(
	w http.ResponseWriter,
	r *http.Request,
	reqs chan *go2apns.Notification) {
	//fmt.Fprintf(w, "hello %s", r.URL.Path[1:])
	log.Printf("Got one req:%s", r.URL.Path)

	no := go2apns.Notification{}

	if r.Method != "POST" {
		writeRes(true, "Only accept POST!", w)
		return
	}

	no.Token = r.PostFormValue("token")
	no.Expiration = r.PostFormValue("expiration")
	no.Priority = r.PostFormValue("priority")
	no.Topic = r.PostFormValue("topic")
	no.Payload = r.PostFormValue("payload")

	log.Printf("Got post data: %v", no)

	resc := make(chan string)
	no.Result = resc
	reqs <- &no
	writeRes(false, <-resc, w)
	//writeRes(false, no.Payload, w)
}

func makeHandler(fn func(http.ResponseWriter, *http.Request, chan *go2apns.Notification), results chan *go2apns.Notification) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fn(w, r, results)
	}
}

func (r *Reader) Start() <-chan *go2apns.Notification {
	// TODO: We should close this channel when receive close signal
	r.results = make(chan *go2apns.Notification)
	http.HandleFunc("/push", makeHandler(handler, r.results))
	go func() {
		log.Println("Start serve the http service")
		err := http.ListenAndServe(r.Host, nil)
		if err != nil {
			log.Fatal("ListenAndServe:", err)
		}
	}()
	return r.results
}
