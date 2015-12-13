/*OpenBCI golang server allows users to control, visualize and store data
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
	"io"
	"os"
	"strconv"
	"time"

	"github.com/golang/glog"
	"github.com/kevinjos/goedf"
	"github.com/kevinjos/gofidlib"
)

// MindControl ...
type MindControl struct {
	SerialDevice     io.ReadWriteCloser
	PacketChan       chan *Packet
	savePacketChan   chan *Packet
	deltaFFT         chan [2]int
	quitGenTest      chan bool
	quitSendPackets  chan bool
	quitSave         chan bool
	quitDecodeStream chan bool
	pauseRead        chan chan bool
	gainC            chan *[8]float64
	shutdown         chan bool
	broadcast        chan *message
	gain             [8]float64
	saving           bool
	genTesting       bool
}

// NewMindControl ...
func NewMindControl(broadcast chan *message, shutdown chan bool, device io.ReadWriteCloser) *MindControl {
	//Set up the serial device
	return &MindControl{
		SerialDevice:     device,
		PacketChan:       make(chan *Packet),
		savePacketChan:   make(chan *Packet),
		deltaFFT:         make(chan [2]int),
		quitGenTest:      make(chan bool),
		quitSendPackets:  make(chan bool),
		quitSave:         make(chan bool),
		quitDecodeStream: make(chan bool),
		pauseRead:        make(chan chan bool),
		gainC:            make(chan *[8]float64),
		shutdown:         shutdown,
		broadcast:        broadcast,
		gain:             [8]float64{24.0, 24.0, 24.0, 24.0, 24.0, 24.0, 24.0, 24.0},
		saving:           false,
		genTesting:       false,
	}
}

// Start necessary go routines
func (mc *MindControl) Start() {
	go DecodeStream(mc.PacketChan, mc.gainC, mc.quitDecodeStream, mc.pauseRead, mc.SerialDevice)
	go mc.sendPackets()
}

// Close go routines and channels started by MindControl
func (mc *MindControl) Close() {
	if mc.saving {
		mc.quitSave <- true
	}
	mc.SerialDevice.Close()
	mc.quitDecodeStream <- true
	close(mc.quitSendPackets)
	close(mc.quitGenTest)
	close(mc.shutdown)
}

func (mc *MindControl) saveBDF() {
	var ns int
	wd, err := os.Getwd()
	if err != nil {
		glog.Errorln(err)
		return
	}
	wd += "/data/"
	files := make([]*os.File, channels)
	tmpdir := wd + strconv.FormatInt(time.Now().Unix(), 10)
	err = os.MkdirAll(tmpdir, 0777)
	if err != nil {
		glog.Errorln(err)
		return
	}
	for i := 0; i < channels; i++ {
		fn := "chan" + strconv.Itoa(i)
		file, err := os.Create(tmpdir + "/" + fn)
		files[i] = file
		if err != nil {
			glog.Errorln(err)
			return
		}
	}
	defer func() {
		mc.saving = false
		for _, f := range files {
			f.Close()
		}
		err = os.RemoveAll(tmpdir)
		if err != nil {
			glog.Errorln(err)
		}
	}()
	startts := time.Now()
	// crunch know EDF header quantities
	version := "\xffBIOSEMI"
	LRID := "Startdate " + startts.Format("02-JAN-2006")
	startdate := startts.Format("02.01.06")
	starttime := startts.Format("15.04.05")
	numbytes := strconv.Itoa(biosigio.FixedHeaderBytes + biosigio.VariableHeaderBytes*channels)
	reserved := "24BIT"
	numsignals := strconv.Itoa(channels)
	numdatar := "1"
	phydims := []string{"uv", "uv", "uv", "uv", "uv", "uv", "uv", "uv"}
	phymins := make([]string, channels)
	phymaxs := make([]string, channels)
	nsreserved := make([]string, channels)
	for idx, val := range mc.gain {
		phymins[idx] = strconv.FormatFloat(scaleToMicroVolts(-8388608, val), 'f', 0, 64)
		phymaxs[idx] = strconv.FormatFloat(scaleToMicroVolts(8388607, val), 'f', 0, 64)
		nsreserved[idx] = "3"
	}
	digmins := []string{"-8388608", "-8388608", "-8388608", "-8388608",
		"-8388608", "-8388608", "-8388608", "-8388608"}
	digmaxs := []string{"8388607", "8388607", "8388607", "8388607",
		"8388607", "8388607", "8388607", "8388607"}
	for {
		select {
		case p := <-mc.savePacketChan:
			ns++
			val := []byte{p.Rchan1[2], p.Rchan1[1], p.Rchan1[0]}
			files[0].Write(val)
			val = []byte{p.Rchan2[2], p.Rchan2[1], p.Rchan2[0]}
			files[1].Write(val)
			val = []byte{p.Rchan3[2], p.Rchan3[1], p.Rchan3[0]}
			files[2].Write(val)
			val = []byte{p.Rchan4[2], p.Rchan4[1], p.Rchan4[0]}
			files[3].Write(val)
			val = []byte{p.Rchan5[2], p.Rchan5[1], p.Rchan5[0]}
			files[4].Write(val)
			val = []byte{p.Rchan6[2], p.Rchan6[1], p.Rchan6[0]}
			files[5].Write(val)
			val = []byte{p.Rchan7[2], p.Rchan7[1], p.Rchan7[0]}
			files[6].Write(val)
			val = []byte{p.Rchan8[2], p.Rchan8[1], p.Rchan8[0]}
			files[7].Write(val)
		case <-mc.quitSave:
			endts := time.Now()
			duration := strconv.FormatFloat(endts.Sub(startts).Seconds(), 'f', 3, 64)
			numsamples := make([]string, channels)
			for idx := range numsamples {
				numsamples[idx] = strconv.Itoa(ns)
			}
			h, err := biosigio.NewHeader(biosigio.Version(version),
				biosigio.LocalRecordID(LRID),
				biosigio.Startdate(startdate),
				biosigio.Starttime(starttime),
				biosigio.NumBytes(numbytes),
				biosigio.Reserved(reserved),
				biosigio.NumDataRecord(numdatar),
				biosigio.Duration(duration),
				biosigio.NumSignal(numsignals),
				biosigio.PhysicalDimensions(phydims),
				biosigio.PhysicalMaxs(phymaxs),
				biosigio.PhysicalMins(phymins),
				biosigio.DigitalMaxs(digmaxs),
				biosigio.DigitalMins(digmins),
				biosigio.NumSamples(numsamples),
				biosigio.NSReserved(nsreserved))
			if err != nil {
				glog.Errorln(err)
			}
			bdf := biosigio.NewBDF(h, []*biosigio.BDFData{})
			buf, err := biosigio.MarshalBDF(bdf)
			if err != nil {
				glog.Errorln(err)
			}
			outfn := wd + strconv.FormatInt(endts.Unix(), 10) + ".edf"
			outfd, err := os.Create(outfn)
			if err != nil {
				glog.Errorln(err)
			}
			defer outfd.Close()
			_, err = outfd.Write(buf)
			if err != nil {
				glog.Errorln(err)
			}
			for _, fd := range files {
				_, err = fd.Seek(0, 0)
				if err != nil {
					glog.Errorln(err)
				}
				_, err := io.Copy(outfd, fd)
				if err != nil {
					glog.Errorln(err)
				}
			}
			return
		}
	}
}

func (mc *MindControl) sendPackets() {
	var i int

	FFTSize := 250
	FFTFreq := 50

	filterDesign, err := gofidlib.NewFilterDesign("BpBe4/1-30", samplesPerSecond)
	if err != nil {
		glog.Fatal("Error creating filter design:", err)
	}

	filter := make([]*gofidlib.Filter, 8)
	for j := 0; j < 8; j++ {
		filter[j] = gofidlib.NewFilter(filterDesign)
	}

	defer func() {
		filterDesign.Free()
		for j := 0; j < 8; j++ {
			filter[j].Free()
		}
	}()

	pbFFT := NewPacketBatcher(FFTSize)
	pbRaw := NewPacketBatcher(RawMsgSize)

	for {
		select {
		case <-mc.quitSendPackets:
			return
		case arr := <-mc.deltaFFT:
			FFTSize = arr[0]
			FFTFreq = arr[1]
			pbFFT = NewPacketBatcher(FFTSize)
			i = 0
		case p := <-mc.PacketChan:
			if mc.saving == true {
				mc.savePacketChan <- p
			}

			p.Chan1 = filter[0].Run(p.Chan1)
			p.Chan2 = filter[1].Run(p.Chan2)
			p.Chan3 = filter[2].Run(p.Chan3)
			p.Chan4 = filter[3].Run(p.Chan4)
			p.Chan5 = filter[4].Run(p.Chan5)
			p.Chan6 = filter[5].Run(p.Chan6)
			p.Chan7 = filter[6].Run(p.Chan7)
			p.Chan8 = filter[7].Run(p.Chan8)

			pbFFT.packets[i%FFTSize] = p
			pbRaw.packets[i%RawMsgSize] = p

			if i%RawMsgSize == RawMsgSize-1 {
				pbRaw.batch()
				mc.broadcast <- newMessage("raw", pbRaw.Chans)
			}

			if i > FFTSize && i%FFTFreq == FFTFreq-1 {
				pbFFT.batch()
				pbFFT.setFFT()
				mc.broadcast <- newMessage("fft", pbFFT.FFTs)
				binMsg := make(map[string][]float64)
				binMsg["fftBins"] = calcFFTBins(FFTSize)
				mc.broadcast <- newMessage("fftBins", binMsg)
			}

			i++

		}
	}
}

type message struct {
	Name    string
	Payload map[string][]float64
}

func newMessage(name string, payload map[string][]float64) *message {
	return &message{
		Name:    name,
		Payload: payload,
	}
}
