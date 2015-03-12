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
	{8388607, 187500},
	{-8388607, -187500},
}

//TODO:Test differnt gain factors
func TestScale(t *testing.T) {
	for _, pair := range testsscale {
		res := scaleToMicroVolts(pair.data, 24)
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
	bc := make(chan *Message)
	sd := make(chan bool)
	mc := NewMindControl(bc, sd)
	for _, pair := range testsencode {
		res := encodePacket(pair.p, 100, &mc.gain, false)
		if *res != *pair.result {
			t.Error(
				"For x", pair.p,
				"expected", pair.result,
				"got", res,
			)
		}
	}
}
