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
	"bytes"
	"github.com/kevinjos/gofidlib"
	"log"
	"math"
	"os"
	"strconv"
	"time"
)

// MindControl ...
type MindControl struct {
	ReadChan         *chan byte
	PacketChan       chan *Packet
	savePacketChan   chan *Packet
	deltaFFT         chan [2]int
	ResetChan        chan bool
	genToggleChan    chan bool
	quitGenTest      chan bool
	quitDecodeStream chan bool
	quitSendPackets  chan bool
	quitSave         chan bool
	shutdown         chan bool
	SerialDevice     *OpenBCI
	broadcast        chan *Message
	gain             [8]float64
	gainC            chan *[8]float64
	saving           bool
}

// NewMindControl ...
func NewMindControl(broadcast chan *Message, shutdown chan bool) *MindControl {
	//Set up the serial device
	serialDevice := NewOpenBCI()
	return &MindControl{
		ReadChan:         &serialDevice.readChan,
		PacketChan:       make(chan *Packet),
		savePacketChan:   make(chan *Packet, 2),
		deltaFFT:         make(chan [2]int),
		ResetChan:        make(chan bool),
		quitGenTest:      make(chan bool),
		quitDecodeStream: make(chan bool),
		quitSendPackets:  make(chan bool),
		quitSave:         make(chan bool),
		genToggleChan:    make(chan bool),
		shutdown:         shutdown,
		SerialDevice:     serialDevice,
		gain:             [8]float64{24.0, 24.0, 24.0, 24.0, 24.0, 24.0, 24.0, 24.0},
		broadcast:        broadcast,
		saving:           false,
	}
}

// Start ...
func (mc *MindControl) Start() {
	go mc.sendPackets()
	go DecodeStream(mc.quitDecodeStream, mc.ReadChan, mc.SerialDevice.readChan, mc.ResetChan,
		mc.SerialDevice.resetChan, mc.SerialDevice.timeoutChan, mc.PacketChan, mc.gainC)
	go mc.SerialDevice.command()
	go mc.GenTestPackets()
	mc.genToggleChan <- false
}

// Open ...
func (mc *MindControl) Open() {
	mc.SerialDevice.open()
	buf := make([]byte, readBufferSize)
	go mc.SerialDevice.read(buf)
	mc.ResetChan <- true
}

// Close ...
func (mc *MindControl) Close() {
	if mc.saving {
		mc.quitSave <- true
	}
	mc.SerialDevice.Close()
	mc.quitDecodeStream <- true
	mc.quitSendPackets <- true
	close(mc.quitGenTest)
	close(mc.PacketChan)
	close(mc.ResetChan)
	close(mc.genToggleChan)
	mc.shutdown <- true
}

func openFile() (*os.File, error) {
	wd, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	fn := time.Now().String()
	file, err := os.Create(wd + "/data/" + fn)
	if err != nil {
		return nil, err
	}
	return file, nil
}

func openTmpFiles(n int) (files []*os.File, err error) {
	files = make([]*os.File, n)
	wd, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	tmpdir := strconv.FormatInt(time.Now().Unix(), 10)
	err = os.MkdirAll(wd+"/data/"+tmpdir, 0777)
	if err != nil {
		return nil, err
	}
	for i := 0; i < n; i++ {
		fn := "chan" + strconv.Itoa(i)
		file, err := os.Create(wd + "/data/" + tmpdir + "/" + fn)
		files[i] = file
		if err != nil {
			return nil, err
		}
	}
	return files, nil
}

func packetToCSV(startTime int64, p *Packet) []byte {
	timeDiff := time.Now().UnixNano() - startTime
	row := bytes.NewBufferString(strconv.FormatInt(timeDiff, 10) + "," +
		strconv.FormatBool(p.Synced) + "," +
		strconv.FormatFloat(p.Chan1, 'G', 8, 64) + "," +
		strconv.FormatFloat(p.Chan2, 'G', 8, 64) + "," +
		strconv.FormatFloat(p.Chan3, 'G', 8, 64) + "," +
		strconv.FormatFloat(p.Chan4, 'G', 8, 64) + "," +
		strconv.FormatFloat(p.Chan5, 'G', 8, 64) + "," +
		strconv.FormatFloat(p.Chan6, 'G', 8, 64) + "," +
		strconv.FormatFloat(p.Chan7, 'G', 8, 64) + "," +
		strconv.FormatFloat(p.Chan8, 'G', 8, 64) + "," +
		strconv.FormatInt(int64(p.AccX), 10) + "," +
		strconv.FormatInt(int64(p.AccY), 10) + "," +
		strconv.FormatInt(int64(p.AccZ), 10) + "\n")
	return row.Bytes()
}

