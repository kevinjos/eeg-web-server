package main

import (
	"testing"
	"time"
)

func TestNewDevice(t *testing.T) {
	loc := "/foo"
	baud := 115200
	readtimeout := time.Second * 1
	byteStream := make(chan byte)
	writeStream := make(chan string)
	quit := make(chan bool)
	reset := make(chan bool)
	d := NewDevice(loc, baud, readtimeout, byteStream, writeStream, quit, reset)
	if d.Baud != baud {
		t.Error(
			"expecting", baud,
			"got", d.Baud,
		)
	}
}
