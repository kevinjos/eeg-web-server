package main

import (
	"log"
	"time"
)

//DecodeStream implements the openbci packet protocol to
//assemble packets and sends packet arrays onto the packetStream
func DecodeStream(quit chan bool, read *chan byte, serialRead chan byte,
	reset chan bool, serialReset chan chan bool, serialTimeout chan bool, packet chan *Packet,
	gain chan *[8]float64) {
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
	resetMonitorChan := make(chan bool)
	syncPktThresh = 2
	gains := [8]float64{24.0, 24.0, 24.0, 24.0, 24.0, 24.0, 24.0, 24.0}
	for {
		select {
		case <-quit:
			return
		case <-resetMonitorChan:
			read = &serialRead
		case <-reset:
			readstate, syncPktCtr = 0, 0
			var bogusChan chan byte
			read = &bogusChan
			serialReset <- reset
		case <-serialTimeout:
			lastPacket := lastPacket
			packet <- encodePacket(&lastPacket, 0, &gains, notSyncd)
		case g := <-gain:
			gains = *g
		case b = <-*read:
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
					log.Println("%d packets behind", seqDiff)
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
					thisPacket[j] = <-*read
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
