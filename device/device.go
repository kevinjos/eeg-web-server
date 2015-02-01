package device

import (
	"fmt"
	"github.com/tarm/goserial"
	"io"
	"os"
	"time"
)

type Device struct {
	Location     string
	Baud         int
	ReadTimeout  time.Duration
	WriteStream  chan string
	byteStreamIn chan byte
	quit         chan bool
	Conn         io.ReadWriteCloser
}

func New(Location string, Baud int, ReadTimeout time.Duration, WriteStream chan string) *Device {
	byteStreamIn := make(chan byte)
	quit := make(chan bool)
	d := &Device{
		Location:     Location,
		Baud:         Baud,
		ReadTimeout:  ReadTimeout,
		WriteStream:  WriteStream,
		byteStreamIn: byteStreamIn,
		quit:         quit,
	}
	return d
}

func (d *Device) Read() {
	buf := make([]byte, 1)
	for {
		n, err := d.Conn.Read(buf)
		if err != nil {
			fmt.Println("Error reading [", n, "] bytes from serial device: [", err, "]")
		} else if n > 0 {
			for i := 0; i < n; i++ {
				d.byteStreamIn <- buf[i]
			}
		}
	}
}

func (d *Device) Write(s string) {
	wb := []byte(s)
	n, err := d.Conn.Write(wb)
	if err != nil {
		fmt.Println("Error writing [", n, "] bytes to serial device: [", err, "]")
	} else if n > 0 {
		fmt.Println("Wrote [", n, "] byte", wb, "to the serial device")
	}
}

func (d *Device) Stream(byteStream chan byte) {
	go d.Read()
	for {
		select {
		case b := <-d.byteStreamIn:
			byteStream <- b
		case s := <-d.WriteStream:
			d.Write(s)
		case <-d.quit:
			fmt.Println("The stream stops flowing")
			return
		}
	}
}

func (d *Device) Open() {
	config := &serial.Config{Name: d.Location, Baud: d.Baud, ReadTimeout: d.ReadTimeout}
	conn, err := serial.OpenPort(config)
	if err != nil {
		fmt.Println("Error conneting to serial device: [", err, "]")
		os.Exit(1)
	}
	d.Conn = conn
	d.Write("v")
	time.Sleep(time.Millisecond * 1000)
}

func (d *Device) Close() {
	d.WriteStream <- "s"
	d.quit <- true
	time.Sleep(time.Millisecond * 1000)
	d.Conn.Close()
	fmt.Println("Safely closed the device")
}
