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
	"flag"
	"net/http"
	"time"
)

var addr = flag.String("addr", ":8888", "http service address")

const (
	channels         = 8
	samplesPerSecond = 250
	readTimeout      = 1000 * time.Millisecond
	readBufferSize   = 33 * samplesPerSecond
	baud             = 115200
	FFTSize          = samplesPerSecond
	RawMsgSize       = 20
)

var location string = "/dev/ttyUSB0"

func main() {
	h := NewHub()
	shutdown := make(chan bool, 1)
	defer func() {
		h.Close()
	}()
	mc := NewMindControl(h.broadcast, shutdown)
	handle := NewHandle(mc)
	http.HandleFunc("/ws", h.wsPacketHandler)
	http.HandleFunc("/", handle.rootHandler)
	http.HandleFunc("/x/", handle.commandHandler)
	http.HandleFunc("/open", handle.openHandler)
	http.HandleFunc("/reset", handle.resetHandler)
	http.HandleFunc("/start", handle.startHandler)
	http.HandleFunc("/stop", handle.stopHandler)
	http.HandleFunc("/close", handle.closeHandler)
	http.HandleFunc("/save", handle.saveHandler)
	http.HandleFunc("/test", handle.testHandler)
	http.HandleFunc("/js/", handle.jsHandler)
	http.HandleFunc("/static/", handle.cssHandler)
	http.HandleFunc("/bootstrap/", handle.bootstrapHandler)

	go h.Run()
	go mc.Start()
	run := func(shutdown <-chan bool) {
		go http.ListenAndServe(*addr, nil)
		for {
			select {
			case <-shutdown:
				return
			}
		}
	}
	run(shutdown)
}
