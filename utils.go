package main

import (
	"bytes"
	"math"
	"strconv"
	"time"
)

func calcFFTBins(fftSize int) (bins []float64) {
	bins = make([]float64, fftSize/2)
	step := float64(samplesPerSecond) / float64(fftSize)
	for idx := range bins {
		bins[idx] = step * float64(idx)
	}
	return bins
}

func genTestPackets(p chan *Packet, quit chan bool) {
	var val float64
	var i float64
	for {
		select {
		case <-quit:
			return
		default:
			i = i + 0.04
			val = 0.1*math.Sin(2.0*math.Pi*i) + 0.1*math.Cos(2.0*math.Pi*0.2*i)
			packet := NewPacket()
			packet.Chan1 = val
			packet.Rchan1 = []byte{'\x21', '\x21', '\x21'}
			packet.Chan2 = val
			packet.Rchan2 = []byte{'\x21', '\x21', '\x21'}
			packet.Chan3 = val
			packet.Rchan3 = []byte{'\x21', '\x21', '\x21'}
			packet.Chan4 = val
			packet.Rchan4 = []byte{'\x21', '\x21', '\x21'}
			packet.Chan5 = val
			packet.Rchan5 = []byte{'\x21', '\x21', '\x21'}
			packet.Chan6 = val
			packet.Rchan6 = []byte{'\x21', '\x21', '\x21'}
			packet.Chan7 = val
			packet.Rchan7 = []byte{'\x21', '\x21', '\x21'}
			packet.Chan8 = val
			packet.Rchan8 = []byte{'\x21', '\x21', '\x21'}
			p <- packet
			time.Sleep(4 * time.Millisecond)
		}
	}
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

func normalizeInPlace(input []float64) {
	var total float64
	for _, val := range input {
		total += val
	}
	for idx := range input {
		input[idx] /= total
	}
}
