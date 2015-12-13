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
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/kevinjos/openbci-driver"
)

var (
	addr        = flag.String("addr", "", "http service address")
	location    = flag.String("loc", "", "serial mount point")
	baud        = flag.Int("baud", 115200, "serial baud rate")
	versionFlag = flag.Bool("version", false, "Print version info and exit.")
	readTimeout = time.Millisecond
	buildInfo   string
)

const (
	channels         = 8
	samplesPerSecond = 250
	readBufferSize   = 1024 * 1024
	RawMsgSize       = 30
)

func init() {
	flag.DurationVar(&readTimeout, "rt", 100*time.Millisecond, "serial readtimeout in milliseconds")
	flag.Parse()
	if *versionFlag {
		log.Printf("%s\n", buildInfo)
		os.Exit(0)
	}
}

func main() {
	h := NewHub()
	defer h.Close()

	var (
		device io.ReadWriteCloser
		err    error
	)
	if *location == "" {
		device = openbci.NewMockDevice()
	} else {
		device, err = openbci.NewDevice(*location, *baud, readTimeout)
		if err != nil {
			log.Fatalf("error opening device: %s\n", err)
		}
	}
	defer device.Close()

	shutdown := make(chan bool, 1)
	mc := NewMindControl(h.broadcast, shutdown, device)
	handle := NewHandle(mc)

	http.HandleFunc("/ws", h.wsPacketHandler)

	http.HandleFunc("/", handle.rootHandler)
	http.HandleFunc("/x/", handle.commandHandler)
	http.HandleFunc("/fft/", handle.fftHandler)
	http.HandleFunc("/reset", handle.resetHandler)
	http.HandleFunc("/start", handle.startHandler)
	http.HandleFunc("/stop", handle.stopHandler)
	http.HandleFunc("/close", handle.closeHandler)
	http.HandleFunc("/save", handle.saveHandler)
	http.HandleFunc("/js/", handle.jsHandler)
	http.HandleFunc("/static/", handle.cssHandler)
	http.HandleFunc("/bootstrap/", handle.bootstrapHandler)
	http.HandleFunc("/js/libs/", handle.libsHandler)

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
