package main

import (
	"flag"
	"net/http"
	"time"
)

var addr = flag.String("addr", ":8888", "http service address")

const (
	samplesPerSecond = 250
	packetBatchSize  = 250
	readTimeout      = 1000 * packetBatchSize / samplesPerSecond * time.Millisecond
	readBufferSize   = 33 * packetBatchSize
	baud             = 115200
	location         = "/dev/ttyUSB0"
)

func main() {
	h := NewHub()
	shutdown := make(chan bool)
	mc := NewMindControl(h.broadcast, shutdown)
	defer h.Close()
	handle := NewHandle(mc)
	http.HandleFunc("/ws", h.wsPacketHandler)
	http.HandleFunc("/", handle.rootHandler)
	http.HandleFunc("/x/", handle.commandHandler)
	http.HandleFunc("/open", handle.openHandler)
	http.HandleFunc("/reset", handle.resetHandler)
	http.HandleFunc("/start", handle.startHandler)
	http.HandleFunc("/stop", handle.stopHandler)
	http.HandleFunc("/close", handle.closeHandler)
	http.HandleFunc("/test", handle.testHandler)
	http.HandleFunc("/js/", handle.jsHandler)
	go h.Run()
	go mc.Start()
	go http.ListenAndServe(*addr, nil)
	for {
		select {
		case <-shutdown:
			return
		}
	}
}
