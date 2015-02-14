package main

import (
	"github.com/tarm/goserial"
	"io"
	"log"
	"os"
	"time"
)

type Device struct {
	Location    string
	Baud        int
	ReadTimeout time.Duration
	writeStream chan string
	byteStream  chan byte
	resetButton chan bool
	quitButton  chan bool
	conn        io.ReadWriteCloser
}

func NewDevice(Location string, Baud int, ReadTimeout time.Duration,
	byteStream chan byte, writeStream chan string, quitButton chan bool, resetButton chan bool) *Device {
	d := &Device{
		Location:    Location,
		Baud:        Baud,
		ReadTimeout: ReadTimeout,
		byteStream:  byteStream,
		writeStream: writeStream,
		quitButton:  quitButton,
		resetButton: resetButton,
	}
	return d
}

func (d *Device) readWriteClose() {
	streamingData := false
	for {
		select {
		case s := <-d.writeStream:
			d.write(s)
			switch {
			case s == "s" || s == "v":
				streamingData = false
			case s == "b":
				streamingData = true
			}
		case <-d.resetButton:
			d.reset()
		case <-d.quitButton:
			defer func() {
				d.write("s")
				d.conn.Close()
				log.Println("Safely closed the device")
				os.Exit(1)
			}()
			return
		default:
			if streamingData {
				d.read()
			} else {
				time.Sleep(100 * time.Millisecond)
			}
		}
	}
}

func (d *Device) read() {
	buf := make([]byte, readBufferSize-1)
	n, err := d.conn.Read(buf)
	if err != nil {
		log.Println("Error reading [", n, "] bytes from serial device: [", err, "]")
	} else if n > 0 {
		for i := 0; i < n; i++ {
			d.byteStream <- buf[i]
		}
	}
}

func (d *Device) testRead(bytestream chan uint8) {
	var readstate uint8
	var seqNum uint8

	for {
		switch readstate {
		case 0:
			d.byteStream <- '\xa0'
			readstate++
		case 1:
			d.byteStream <- seqNum
			seqNum++
			readstate++
		case 2:
			for i := 0; i < 30; i++ {
				d.byteStream <- '\xa1'
			}
			readstate++
		case 3:
			d.byteStream <- '\xc0'
			readstate = 0
		}
	}
}

func (d *Device) write(s string) {
	wb := []byte(s)
	if n, err := d.conn.Write(wb); err != nil {
		log.Println("Error writing [", n, "] bytes to serial device: [", err, "]")
	} else {
		log.Println("Wrote [", n, "] byte", wb, "to the serial device")
	}
	return
}

func (d *Device) open() {
	config := &serial.Config{Name: d.Location, Baud: d.Baud, ReadTimeout: d.ReadTimeout}
	conn, err := serial.OpenPort(config)
	if err != nil {
		log.Println("Error conneting to serial device: [", err, "]")
		os.Exit(1)
	}
	d.conn = conn
}

//Reset sends the stop and reset message to the serial device,
//reads up to the init message [$$$], then sends the message
//to start the binary data stream
func (d *Device) reset() {
	var (
		scrolling  [3]byte
		init_array [3]byte
		index      int
	)

	init_array = [3]byte{'\x24', '\x24', '\x24'}

	d.write("s")
	time.Sleep(1 * time.Second)
	d.write("v")
	time.Sleep(1 * time.Second)

	for {
		select {
		case b := <-d.byteStream:
			scrolling[index%3] = b
			index++
		default:
			if scrolling == init_array {
				d.writeStream <- "b"
				return
			} else {
				d.read()
			}
		}
	}
}
