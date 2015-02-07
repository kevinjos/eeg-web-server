package main

type Packet struct {
	header, footer, seqNum                                 byte
	Chan1, Chan2, Chan3, Chan4, Chan5, Chan6, Chan7, Chan8 float64
	AccX, AccY, AccZ                                       int16
}

func NewPacket() *Packet {
	return &Packet{
		header: '\xa0',
		footer: '\xc0',
	}
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

func scaleToVolts(c int32) float64 {
	scaleFac := 4.5 / 24 / ((1 << 23) - 1)
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

func encodePacket(p *[33]byte) *Packet {
	packet := NewPacket()
	packet.seqNum = p[1]
	packet.Chan1 = scaleToVolts(convert24bitTo32bit(p[2:5]))
	packet.Chan2 = scaleToVolts(convert24bitTo32bit(p[5:8]))
	packet.Chan3 = scaleToVolts(convert24bitTo32bit(p[8:11]))
	packet.Chan4 = scaleToVolts(convert24bitTo32bit(p[11:14]))
	packet.Chan5 = scaleToVolts(convert24bitTo32bit(p[14:17]))
	packet.Chan6 = scaleToVolts(convert24bitTo32bit(p[17:20]))
	packet.Chan7 = scaleToVolts(convert24bitTo32bit(p[20:23]))
	packet.Chan8 = scaleToVolts(convert24bitTo32bit(p[23:26]))
	packet.AccX = convert16bitTo32bit(p[26:28])
	packet.AccY = convert16bitTo32bit(p[28:30])
	packet.AccZ = convert16bitTo32bit(p[30:32])
	return packet
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

//decodeStream implements the openbci packet protocol to
//assemble packets and sends packet arrays onto the packetStream
func decodeStream(byteStream chan byte, packetStream chan *Packet) {
	var (
		thisPacket [33]byte
		lastPacket [33]byte
		seqDiff    uint8
		sampPacket *Packet
	)
	sampPacket = NewPacket()

	for {
		b := <-byteStream
		if b == sampPacket.header {
			thisPacket[0] = b
			thisPacket[1] = <-byteStream

			switch {
			case lastPacket != [33]byte{}:
				seqDiff = difference(thisPacket[1], lastPacket[1])
			case lastPacket == [33]byte{}:
				seqDiff = 1
			}

			for i := 2; i < 32; i++ {
				thisPacket[i] = <-byteStream
			}

			footer := <-byteStream
			if footer != sampPacket.footer {
				//fmt.Println("expected footer [", sampPacket.footer, "] and received [", footer, "]")
				//fmt.Println(thisPacket)
				//fmt.Println(lastPacket)
			}
			thisPacket[32] = footer

			if seqDiff != 1 {
				//fmt.Println("Last seen sequence number [", lastPacket[1], "]. This sequence number [", thisPacket[1], "]")
				for seqDiff > 1 {
					lastPacket[1]++
					packetStream <- encodePacket(&lastPacket)
					seqDiff--
				}
			}

			packetStream <- encodePacket(&thisPacket)
			lastPacket = thisPacket
		}
	}
}
