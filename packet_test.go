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

//TODO:Test differnt gain factors
func TestScale(t *testing.T) {
	for _, pair := range testsscale {
		res := scaleToVolts(pair.data, 24)
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
	mc := NewMindController()
	for _, pair := range testsencode {
		res := mc.encodePacket(pair.p)
		if *res != *pair.result {
			t.Error(
				"For x", pair.p,
				"expected", pair.result,
				"got", res,
			)
		}
	}
}
