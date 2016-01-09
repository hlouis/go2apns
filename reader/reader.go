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

func handler(
	w http.ResponseWriter,
	r *http.Request,
	reqs chan *go2apns.Notification) {
	//fmt.Fprintf(w, "hello %s", r.URL.Path[1:])
	log.Printf("Got one req:%s", r.URL.Path)
	res := make(chan string)
	reqs <- &go2apns.Notification{
		Alert:  "hehe",
		Result: res,
	}

	fmt.Fprintf(w, <-res)
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
