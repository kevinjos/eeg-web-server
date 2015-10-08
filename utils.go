package main

import (
	"github.com/kevinjos/openbci-golang-server/int24"
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
	var val int32
	for {
		select {
		case <-quit:
			return
		default:
			val = val + 10000
			if val == 1<<22 {
				val = -1 * (1 << 22)
			}
			packet := NewPacket()
			v1 := val * (1 << 1)
			packet.Rchan1 = int24.MarshalSBE(v1)
			v2 := val * (1 << 2)
			packet.Rchan2 = int24.MarshalSBE(v2)
			v3 := val * (1 << 3)
			packet.Rchan3 = int24.MarshalSBE(v3)
			v4 := val * (1 << 4)
			packet.Rchan4 = int24.MarshalSBE(v4)
			v5 := val * (1 << 5)
			packet.Rchan5 = int24.MarshalSBE(v5)
			v6 := val * (1 << 6)
			packet.Rchan6 = int24.MarshalSBE(v6)
			v7 := val * (1 << 7)
			packet.Rchan7 = int24.MarshalSBE(v7)
			v8 := val * (1 << 8)
			packet.Rchan8 = int24.MarshalSBE(v8)
			packet.Chan1 = scaleToMicroVolts(v1, 24.0)
			packet.Chan2 = scaleToMicroVolts(v2, 24.0)
			packet.Chan3 = scaleToMicroVolts(v3, 24.0)
			packet.Chan4 = scaleToMicroVolts(v4, 24.0)
			packet.Chan5 = scaleToMicroVolts(v5, 24.0)
			packet.Chan6 = scaleToMicroVolts(v6, 24.0)
			packet.Chan7 = scaleToMicroVolts(v7, 24.0)
			packet.Chan8 = scaleToMicroVolts(v8, 24.0)
			p <- packet
			time.Sleep(4 * time.Millisecond)
		}
	}
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
