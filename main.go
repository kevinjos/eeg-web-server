package main

import (
	"flag"
	"log"
	"net/http"
	"time"
)

var mc *MindControl = NewMindControl()
var addr = flag.String("addr", ":8888", "http service address")

const (
	samplesPerSecond = 250
	packetBatchSize  = 100
	readTimeout      = 1000 * packetBatchSize / samplesPerSecond * time.Millisecond
	readBufferSize   = 33 * packetBatchSize
	baud             = 115200
	location         = "/dev/ttyUSB0"
)

func sendPackets() {
	last_second := time.Now().UnixNano()
	second := time.Now().UnixNano()

	pb := NewPacketBatcher()
	for i := 1; ; i++ {
		p := <-mc.PacketStream
		pb.packets[i%packetBatchSize] = p

		if i%packetBatchSize == 0 {
			pb.batch()
			h.broadcast <- pb
			pb = NewPacketBatcher()
		}
		if i%250 == 0 {
			second = time.Now().UnixNano()
			log.Println(second-last_second, "nanoseconds have elapsed between 250 samples.")
			last_second = second
		}
	}
}

func main() {
	http.HandleFunc("/", rootHandler)
	http.HandleFunc("/x/", commandHandler)
	http.HandleFunc("/ws", wsPacketHandler)
	http.HandleFunc("/open", openHandler)
	http.HandleFunc("/reset", resetHandler)
	http.HandleFunc("/start", startHandler)
	http.HandleFunc("/stop", stopHandler)
	http.HandleFunc("/close", closeHandler)
	http.HandleFunc("/test", testHandler)
	http.HandleFunc("/js/", jsHandler)
	mindController := MindController(mc)
	go h.Run()
	go sendPackets()
	go mindController.DecodeStream()
	go mindController.ReadWriteClose()
	for {
		http.ListenAndServe(*addr, nil)
	}
}
