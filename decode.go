package main

import (
	"io"
	"log"
	"time"
)

//DecodeStream implements the openbci packet protocol to
//assemble packets and sends packet arrays onto the packetStream
func DecodeStream(packet chan *Packet, gain chan *[8]float64, quit chan bool,
	pause chan chan bool, device io.ReadWriteCloser) {
	var (
		b             uint8
		readstate     uint8
		thisPacket    [33]byte
		lastPacket    [33]byte
		seqDiff       uint8
		syncPktCtr    uint8
		syncPktThresh uint8
		notSyncd      bool
	)
	buf := make([]byte, 1)
	syncPktThresh = 2
	gains := [8]float64{24.0, 24.0, 24.0, 24.0, 24.0, 24.0, 24.0, 24.0}
	for {
		select {
		case <-quit:
			return
		case g := <-gain:
			gains = *g
		case resume := <-pause:
			<-resume
		default:
			_, err := device.Read(buf)
			if err == io.EOF {
				continue
			} else if err != nil {
				log.Fatalf("error reading from device: %s", err)
			}
			b = buf[0]
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
					log.Printf("%d packets behind\n", seqDiff)
				}
				for seqDiff > 1 {
					lastPacket[1]++
					time.Sleep(4 * time.Millisecond)
					packet <- encodePacket(&lastPacket, 100-seqDiff, &gains, notSyncd)
					seqDiff--
				}
				fallthrough
			case 3:
				for j := 2; j < 32; j++ {
					_, err = device.Read(buf)
					if err != nil {
						log.Fatalf("error reading from device: %s\n", err)
					}
					thisPacket[j] = buf[0]
				}
				readstate = 4
			case 4:
				switch {
				case b == '\xc0':
					thisPacket[32] = b
					lastPacket = thisPacket
					if syncPktCtr > syncPktThresh {
						packet <- encodePacket(&thisPacket, 100, &gains, !notSyncd)
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
