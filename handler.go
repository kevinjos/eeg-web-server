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
	"fmt"
	"html/template"
	"net/http"
	"strconv"
	"strings"

	"github.com/kevinjos/openbci-driver"
)

type Handle struct {
	mc *MindControl
}

func NewHandle(mindcontrol *MindControl) *Handle {
	return &Handle{
		mc: mindcontrol,
	}
}

func (handle *Handle) parseCommand(path string) string {
	var command string
	channelOn := map[string]string{"1": "!", "2": "@", "3": "#", "4": "$", "5": "%", "6": "^", "7": "&", "8": "*"}
	gainMap := map[string]float64{"0": 1.0, "1": 2.0, "2": 4.0, "3": 6.0, "4": 8.0, "5": 12.0, "6": 24.0}
	p := strings.Split(path, "/")
	channel := p[2]
	switch {
	case len(p) < 4:
		command = ""
	case p[3] == "true":
		command = channelOn[channel]
	case p[3] == "false":
		command = channel
	case channel == "0": //send command to all channels
		c := make([]string, 8)
		for i := 0; i < 8; i++ {
			c[i] = p[3][0:1] + strconv.Itoa(i+1) + p[3][2:]
			handle.mc.gain[i] = gainMap[p[3][3:4]]
		}
		command = c[0] + c[1] + c[2] + c[3] + c[4] + c[5] + c[6] + c[7]
	case p[3][0:1] == "x":
		ci := channel[0] - 49
		handle.mc.gain[ci] = gainMap[p[3][3:4]]
		command = p[3]
	}
	return command
}

func (handle *Handle) jsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", 405)
		return
	}
	p := strings.Split(r.URL.Path, "/")
	f := p[len(p)-1]
	http.ServeFile(w, r, "js/"+f)
}

func (handle *Handle) rootHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.Error(w, "Not found", 404)
		return
	}
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", 405)
		return
	}
	rootTempl := template.Must(template.ParseFiles("static/index.html"))
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	rootTempl.Execute(w, r.Host)
}

func (handle *Handle) cssHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", 405)
		return
	}
	p := strings.Split(r.URL.Path, "/")
	f := p[len(p)-1]
	http.ServeFile(w, r, "static/"+f)
}

func (handle *Handle) bootstrapHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", 405)
		return
	}
	p := strings.Split(r.URL.Path, "/")
	t := p[len(p)-2]
	f := p[len(p)-1]
	http.ServeFile(w, r, "bootstrap/"+t+"/"+f)
}

func (handle *Handle) libsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", 405)
		return
	}
	p := strings.Split(r.URL.Path, "/")
	f := p[len(p)-1]
	http.ServeFile(w, r, "js/libs/"+f)
}

func (handle *Handle) commandHandler(w http.ResponseWriter, r *http.Request) {
	if strings.Contains(r.URL.Path, "/x/") == false {
		http.Error(w, "Not found", 404)
		return
	}
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", 405)
		return
	}
	command := handle.parseCommand(r.URL.Path)
	lenCommand := len(command)
	if lenCommand > 72 {
		http.Error(w, "Method not allowed", 405)
		return
	}

	for i := 0; i < lenCommand; i++ {
		if lenCommand < 9 {
			handle.mc.SerialDevice.Write([]byte{command[i]})
		} else if i%9 == 0 && i+9 < lenCommand {
			handle.staggerWriter(command[i : i+9])
		} else if i%9 == 0 {
			handle.staggerWriter(command[i:])
		}
	}
}

func (handle *Handle) staggerWriter(c string) {
	handle.mc.SerialDevice.Write([]byte{c[0]})
	handle.mc.SerialDevice.Write([]byte{c[1], c[2], c[3], c[4], c[5], c[6], c[7]})
	handle.mc.SerialDevice.Write([]byte{c[8]})
}

func (handle *Handle) closeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", 405)
		return
	}
	handle.mc.Close()
}

func (handle *Handle) startHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", 405)
		return
	}
	handle.mc.SerialDevice.Write(openbci.Command["start"])
}

func (handle *Handle) stopHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", 405)
		return
	}
	handle.mc.SerialDevice.Write(openbci.Command["stop"])
}

func (handle *Handle) saveHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", 405)
		return
	}
	handle.mc.saving = handle.mc.saving != true
	if handle.mc.saving == true {
		go handle.mc.saveBDF()
	} else {
		handle.mc.quitSave <- true
	}
}

func (handle *Handle) resetHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", 405)
		return
	}
	resume := make(chan bool)
	handle.mc.pauseRead <- resume
	_, err := handle.mc.SerialDevice.Write(openbci.Command["reset"])
	if err != nil {
		fmt.Printf("error reseting device: %s\n", err)
	}
	resume <- true
}

func (handle *Handle) testHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", 405)
		return
	}
	handle.mc.genTesting = !handle.mc.genTesting
	if handle.mc.genTesting == false {
		handle.mc.quitGenTest <- true
	} else {
		go genTestPackets(handle.mc.PacketChan, handle.mc.quitGenTest)
	}
}

func (handle *Handle) fftHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", 405)
		return
	}
	path := r.URL.Path
	p := strings.Split(path, "/")
	data := strings.Split(p[2], "&")
	fftsize, err := strconv.Atoi(data[0])
	if err != nil {
		http.Error(w, "Bad Request, only integers understood", 400)
		return
	}
	fftfreq, err := strconv.Atoi(data[1])
	if err != nil {
		http.Error(w, "Bad Request, only integers understood", 400)
		return
	}
	handle.mc.deltaFFT <- [2]int{fftsize, fftfreq}
}
