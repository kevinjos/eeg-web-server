package main

import (
	"fmt"
	"github.com/kevinjos/openbci-driver/device"
	"os"
	"os/signal"
	"time"
)

func difference(x uint8, y uint8) uint8 {
	if x > y {
		return x - y
	} else if x == 0 && y == 255 {
		return 1
	} else {
		return (255 - y) + x + 1
	}
}

func decodeStream(byteStream chan byte, packetStream chan [33]byte) {
	var packet_array [33]byte
	var last_packet [33]byte
	var seq_diff uint8
	for {
		b := <-byteStream
		if b == '\xa0' {
			packet_array[0] = b
			packet_array[1] = <-byteStream
			seq_diff = difference(packet_array[1], last_packet[1])
			if seq_diff == 1 {
				for i := 2; i < 33; i++ {
					packet_array[i] = <-byteStream
				}
			} else {
				fmt.Println("Last seen sequence number [", last_packet[1], "]. This sequence number [", packet_array[1], "]")
				for seq_diff > 1 {
					packetStream <- last_packet
					seq_diff--
				}
			}
			packetStream <- packet_array
			last_packet = packet_array
		}
	}
}

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
	packetStream := make(chan [33]byte)

	//Set up the serial device
	location = "/dev/ttyUSB0"
	baud = 115200
	readTimeout = 400 * time.Microsecond
	serialDevice := device.New(location, baud, readTimeout, writeStream)
	serialDevice.Open()

	//Capture close and exit on interrupt signal
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for sig := range c {
			fmt.Println("Captured signal [", sig, "]")
			serialDevice.Close()
			os.Exit(1)
		}
	}()

	//Fire off some go routines
	go serialDevice.Stream(byteStream)
	go decodeStream(byteStream, packetStream)
	writeStream <- "b"
	last_second := time.Now().UnixNano()
	second := time.Now().UnixNano()
	for i := 0; ; i++ {
		<-packetStream
		if i%250 == 0 {
			second = time.Now().UnixNano()
			fmt.Println(second-last_second, "nanoseconds have elapsed between 250 samples")
			last_second = second
		}
	}
}
