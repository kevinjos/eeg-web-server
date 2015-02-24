/*  OpenBCI golang server allows users to control, visualize and store data
    collected from the OpenBCI microcontroller.
    Copyright (C) 2015  Kevin Schiesser

    This program is free software: you can redistribute it and/or modify
    it under the terms of the GNU Affero General Public License as
    published by the Free Software Foundation, either version 3 of the
    License, or (at your option) any later version.

    This program is distributed in the hope that it will be useful,
    but WITHOUT ANY WARRANTY; without even the implied warranty of
    MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
    GNU Affero General Public License for more details.

    You should have received a copy of the GNU Affero General Public License
    along with this program.  If not, see <http://www.gnu.org/licenses/>.
*/

package main

import (
	"github.com/pkg/term"
	"io"
	"log"
	"time"
)

type OpenBCI struct {
	writeChan     chan string
	readChan      chan byte
	timeoutChan   chan bool
	resetChan     chan chan bool
	pauseReadChan chan chan bool
	quitCommand   chan bool
	quitRead      chan bool
	conn          io.ReadWriteCloser
}

func NewOpenBCI() *OpenBCI {
	return &OpenBCI{
		writeChan:     make(chan string, 64),
		readChan:      make(chan byte, readBufferSize),
		timeoutChan:   make(chan bool),
		quitCommand:   make(chan bool),
		quitRead:      make(chan bool),
		resetChan:     make(chan chan bool),
		pauseReadChan: make(chan chan bool),
	}
}

func (d *OpenBCI) Close() {
	d.write("s")
	d.quitCommand <- true
	if d.conn != nil {
		d.quitRead <- true
		d.conn.Close()
		log.Println("Safely closed the device")
	}
	close(d.timeoutChan)
	close(d.resetChan)
	close(d.pauseReadChan)
	close(d.readChan)
	close(d.writeChan)
}

func (d *OpenBCI) command() {
	for {
		select {
		case s := <-d.writeChan:
			d.write(s)
		case resumePacketChan := <-d.resetChan:
			go d.reset(resumePacketChan)
		case <-d.quitCommand:
			return
		}
	}
}

func (d *OpenBCI) read() {
  bufSize := 8
	buf := make([]byte, bufSize)
  readChan := make(chan byte, bufSize)
	for {
		select {
		case resumeReadChan := <-d.pauseReadChan:
			<-resumeReadChan
		case <-d.quitRead:
			return
		default:
      if d.conn != nil {
        go func() { 
            n, err := d.conn.Read(buf)
            if err == io.EOF {
              d.timeoutChan <- true
            } else if err != nil {
              log.Fatal("Error reading from serial device: [", err, "]")
            }
            for i := 0; i < n; i++ {
              readChan <- buf[i]
            }
        }()
      }
      select {
        case <-time.After(readTimeout):
          d.timeoutChan <- true
        case b := <-readChan:
          d.readChan <- b
      }
    }
  }
}

func (d *OpenBCI) write(s string) {
	wb := []byte(s)
	if d.conn != nil {
		if n, err := d.conn.Write(wb); err != nil {
			log.Println("Error writing [", n, "] bytes to serial device: [", err, "]")
		} else {
			log.Println("Wrote [", n, "] byte", wb, "to the serial device")
		}
	}
}

func (d *OpenBCI) open() {
	//config := &serial.Config{Name: location, Baud: baud, ReadTimeout: readTimeout}
	//conn, err := serial.OpenPort(config)
  conn, err := term.Open(location, term.Speed(baud))
	if err != nil {
		log.Fatal("Error conneting to serial device at [", location, "]: [", err, "]")
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
