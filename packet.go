package main

import (
	"strconv"
)

type PacketBatcher struct {
	packets       [packetBatchSize]*Packet
	Chans         map[string][packetBatchSize]float64
	SignalQuality float64
}

func (pb *PacketBatcher) batch() {
	var (
		chan1         [packetBatchSize]float64
		chan2         [packetBatchSize]float64
		chan3         [packetBatchSize]float64
		chan4         [packetBatchSize]float64
		chan5         [packetBatchSize]float64
		chan6         [packetBatchSize]float64
		chan7         [packetBatchSize]float64
		chan8         [packetBatchSize]float64
		signalQuality [packetBatchSize]uint8
		chans         = [8]*[packetBatchSize]float64{&chan1, &chan2, &chan3, &chan4, &chan5, &chan6, &chan7, &chan8}
	)
	for i, p := range pb.packets {
		chan1[i] = p.Chan1
		chan2[i] = p.Chan2
		chan3[i] = p.Chan3
		chan4[i] = p.Chan4
		chan5[i] = p.Chan5
		chan6[i] = p.Chan7
		chan7[i] = p.Chan7
		chan8[i] = p.Chan8
		signalQuality[i] = p.SignalQuality
	}
	var emptyChan [packetBatchSize]float64
	for i, ch := range chans {
		if *ch != emptyChan {
			pb.Chans["Chan"+strconv.Itoa(i+1)] = *ch
		}
	}
	for _, sq := range signalQuality {
		pb.SignalQuality += float64(sq)
	}
	pb.SignalQuality /= packetBatchSize
}

func NewPacketBatcher() *PacketBatcher {
	chans := make(map[string][packetBatchSize]float64)
	var pkts [packetBatchSize]*Packet
	return &PacketBatcher{
		Chans:   chans,
		packets: pkts,
	}
}

type Packet struct {
	header, footer, seqNum                                 byte
	Chan1, Chan2, Chan3, Chan4, Chan5, Chan6, Chan7, Chan8 float64
	AccX, AccY, AccZ                                       int16
	SignalQuality                                          uint8
}

func NewPacket() *Packet {
	return &Packet{
		header:        '\xa0',
		footer:        '\xc0',
		SignalQuality: 100,
	}
}

func encodePacket(p *[33]byte, sq byte, gain *[8]float64) *Packet {
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
