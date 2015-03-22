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
	"bytes"
	"github.com/kevinjos/gofidlib"
	"log"
	"math"
	"os"
	"strconv"
	"time"
)

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
	saving           bool
}

func NewMindControl(broadcast chan *Message, shutdown chan bool) *MindControl {
	//Set up the serial device
	serialDevice := NewOpenBCI()
	return &MindControl{
		ReadChan:         &serialDevice.readChan,
		PacketChan:       make(chan *Packet),
		savePacketChan:   make(chan *Packet),
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

func (mc *MindControl) Start() {
	go mc.sendPackets()
	go mc.DecodeStream()
	go mc.SerialDevice.command()
	go mc.GenTestPackets()
	mc.genToggleChan <- false
}

func (mc *MindControl) Open() {
	mc.SerialDevice.open()
	buf := make([]byte, readBufferSize)
	go mc.SerialDevice.read(buf)
	mc.ResetChan <- true
}

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

//decodeStream implements the openbci packet protocol to
//assemble packets and sends packet arrays onto the packetStream
func (mc *MindControl) DecodeStream() {
	var (
		b             uint8
		readstate     uint8
		thisPacket    [33]byte
		lastPacket    [33]byte
		seqDiff       uint8
		syncPktCtr    uint8
		syncPktThresh uint8
	)
	resetMonitorChan := make(chan bool)
	syncPktThresh = 2
	for {
		select {
		case <-mc.quitDecodeStream:
			return
		case <-resetMonitorChan:
			mc.ReadChan = &mc.SerialDevice.readChan
		case <-mc.ResetChan:
			readstate, syncPktCtr = 0, 0
			var bogusChan chan byte
			mc.ReadChan = &bogusChan
			mc.SerialDevice.resetChan <- resetMonitorChan
		case <-mc.SerialDevice.timeoutChan:
			lastPacket := lastPacket
			mc.PacketChan <- encodePacket(&lastPacket, 0, &mc.gain, false)
		case b = <-*mc.ReadChan:
			switch readstate {
			case 0:
				if b == '\xc0' {
					readstate++
				}
			case 1:
				if b == '\xa0' {
					thisPacket[0] = b
					readstate++
				} else {
					readstate = 0
				}
			case 2:
				thisPacket[1] = b
				seqDiff = difference(b, lastPacket[1])
				if seqDiff > 1 && syncPktCtr > syncPktThresh {
					log.Println(seqDiff, "packets behind. Got:", b, "Expected:", lastPacket[1])
				}
				for seqDiff > 1 {
					lastPacket[1]++
					time.Sleep(4 * time.Millisecond)
					mc.PacketChan <- encodePacket(&lastPacket, 100-seqDiff, &mc.gain, false)
					seqDiff--
				}
				fallthrough
			case 3:
				for j := 2; j < 32; j++ {
					thisPacket[j] = <-*mc.ReadChan
				}
				readstate = 4
			case 4:
				switch {
				case b == '\xc0':
					thisPacket[32] = b
					lastPacket = thisPacket
					if syncPktCtr > syncPktThresh {
						mc.PacketChan <- encodePacket(&thisPacket, 100, &mc.gain, true)
					} else {
						syncPktCtr++
					}
					readstate = 1
				case b != '\xc0':
					readstate = 0
					fallthrough
				case syncPktCtr > syncPktThresh:
					log.Println("Footer out of sync")
				}
			}
		}
	}
}

func (mc *MindControl) GenTestPackets() {
	var on bool
	var val float64
	var i float64 = 0.0
	for {
		select {
		case <-mc.quitGenTest:
			return
		case on = <-mc.genToggleChan:
			on = <-mc.genToggleChan
		default:
			if on {
				i = i + 0.04
				val = math.Sin(2.0 * math.Pi * i)
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

type Message struct {
	Name    string
	Payload map[string][]float64
}

func NewMessage(name string, payload map[string][]float64) *Message {
	return &Message{
		Name:    name,
		Payload: payload,
	}

}

func (mc *MindControl) sendPackets() {
	var (
		i int
	)

	FFTSize := 250
	FFTFreq := 50

	last_second := time.Now().UnixNano()
	second := time.Now().UnixNano()

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
				//pbRaw = NewPacketBatcher(RawMsgSize)
			}

			if i > FFTSize && i%FFTFreq == FFTFreq-1 {
				pbFFT.batch()
				pbFFT.setFFT()
				mc.broadcast <- NewMessage("fft", pbFFT.FFTs)
			}

			if i%250 == 0 {
				second = time.Now().UnixNano()
				log.Println(second-last_second, "nanoseconds have elapsed between 250 samples.")
				last_second = second
			}

			i++

		}
	}
}
