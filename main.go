package main

import (
	"fmt"
	"os"
	"os/signal"
	"time"
)

func main() {
	//Configuration parameters for serial IO
	var (
		location    string
		baud        int
		readTimeout time.Duration
	)

	//Make some channels
	byteStream := make(chan byte, 256*33)
	writeStream := make(chan string)
	packetStream := make(chan *Packet)
	quitChan := make(chan bool)

	//Capture close and exit on interrupt signal
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		switch {
		case os.Interrupt == <-c:
			quitChan <- true
		}
	}()

	//Set up the serial device
	location = "/dev/ttyUSB0"
	baud = 115200
	readTimeout = 5 * time.Second
	serialDevice := NewDevice(location, baud, readTimeout, byteStream, writeStream, quitChan)
	serialDevice.Open()

	//Start streaming data from the device
	writeStream <- "b"
	go decodeStream(byteStream, packetStream)
	last_second := time.Now().UnixNano()
	second := time.Now().UnixNano()

	for i := 0; ; i++ {
		<-packetStream
		if i%250 == 0 {
			second = time.Now().UnixNano()
			fmt.Println(second-last_second, "nanoseconds have elapsed between 250 samples.")
			//fmt.Println("Chans 1-4:", p.chan1, p.chan2, p.chan3, p.chan4)
			//fmt.Println("Chans 5-8:", p.chan5, p.chan6, p.chan7, p.chan8)
			//fmt.Println("Acc Data:", p.accX, p.accY, p.accZ)
			last_second = second
		}
	}
}
