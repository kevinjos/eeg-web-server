package main

import (
	"github.com/tarm/goserial"
	"io"
	"log"
	"os"
	"time"
)

type Device struct {
	writeStream chan string
	byteStream  chan byte
	rtStream    chan bool
	resetButton chan bool
	quitButton  chan bool
	conn        io.ReadWriteCloser
}

func (d *Device) ReadWriteClose() {
	streamingData := false
	for {
		select {
		case s := <-d.writeStream:
			if s == "s" {
				streamingData = false
			} else if s == "b" {
				streamingData = true
			}
			d.write(s)
		case <-d.resetButton:
			streamingData = false
			go d.reset()
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
				buf := make([]byte, readBufferSize-1)
				d.read(buf)
			} else {
				time.Sleep(4 * time.Millisecond)
			}
		}
	}
}

func (d *Device) read(buf []byte) {
	n, err := d.conn.Read(buf)
	if err == io.EOF {
		log.Println("Read timeout")
		d.rtStream <- true
	} else if err != nil {
		log.Fatal("Error reading [", n, "] bytes from serial device: [", err, "]")
	}
	if n > 0 {
		for i := 0; i < n; i++ {
			d.byteStream <- buf[i]
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
}

func (d *Device) open() {
	config := &serial.Config{Name: location, Baud: baud, ReadTimeout: readTimeout}
	conn, err := serial.OpenPort(config)
	if err != nil {
		log.Fatal("Error conneting to serial device: [", err, "]")
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

	d.writeStream <- "s"
	time.Sleep(400 * time.Millisecond)
	d.writeStream <- "v"
	time.Sleep(1000 * time.Millisecond)

	init_array = [3]byte{'\x24', '\x24', '\x24'}

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
				buf := make([]byte, 1)
				d.read(buf)
			}
		}
	}
}
