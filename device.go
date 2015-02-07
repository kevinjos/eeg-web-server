package main

import (
	"fmt"
	"github.com/tarm/goserial"
	"io"
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
	buf := make([]byte, 1)
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
		case call := <-d.resetButton:
			if call == true {
				go d.reset()
			} else {
				d.read(buf)
			}
		case <-d.quitButton:
			defer func() {
				d.write("s")
				d.conn.Close()
				fmt.Println("Safely closed the device")
				os.Exit(1)
			}()
			return
		default:
			switch {
			case streamingData == true:
				d.read(buf)
			}
		}
	}
}

func (d *Device) read(buf []byte) {
	n, err := d.conn.Read(buf)
	if err != nil {
		fmt.Println("Error reading [", n, "] bytes from serial device: [", err, "]")
	} else if n > 0 {
		for i := 0; i < n; i++ {
			d.byteStream <- buf[i]
		}
	}
}

func (d *Device) write(s string) {
	wb := []byte(s)
	n, err := d.conn.Write(wb)
	time.Sleep(1000 * time.Millisecond)
	if err != nil {
		fmt.Println("Error writing [", n, "] bytes to serial device: [", err, "]")
	} else if n > 0 {
		fmt.Println("Wrote [", n, "] byte", wb, "to the serial device")
	}
	return
}

func (d *Device) open() {
	config := &serial.Config{Name: d.Location, Baud: d.Baud, ReadTimeout: d.ReadTimeout}
	conn, err := serial.OpenPort(config)
	if err != nil {
		fmt.Println("Error conneting to serial device: [", err, "]")
		os.Exit(1)
	}
	d.conn = conn
	go d.readWriteClose()
	d.reset()
}

//Reset sends the reset message to the serial device,
//waits one seconds and then reads up to the init
//message [$$$]. Should find a better way than firing
//off a read go routine every time.
//TODO: figure out why cpu usage soars during reset routine
func (d *Device) reset() {
	var (
		scrolling  [3]byte
		init_array [3]byte
		index      int
	)

	init_array = [3]byte{'\x24', '\x24', '\x24'}

	d.writeStream <- "s"
	time.Sleep(1000 * time.Millisecond)
	d.writeStream <- "v"
	time.Sleep(1000 * time.Millisecond)

	for {
		select {
		case b := <-d.byteStream:
			if b < 123 {
				fmt.Print(string(b))
				scrolling[index%3] = b
				index++
			}
		default:
			if scrolling == init_array {
				fmt.Print("\n")
				d.writeStream <- "b"
				return
			} else {
				d.resetButton <- false
			}
		}
	}
}
