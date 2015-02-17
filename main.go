package main

import (
	"flag"
	"net/http"
	"time"
)

var addr = flag.String("addr", ":8888", "http service address")

const (
	samplesPerSecond = 250
	packetBatchSize  = 100
	readTimeout      = 1000 * packetBatchSize / samplesPerSecond * time.Millisecond
	readBufferSize   = 33 * packetBatchSize
	baud             = 115200
	location         = "/dev/ttyUSB0"
)

func main() {
	var mc *MindControl = NewMindControl()
	handle := NewHandle(mc)
	http.HandleFunc("/ws", wsPacketHandler)
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
	go mc.sendPackets()
	go mc.DecodeStream()
	go mc.ReadWriteClose()
	for {
		http.ListenAndServe(*addr, nil)
	}
}
