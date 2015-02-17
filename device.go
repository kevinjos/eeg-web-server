package main

import (
	"github.com/tarm/goserial"
	"io"
	"log"
	"os"
	"time"
)

type OpenBCI struct {
	writeChan chan string
	readChan  chan byte
	timeoutChan    chan bool
	resetChan chan chan bool
	quitChan  chan bool
	pauseReadChan chan chan bool
	conn        io.ReadWriteCloser
}

func NewOpenBCI() *OpenBCI {
	return &OpenBCI{
		writeChan: make(chan string, 64),
		readChan:	make(chan byte, readBufferSize),
		timeoutChan: make(chan bool),
		resetChan:	make(chan chan bool),
		quitChan:	make(chan bool),
		pauseReadChan: make(chan chan bool),
	}
}

func (d *OpenBCI) ReadWriteClose() {
	for {
		select {
		case s := <-d.writeChan:
			d.write(s)
		case resumePacketStream := <-d.resetChan:
			go d.reset(resumePacketStream)
		case <-d.quitChan:
			defer func() {
				d.write("s")
				d.conn.Close()
				log.Println("Safely closed the device")
				os.Exit(1)
			}()
			return
		}
	}
}

func (d *OpenBCI) read() {
	buf := make([]byte, 8)
	for {
		select {
		case resumeReadChan := <-d.pauseReadChan:
			<-resumeReadChan
		default:
			n, err := d.conn.Read(buf)
			if err == io.EOF {
				d.timeoutChan <- true
			} else if err != nil {
				log.Fatal("Error reading [", n, "] bytes from serial device: [", err, "]")
			}
			for i := 0; i < n; i++ {
				d.readChan <- buf[i]
			}
		}
	}
}

func (d *OpenBCI) write(s string) {
	wb := []byte(s)
	if n, err := d.conn.Write(wb); err != nil {
		log.Println("Error writing [", n, "] bytes to serial device: [", err, "]")
	} else {
		log.Println("Wrote [", n, "] byte", wb, "to the serial device")
	}
}

func (d *OpenBCI) open() {
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
func (d *OpenBCI) reset(resumeChan chan bool) {
	var (
		scrolling  [3]byte
		init_array [3]byte
		index      int
	)

	d.writeChan <- "s"
	time.Sleep(400 * time.Millisecond)
	resumeRead := make(chan bool)
	defer close(resumeRead)
	d.pauseReadChan <- resumeRead
	d.writeChan <- "v"
	time.Sleep(1000 * time.Millisecond)
	resumeRead <- true

	init_array = [3]byte{'\x24', '\x24', '\x24'}

	for {
		select {
		case b := <-d.readChan:
			scrolling[index%3] = b
			index++
			if scrolling == init_array {
				d.writeChan <- "b"
				resumeChan <- true
				return
			} 
		}
	}
}
