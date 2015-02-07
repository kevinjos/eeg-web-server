package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"time"
)

var addr = flag.String("addr", ":8888", "http service address")
var mc *MindController = NewMindController()

func main() {
	//Capture close and exit on interrupt signal
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		switch {
		case os.Interrupt == <-c:
			mc.QuitButton <- true
		}
	}()

	last_second := time.Now().UnixNano()
	second := time.Now().UnixNano()

	http.HandleFunc("/", rootHandler)
	http.HandleFunc("/ws", wsPacketHandler)
	http.HandleFunc("/reset", resetHandler)
	http.HandleFunc("/start", startHandler)
	http.HandleFunc("/stop", stopHandler)
	http.HandleFunc("/close", closeHandler)

	mc.Open()
	go h.run()
	go http.ListenAndServe(*addr, nil)

	for i := 0; ; i++ {
		p := <-mc.PacketStream

		if i%1250 == 0 {
			h.broadcast <- p
			second = time.Now().UnixNano()
			fmt.Println(second-last_second, "nanoseconds have elapsed between 1250 samples.")
			//fmt.Println("Chans 1-4:", p.chan1, p.chan2, p.chan3, p.chan4)
			//fmt.Println("Chans 5-8:", p.chan5, p.chan6, p.chan7, p.chan8)
			//fmt.Println("Acc Data:", p.accX, p.accY, p.accZ)
			last_second = second
		}
	}
}
