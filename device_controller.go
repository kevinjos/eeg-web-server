package main

import (
	"log"
	"math/rand"
	"time"
	"os"
)

type MindControl struct {
	ReadChan     *chan byte
	PacketChan   chan *Packet
	ResetChan    chan bool
	genToggleChan chan bool
	SerialDevice *OpenBCI
	gain         [8]float64
}

func NewMindControl() *MindControl {
	//Set up the serial device
	serialDevice := NewOpenBCI()
	return &MindControl{
		ReadChan:     &serialDevice.readChan,
		PacketChan:   make(chan *Packet),
		ResetChan:    make(chan bool),
		genToggleChan:	make(chan bool),
		SerialDevice: serialDevice,
		gain:         [8]float64{24.0, 24.0, 24.0, 24.0, 24.0, 24.0, 24.0, 24.0},
	}
}

func (mc *MindControl) Open() {
	mc.SerialDevice.open()
	mc.ResetChan <- true
	go mc.SerialDevice.read()
}

func (mc *MindControl) Close() {
	mc.SerialDevice.Close()
	h.Close()
	close(mc.PacketChan)
	close(mc.ResetChan)
	close(mc.genToggleChan)
	os.Exit(1)
}

func (mc *MindControl) Start() {
	go mc.sendPackets()
	go mc.DecodeStream()
	go mc.SerialDevice.command()
	go mc.GenTestPackets()
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
		case <-resetMonitorChan:
			mc.ReadChan = &mc.SerialDevice.readChan
		case reset := <-mc.ResetChan:
			readstate, syncPktCtr = 0, 0
			if reset {
				var bogusChan chan byte
				mc.ReadChan = &bogusChan
				mc.SerialDevice.resetChan <- resetMonitorChan
			} else {
				mc.SerialDevice.writeChan <- "s"
			}
		case <-mc.SerialDevice.timeoutChan:
			lastPacket := lastPacket
			mc.PacketChan <- encodePacket(&lastPacket, 0, &mc.gain)
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
					mc.PacketChan <- encodePacket(&lastPacket, 100-seqDiff, &mc.gain)
					time.Sleep(4 * time.Millisecond)
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
						mc.PacketChan <- encodePacket(&thisPacket, 100, &mc.gain)
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
	sign := func() int32 {
		if rand.Int31() > (1 << 30) {
			return -1
		} else {
			return 1
		}
	}
	for {
		select {
		case <-mc.genToggleChan:
			for {
				select {
				case <-mc.genToggleChan:
					<-mc.genToggleChan
				default:
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
}
func (mc *MindControl) sendPackets() {
	last_second := time.Now().UnixNano()
	second := time.Now().UnixNano()

	pb := NewPacketBatcher()
	for i := 1; ; i++ {
		p := <-mc.PacketChan
		pb.packets[i%packetBatchSize] = p

		if i%packetBatchSize == 0 {
			pb.batch()
			h.broadcast <- pb
			pb = NewPacketBatcher()
		}
		if i%250 == 0 {
			second = time.Now().UnixNano()
			log.Println(second-last_second, "nanoseconds have elapsed between 250 samples.")
			last_second = second
		}
	}
}
