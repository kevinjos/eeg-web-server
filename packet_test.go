package main

import (
	"testing"
)

type test24pair struct {
	data   []byte
	result int32
}

var tests24 = []test24pair{
	{[]byte{0, 0, 0}, 0},
	{[]byte{255, 255, 255}, -1},
	{[]byte{128, 0, 0}, -8388608},
	{[]byte{127, 255, 255}, 8388607},
}

func TestConvert24bitTo32bit(t *testing.T) {
	for _, pair := range tests24 {
		res := convert24bitTo32bit(pair.data)
		if res != pair.result {
			t.Error(
				"For", pair.data,
				"expected", pair.result,
				"got", res,
			)
		}
	}
}

type test16pair struct {
	data   []byte
	result int16
}

var tests16 = []test16pair{
	{[]byte{0, 0}, 0},
	{[]byte{255, 255}, -1},
	{[]byte{128, 0}, -32768},
	{[]byte{127, 255}, 32767},
}

func TestConvert16bitTo32bit(t *testing.T) {
	for _, pair := range tests16 {
		res := convert16bitTo32bit(pair.data)
		if res != pair.result {
			t.Error(
				"For", pair.data,
				"expected", pair.result,
				"got", res,
			)
		}
	}
}

type testscalepair struct {
	data   int32
	result float64
}

var testsscale = []testscalepair{
	{0, 0},
	{8388607, 0.1875},
	{-8388607, -0.1875},
}

func TestScale(t *testing.T) {
	for _, pair := range testsscale {
		res := scaleToVolts(pair.data)
		if res != pair.result {
			t.Error(
				"For", pair.data,
				"expected", pair.result,
				"got", res,
			)
		}
	}
}

type testdiffpair struct {
	x, y   uint8
	result uint8
}

var testsdiff = []testdiffpair{
	{255, 254, 1},
	{0, 255, 1},
	{0, 0, 255},
	{5, 1, 4},
	{5, 255, 6},
}

func TestDifference(t *testing.T) {
	for _, pair := range testsdiff {
		res := difference(pair.x, pair.y)
		if res != pair.result {
			t.Error(
				"For x", pair.x, "and y", pair.y,
				"expected", pair.result,
				"got", res,
			)
		}
	}
}

func TestNewPacket(t *testing.T) {
	p := NewPacket()
	if p.header != '\xa0' || p.footer != '\xc0' {
		t.Error(
			"For packet header and packet footer",
			"expected \xa0 and \xc0",
			"got", p.header, "and", p.footer,
		)
	}
}

type testencodepair struct {
	p      *[33]byte
	result *Packet
}

var testsencode = []testencodepair{
	{&[33]byte{}, NewPacket()},
}

func TestEncodePacket(t *testing.T) {
	for _, pair := range testsencode {
		res := encodePacket(pair.p)
		if *res != *pair.result {
			t.Error(
				"For x", pair.p,
				"expected", pair.result,
				"got", res,
			)
		}
	}
}

func TestDecodeStream(t *testing.T) {
	byteStream := make(chan byte)
	packetStream := make(chan *Packet)
	packet_a := [33]byte{}

	go decodeStream(byteStream, packetStream)

	byteStream <- '\x00' //Send non header byte onto bytestream
	byteStream <- '\xa0' //Send header byte onto bytestream as header
	packet_a[0] = '\xa0'
	byteStream <- '\xff' //Send seqeunce number byte... the first time through the sequence number diff is aleays 1
	packet_a[1] = '\xff'
	for i := 0; i < 30; i++ {
		byteStream <- '\xa0' //Send payload bytes
		packet_a[i+2] = '\xa0'
	}
	byteStream <- '\xc0' //Send footer byte

	packet := <-packetStream
	expectedPacket := encodePacket(&packet_a)

	if *packet != *expectedPacket {
		t.Error(
			"For byte array", packet_a,
			"expected", expectedPacket,
			"got", packet,
		)
	}

	//Have another go around the packet loop with the correct sequence number
	byteStream <- '\x00' //Send non header byte onto bytestream
	byteStream <- '\xa0' //Send header byte onto bytestream as header
	packet_a[0] = '\xa0'
	byteStream <- '\x00' //Send seqeunce number byte... the first time through the sequence number diff is aleays 1
	packet_a[1] = '\x00'
	for i := 0; i < 30; i++ {
		byteStream <- byte(i) //Send payload bytes
		packet_a[i+2] = byte(i)
	}
	byteStream <- '\xc0' //Send footer byte

	packet = <-packetStream
	expectedPacket = encodePacket(&packet_a)

	if *packet != *expectedPacket {
		t.Error(
			"For byte array", packet_a,
			"expected", expectedPacket,
			"got", packet,
		)
	}

	byteStream <- '\x00' //Send non header byte onto bytestream
	byteStream <- '\xa0' //Send header byte onto bytestream as header
	byteStream <- '\x00' //Send an unaligned seqeunce number byte
	for i := 0; i < 30; i++ {
		byteStream <- byte(i + 30) //Send payload bytes
	}
	byteStream <- '\xc0' //Send footer byte

	packet = <-packetStream
	expectedPacket.seqNum++

	//When sequence numbers are unaligned, decodeStream sends the last valid packet through the packetStream
	//until sequence number alignment occurs
	if *packet != *expectedPacket {
		t.Error(
			"For byte array", packet_a,
			"expected", expectedPacket,
			"got", packet,
		)
	}
}
