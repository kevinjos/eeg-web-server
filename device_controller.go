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
	"log"
	"math/rand"
	"os"
	"strconv"
	"time"
)

type MindControl struct {
	ReadChan         *chan byte
	PacketChan       chan *Packet
	savePacketChan   chan *Packet
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
	var gain float64 = 24
	var on bool
	sign := func() int32 {
		if rand.Int31() > (1 << 30) {
			return -1
		} else {
			return 1
		}
	}
	for {
		select {
		case <-mc.quitGenTest:
			return
		case on = <-mc.genToggleChan:
			on = <-mc.genToggleChan
		default:
			if on {
				packet := NewPacket()
				packet.Chan1 = scaleToVolts(rand.Int31n(1<<23)*sign(), gain)
				packet.Chan2 = scaleToVolts(rand.Int31n(1<<23)*sign(), gain)
				packet.Chan3 = scaleToVolts(rand.Int31n(1<<23)*sign(), gain)
				packet.Chan4 = scaleToVolts(rand.Int31n(1<<23)*sign(), gain)
				packet.Chan5 = scaleToVolts(rand.Int31n(1<<23)*sign(), gain)
				packet.Chan6 = scaleToVolts(rand.Int31n(1<<23)*sign(), gain)
				packet.Chan7 = scaleToVolts(rand.Int31n(1<<23)*sign(), gain)
				packet.Chan8 = scaleToVolts(rand.Int31n(1<<23)*sign(), gain)
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
	var m *Message
	last_second := time.Now().UnixNano()
	second := time.Now().UnixNano()
	var i int

	pbFFT := NewPacketBatcher(FFTSize)
	pbRaw := NewPacketBatcher(RawMsgSize)

	for {
		select {
		case <-mc.quitSendPackets:
			return
		case p := <-mc.PacketChan:

			i++

			if mc.saving == true {
				mc.savePacketChan <- p
			}

			pbFFT.packets[i%FFTSize] = p
			pbRaw.packets[i%RawMsgSize] = p

			if i%RawMsgSize == 0 {
				pbRaw.batch()
				m = NewMessage("raw", pbRaw.Chans)
				mc.broadcast <- m
				pbRaw = NewPacketBatcher(RawMsgSize)
			}
			if i%FFTSize == 0 {
				pbFFT.batch()
				pbFFT.setFFT()
				m = NewMessage("fft", pbFFT.FFTs)
				mc.broadcast <- m
				pbFFT = NewPacketBatcher(FFTSize)
			}
			if i%250 == 0 {
				second = time.Now().UnixNano()
				log.Println(second-last_second, "nanoseconds have elapsed between 250 samples.")
				last_second = second
			}
		}
	}
}
