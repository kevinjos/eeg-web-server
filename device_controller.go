package main

import (
	"log"
	"time"
	"math/rand"
)

type MindController interface {
	Open()
	DecodeStream()
	ReadWriteClose()
}

type MindControl struct {
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

func NewMindControl() *MindControl {
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
	return &MindControl{
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

func (mc *MindControl) Open() {
	mc.SerialDevice.open()
	mc.ResetButton <- true
}

func (mc *MindControl) ReadWriteClose() {
	mc.SerialDevice.ReadWriteClose()
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
			mc.PacketStream <- encodePacket(&lastPacket, 0, &mc.gain)
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
					mc.PacketStream <- encodePacket(&lastPacket, 100-seqDiff, &mc.gain)
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
						mc.PacketStream <- encodePacket(&thisPacket, 100, &mc.gain)
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

func GenTestPackets(stop chan bool) {
	var gain float64 = 24
	sign := func() int32 {
		if rand.Int31() > 1<<16 {
			return -1
		} else {
			return 1
		}
	}
	for {
		select {
		case <-stop:
			return
		default:
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
}