func (mc *MindControl) saveBDF() {
	files, err := openTmpFiles(8)
	if err != nil {
		log.Println(err)
		return
	}
	defer func() {
		mc.saving = false
		for _, f := range files {
			f.Close()
		}
	}()
	for {
		select {
		case p := <-mc.savePacketChan:
			files[0].Write(p.Rchan1)
			files[1].Write(p.Rchan2)
			files[2].Write(p.Rchan3)
			files[3].Write(p.Rchan4)
			files[4].Write(p.Rchan5)
			files[5].Write(p.Rchan6)
			files[6].Write(p.Rchan7)
			files[7].Write(p.Rchan8)
		case <-mc.quitSave:
			return
		}
	}
}

func (mc *MindControl) save() {
	var file *os.File
	fileState := 1
	startTime := time.Now().UnixNano()
	for {
		select {
		case p := <-mc.savePacketChan:
			switch fileState {
			case 1:
				file, _ = openFile()
				defer func() {
					file.Close()
					mc.saving = false
				}()
				header := bytes.NewBufferString(`NanoSec,Synced,Chan1,Chan2,Chan3,Chan4,Chan5,Chan6,Chan7,Chan8,AccX,AccY,AccZ
`)
				_, err := file.Write(header.Bytes())
				if err != nil {
					log.Println(err)
					return
				}
				fileState++
				fallthrough
			case 2:
				row := packetToCSV(startTime, p)
				_, err := file.Write(row)
				if err != nil {
					log.Println(err)
					return
				}
			}
		case <-mc.quitSave:
			return
		}
	}
}

// GenTestPackets ...
func (mc *MindControl) GenTestPackets() {
	var on bool
	var val float64
	var i float64
	for {
		select {
		case <-mc.quitGenTest:
			return
		case on = <-mc.genToggleChan:
			on = <-mc.genToggleChan
		default:
			if on {
				i = i + 0.04
				val = 0.1*math.Sin(2.0*math.Pi*i) + 0.1*math.Cos(2.0*math.Pi*0.2*i)
				packet := NewPacket()
				packet.Chan1 = val
				packet.Chan2 = val
				packet.Chan3 = val
				packet.Chan4 = val
				packet.Chan5 = val
				packet.Chan6 = val
				packet.Chan7 = val
				packet.Chan8 = val
				mc.PacketChan <- packet
				time.Sleep(4 * time.Millisecond)
			}
		}
	}
}

// Message ...
type Message struct {
	Name    string
	Payload map[string][]float64
}

// NewMessage ...
func NewMessage(name string, payload map[string][]float64) *Message {
	return &Message{
		Name:    name,
		Payload: payload,
	}

}

// CalcFFTBins ...
func CalcFFTBins(fftSize int) (bins []float64) {
	bins = make([]float64, fftSize/2)
	step := float64(samplesPerSecond) / float64(fftSize)
	for idx := range bins {
		bins[idx] = step * float64(idx)
	}
	return bins
}

func (mc *MindControl) sendPackets() {
	var (
		i int
	)

	FFTSize := 250
	FFTFreq := 50

	filterDesign, err := gofidlib.NewFilterDesign("BpBe4/1-50", samplesPerSecond)
	if err != nil {
		log.Fatal("Error creating filter design:", err)
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
				mc.broadcast <- NewMessage("raw", pbRaw.Chans)
			}

			if i > FFTSize && i%FFTFreq == FFTFreq-1 {
				pbFFT.batch()
				pbFFT.setFFT()
				mc.broadcast <- NewMessage("fft", pbFFT.FFTs)
				binMsg := make(map[string][]float64)
				binMsg["fftBins"] = CalcFFTBins(FFTSize)
				mc.broadcast <- NewMessage("fftBins", binMsg)
			}

			i++

		}
	}
}
