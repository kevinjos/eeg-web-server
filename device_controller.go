package main

import (
	"log"
	"time"
)

type MindController struct {
	WriteStream  chan string
	PacketStream chan *Packet
	ResetButton  chan bool
	toDevReset   chan bool
	QuitButton   chan bool
	serialDevice *Device
	byteStream   chan byte
	gain         [8]float64
}

func NewMindController() *MindController {
	//Configuration parameters for serial IO
	var (
		location    string
		baud        int
		readTimeout time.Duration
	)

	//Make some channels
	byteStream := make(chan byte, readBufferSize)
	writeStream := make(chan string, 64)
	packetStream := make(chan *Packet)
	quitChan := make(chan bool)
	resetChan := make(chan bool)
	tdreset := make(chan bool)
	gain := [8]float64{24.0, 24.0, 24.0, 24.0, 24.0, 24.0, 24.0, 24.0}
	//Set up the serial device
	location = "/dev/ttyUSB0"
	baud = 115200
	readTimeout = 5 * time.Second
	serialDevice := NewDevice(location, baud, readTimeout, byteStream, writeStream, quitChan, tdreset)
	return &MindController{
		WriteStream:  writeStream,
		PacketStream: packetStream,
		ResetButton:  resetChan,
		toDevReset:   tdreset,
		QuitButton:   quitChan,
		serialDevice: serialDevice,
		byteStream:   byteStream,
		gain:         gain,
	}
}

func (mc *MindController) Open() {
	mc.serialDevice.open()
	mc.serialDevice.reset()
	go mc.decodeStream(mc.byteStream, mc.PacketStream)
}

func (mc *MindController) encodePacket(p *[33]byte) *Packet {
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
	return packet
}

//decodeStream implements the openbci packet protocol to
//assemble packets and sends packet arrays onto the packetStream
func (mc *MindController) decodeStream(byteStream chan byte, packetStream chan *Packet) {
	var (
		readstate  uint8
		b          uint8
		thisPacket [33]byte
		lastPacket [33]byte
		seqDiff    uint8
	)
	for {
		select {
		case reset := <-mc.ResetButton:
			lastPacket = [33]byte{}
			if reset {
				mc.toDevReset <- reset
			} else {
				mc.WriteStream <- "s"
			}
			readstate = 0
		case b = <-byteStream:
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
				if lastPacket != [33]byte{} {
					seqDiff = difference(b, lastPacket[1])
				} else {
					seqDiff = 1
				}
				if seqDiff > 1 {
					log.Println("Dropped", seqDiff, "packets")
				}
				for seqDiff > 1 {
					lastPacket[1]++
					packetStream <- mc.encodePacket(&lastPacket)
					seqDiff--
				}
				fallthrough
			case 3:
				for j := 2; j < 32; j++ {
					thisPacket[j] = <-byteStream
				}
				readstate = 4
			case 4:
				if b == '\xc0' {
					thisPacket[32] = b
					lastPacket = thisPacket
					packetStream <- mc.encodePacket(&thisPacket)
					readstate = 1
				} else {
					log.Println("Footer out of sync")
					readstate = 0
				}
			}
		}
	}
}
