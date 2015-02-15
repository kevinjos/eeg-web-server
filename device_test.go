package main

import (
	"testing"
)

func TestOpenDevice(t *testing.T) {
	rtStream := make(chan bool)
	byteStream := make(chan byte)
	writeStream := make(chan string)
	quit := make(chan bool)
	reset := make(chan bool)
	d := &Device{
		byteStream:  byteStream,
		rtStream:    rtStream,
		writeStream: writeStream,
		quitButton:  quit,
		resetButton: reset,
	}
	c := make(chan bool)
	rt := false
	go func(c chan bool) {
		rt := <-rtStream
		c <- rt
	}(c)
	d.open()
	buf := make([]byte, 1)
	d.read(buf)
	if rt = <-c; rt != true {
		t.Error(
			"expecting", true,
			"got", rt,
		)
	}
}
