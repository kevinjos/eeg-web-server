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
	WriteStream chan string
	ByteStream  chan byte
	reset       chan bool
	Quit        chan bool
	Conn        io.ReadWriteCloser
}

func NewDevice(Location string, Baud int, ReadTimeout time.Duration,
	ByteStream chan byte, WriteStream chan string, Quit chan bool) *Device {
	resetChan := make(chan bool)
	d := &Device{
		Location:    Location,
		Baud:        Baud,
		ReadTimeout: ReadTimeout,
		ByteStream:  ByteStream,
		WriteStream: WriteStream,
		Quit:        Quit,
		reset:       resetChan,
	}
	return d
}

func (d *Device) ReadWriteClose() {
	streamingData := false
	buf := make([]byte, 1)
	for {
		select {
		case s := <-d.WriteStream:
			d.Write(s)
			switch {
			case s == "s":
				streamingData = false
			case s == "b":
				streamingData = true
			}
		case <-d.Quit:
			defer func() {
				d.Write("s")
				d.Conn.Close()
				fmt.Println("Safely closed the device")
				os.Exit(1)
			}()
			return
		case <-d.reset:
			d.Read(buf)
		default:
			if streamingData == true {
				d.Read(buf)
			}
		}
	}
}

func (d *Device) Read(buf []byte) {
	n, err := d.Conn.Read(buf)
	if err != nil {
		fmt.Println("Error reading [", n, "] bytes from serial device: [", err, "]")
	} else if n > 0 {
		for i := 0; i < n; i++ {
			d.ByteStream <- buf[i]
		}
	}
}

func (d *Device) Write(s string) {
	wb := []byte(s)
	n, err := d.Conn.Write(wb)
	time.Sleep(600 * time.Millisecond)
	if err != nil {
		fmt.Println("Error writing [", n, "] bytes to serial device: [", err, "]")
	} else if n > 0 {
		fmt.Println("Wrote [", n, "] byte", wb, "to the serial device")
	}
	return
}

func (d *Device) Open() {
	config := &serial.Config{Name: d.Location, Baud: d.Baud, ReadTimeout: d.ReadTimeout}
	conn, err := serial.OpenPort(config)
	if err != nil {
		fmt.Println("Error conneting to serial device: [", err, "]")
		os.Exit(1)
	}
	d.Conn = conn
	go d.ReadWriteClose()
	d.Reset()
}

//Reset sends the reset message to the serial device,
//waits one seconds and then reads up to the init
//message [$$$]. Should find a better way than firing
//off a read go routine every time.
func (d *Device) Reset() {
	var (
		scrolling  [3]byte
		init_array [3]byte
		index      int
	)

	init_array = [3]byte{'\x24', '\x24', '\x24'}

	d.WriteStream <- "s"
	d.WriteStream <- "v"
	d.reset <- true

	for {
		b := <-d.ByteStream
		fmt.Print(string(b))
		scrolling[index%3] = b
		index++
		if scrolling == init_array {
			return
		} else {
			d.reset <- true
		}
	}
}
