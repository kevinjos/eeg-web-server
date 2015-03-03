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
	"github.com/runningwild/go-fftw/fftw"
	"math/cmplx"
	"strconv"
)

type PacketBatcher struct {
	Chans         map[string][]float64
	FFTs          map[string][]float64
	SignalQuality float64
	packets       []*Packet
	size          int
}

func NewPacketBatcher(size int) *PacketBatcher {
	chans := make(map[string][]float64)
	ffts := make(map[string][]float64)
	for i := 1; i <= channels; i++ {
		chans["Chan"+strconv.Itoa(i)] = make([]float64, size)
		ffts["Chan"+strconv.Itoa(i)] = make([]float64, size/2)
	}
	return &PacketBatcher{
		Chans:   chans,
		FFTs:    ffts,
		packets: make([]*Packet, size),
		size:    size,
	}
}

func (pb *PacketBatcher) dft(input []float64) []float64 {
	data := fftw.NewArray(pb.size)
	for idx, val := range input {
		data.Set(idx, complex(val, 0.0))
	}
	forward := fftw.NewPlan(data, data, fftw.Forward, fftw.Estimate)
	forward.Execute()
	data_out := make([]float64, pb.size)
	for idx, val := range data.Elems {
		data_out[idx] = cmplx.Abs(val)
	}
	return data_out
}

func (pb *PacketBatcher) batch() {
	for i, p := range pb.packets {
		pb.Chans["Chan1"][i] = p.Chan1
		pb.Chans["Chan2"][i] = p.Chan2
		pb.Chans["Chan3"][i] = p.Chan3
		pb.Chans["Chan4"][i] = p.Chan4
		pb.Chans["Chan5"][i] = p.Chan5
		pb.Chans["Chan6"][i] = p.Chan7
		pb.Chans["Chan7"][i] = p.Chan7
		pb.Chans["Chan8"][i] = p.Chan8
	}
	pb.deleteEmptyChans()
}

func (pb *PacketBatcher) deleteEmptyChans() {
	for key, val := range pb.Chans {
		for i, v := range val {
			if v != 0 {
				break
			}
			if i == len(val)-1 {
				delete(pb.Chans, key)
			}
		}
	}
}

func (pb *PacketBatcher) setFFT() {
	for key, val := range pb.Chans {
		mirrored := pb.dft(val)
		pb.FFTs[key] = mirrored[:len(mirrored)/2]
	}
}

type Packet struct {
	header, footer, seqNum                                 byte
	Chan1, Chan2, Chan3, Chan4, Chan5, Chan6, Chan7, Chan8 float64
	AccX, AccY, AccZ                                       int16
	SignalQuality                                          uint8
	Synced                                                 bool
}

func NewPacket() *Packet {
	return &Packet{
		header:        '\xa0',
		footer:        '\xc0',
		SignalQuality: 100,
	}
}

func (p *Packet) RawChans() map[string][]float64 {
	m := make(map[string][]float64)
	m["Chan1"] = []float64{p.Chan1}
	m["Chan2"] = []float64{p.Chan2}
	m["Chan3"] = []float64{p.Chan3}
	m["Chan4"] = []float64{p.Chan4}
	m["Chan5"] = []float64{p.Chan5}
	m["Chan6"] = []float64{p.Chan6}
	m["Chan7"] = []float64{p.Chan7}
	m["Chan8"] = []float64{p.Chan8}
	return m
}

func encodePacket(p *[33]byte, sq byte, gain *[8]float64, synced bool) *Packet {
	packet := NewPacket()
	packet.seqNum = p[1]
	packet.Chan1 = scaleToVolts(convert24bitTo32bit(p[2:5]), gain[0])
	packet.Chan2 = scaleToVolts(convert24bitTo32bit(p[5:8]), gain[1])
	packet.Chan3 = scaleToVolts(convert24bitTo32bit(p[8:11]), gain[2])
	packet.Chan4 = scaleToVolts(convert24bitTo32bit(p[11:14]), gain[3])
	packet.Chan5 = scaleToVolts(convert24bitTo32bit(p[14:17]), gain[4])
	packet.Chan6 = scaleToVolts(convert24bitTo32bit(p[17:20]), gain[5])
	packet.Chan7 = scaleToVolts(convert24bitTo32bit(p[20:23]), gain[6])
	packet.Chan8 = scaleToVolts(convert24bitTo32bit(p[23:26]), gain[7])
	packet.AccX = convert16bitTo32bit(p[26:28])
	packet.AccY = convert16bitTo32bit(p[28:30])
	packet.AccZ = convert16bitTo32bit(p[30:32])
	packet.SignalQuality = sq
	packet.Synced = synced
	return packet
}

//conver24bitTo32bit takes a byte slice of len 3
//and converts the 24bit 2's complement integer
//to the type int32 representation
func convert24bitTo32bit(c []byte) int32 {
	x := int((int(c[0]) << 16) | (int(c[1]) << 8) | int(c[2]))
	if (x & 8388608) > 0 {
		x |= 4278190080
	} else {
		x &= 16777215
	}
	return int32(x)
}

func scaleToVolts(c int32, gain float64) float64 {
	scaleFac := 4.5 / gain / ((1 << 23) - 1)
	return scaleFac * float64(c)
}

//conver16bitTo32bit takes a byte slice of len 2
//and converts the 16bit 2's complement integer
//to the type int32 representation
func convert16bitTo32bit(a []byte) int16 {
	x := int((int(a[0]) << 8) | int(a[1]))
	if (x & 32768) > 0 {
		x |= 4294901760
	} else {
		x &= 65535
	}
	return int16(x)
}

//difference calculates the difference in sequence numbers
//accounting for wrap around of uint8s
func difference(x uint8, y uint8) uint8 {
	switch {
	case x > y:
		return x - y
	case x == 0 && y == 255:
		return 1
	case x == y:
		return 255
	}
	return (255 - y) + x + 1
}
