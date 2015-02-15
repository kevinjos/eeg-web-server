package main

import (
	"log"
	"time"
	"math/rand"
)

type MindController struct {
	WriteStream        chan string
	PacketStream       chan *Packet
	ReadTimedoutStream chan bool
	ResetButton        chan bool
	ToDevReset         chan bool
	QuitButton         chan bool
	SerialDevice       *Device
	ByteStream         chan byte
	gain               [8]float64
}

func NewMindController() *MindController {
	//Make some channels
	byteStream := make(chan byte, readBufferSize)
	writeStream := make(chan string, 64)
	packetStream := make(chan *Packet)
	rtStream := make(chan bool)
	quitChan := make(chan bool)
	resetChan := make(chan bool)
	tdreset := make(chan bool)
	gain := [8]float64{24.0, 24.0, 24.0, 24.0, 24.0, 24.0, 24.0, 24.0}
	//Set up the serial device
	serialDevice := &Device{
		byteStream:  byteStream,
		rtStream:    rtStream,
		writeStream: writeStream,
		quitButton:  quitChan,
		resetButton: tdreset,
	}
	return &MindController{
		WriteStream:        writeStream,
		PacketStream:       packetStream,
		ReadTimedoutStream: rtStream,
		ResetButton:        resetChan,
		ToDevReset:         tdreset,
		QuitButton:         quitChan,
		SerialDevice:       serialDevice,
		ByteStream:         byteStream,
		gain:               gain,
	}
}

func (mc *MindController) Open() {
	mc.SerialDevice.open()
	mc.ResetButton <- true
}

func (mc *MindController) encodePacket(p *[33]byte, sq byte) *Packet {
	packet := NewPacket()
	packet.seqNum = p[1]
	packet.Chan1 = scaleToVolts(convert24bitTo32bit(p[2:5]), mc.gain[0])
	packet.Chan2 = scaleToVolts(convert24bitTo32bit(p[5:8]), mc.gain[1])
	packet.Chan3 = scaleToVolts(convert24bitTo32bit(p[8:11]), mc.gain[2])
	packet.Chan4 = scaleToVolts(convert24bitTo32bit(p[11:14]), mc.gain[3])
	packet.Chan5 = scaleToVolts(convert24bitTo32bit(p[14:17]), mc.gain[4])
	packet.Chan6 = scaleToVolts(convert24bitTo32bit(p[17:20]), mc.gain[5])
	packet.Chan7 = scaleToVolts(convert24bitTo32bit(p[20:23]), mc.gain[6])
	packet.Chan8 = scaleToVolts(convert24bitTo32bit(p[23:26]), mc.gain[7])
	packet.AccX = convert16bitTo32bit(p[26:28])
	packet.AccY = convert16bitTo32bit(p[28:30])
	packet.AccZ = convert16bitTo32bit(p[30:32])
	packet.SignalQuality = sq
	return packet
}

//decodeStream implements the openbci packet protocol to
//assemble packets and sends packet arrays onto the packetStream
func (mc *MindController) DecodeStream() {
	var (
		b             uint8
		readstate     uint8
		thisPacket    [33]byte
		lastPacket    [33]byte
		seqDiff       uint8
		syncPktCtr    uint8
		syncPktThresh uint8
	)
	syncPktThresh = 2
	for {
		select {
		case reset := <-mc.ResetButton:
			readstate, syncPktCtr = 0, 0
			if reset {
				mc.ToDevReset <- reset
			} else {
				mc.WriteStream <- "s"
			}
		case <-mc.ReadTimedoutStream:
			lastPacket := lastPacket
			mc.PacketStream <- mc.encodePacket(&lastPacket, 0)
		case b = <-mc.ByteStream:
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
					mc.PacketStream <- mc.encodePacket(&lastPacket, 100-seqDiff)
					time.Sleep(4 * time.Millisecond)
					seqDiff--
				}
				fallthrough
			case 3:
				for j := 2; j < 32; j++ {
					thisPacket[j] = <-mc.ByteStream
				}
				readstate = 4
			case 4:
				switch {
				case b == '\xc0':
					thisPacket[32] = b
					lastPacket = thisPacket
					if syncPktCtr > syncPktThresh {
						mc.PacketStream <- mc.encodePacket(&thisPacket, 100)
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

func (mc *MindController) GenTestPackets() {
	var gain float64 = 24
	for x := 0; ; x++{
		sign := func() int32 {
			if rand.Int31() > 1<<16 {
				return -1
			} else {
				return 1
			}
		}
		packet := NewPacket()
		packet.Chan1 = scaleToVolts(rand.Int31n(1<<23) * sign(), gain)
		packet.Chan2 = scaleToVolts(rand.Int31n(1<<23) * sign(), gain)
		packet.Chan3 = scaleToVolts(rand.Int31n(1<<23) * sign(), gain)
		packet.Chan4 = scaleToVolts(rand.Int31n(1<<23) * sign(), gain)
		packet.Chan5 = scaleToVolts(rand.Int31n(1<<23) * sign(), gain)
		packet.Chan6 = scaleToVolts(rand.Int31n(1<<23) * sign(), gain)
		packet.Chan7 = scaleToVolts(rand.Int31n(1<<23) * sign(), gain)
		packet.Chan8 = scaleToVolts(rand.Int31n(1<<23) * sign(), gain)
		mc.PacketStream <- packet
		time.Sleep(4 * time.Millisecond)
	}
}
