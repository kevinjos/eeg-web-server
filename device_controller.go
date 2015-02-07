package main

import "time"

type MindController struct {
	WriteStream  chan string
	PacketStream chan *Packet
	ResetButton  chan bool
	QuitButton   chan bool
	serialDevice *Device
	byteStream   chan byte
}

func NewMindController() *MindController {
	//Configuration parameters for serial IO
	var (
		location    string
		baud        int
		readTimeout time.Duration
	)

	//Make some channels
	byteStream := make(chan byte, 2)
	writeStream := make(chan string)
	packetStream := make(chan *Packet)
	quitChan := make(chan bool)
	resetChan := make(chan bool)

	//Set up the serial device
	location = "/dev/ttyUSB0"
	baud = 115200
	readTimeout = 5 * time.Second
	serialDevice := NewDevice(location, baud, readTimeout, byteStream, writeStream, quitChan, resetChan)
	return &MindController{
		WriteStream:  writeStream,
		PacketStream: packetStream,
		ResetButton:  resetChan,
		QuitButton:   quitChan,
		serialDevice: serialDevice,
		byteStream:   byteStream,
	}
}

func (mc *MindController) Open() {
	mc.serialDevice.open()
	go decodeStream(mc.byteStream, mc.PacketStream)
}
