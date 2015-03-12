package main

import "testing"

func TestMCClose(t *testing.T) {
	resChan := make(chan bool)
	broadcast := make(chan *Message)
	shutdown := make(chan bool)
	mc := NewMindControl(broadcast, shutdown)
	mc.SerialDevice.conn = NewMockConn()
	go func() {
		<-mc.SerialDevice.quitCommand
		<-mc.SerialDevice.quitRead
	}()
	go func() {
		if mc.saving == true {
			<-mc.quitSave
		}
		<-mc.quitDecodeStream
		<-mc.quitSendPackets
		<-mc.quitGenTest
		res := <-mc.shutdown
		resChan <- res
	}()
	mc.Close()
	res := <-resChan
	if res != true {
		t.Error(
			"For MindControl Close",
			"Expected true",
			"Got", res,
		)
	}
}

func TestMCStart(t *testing.T) {
	broadcast := make(chan *Message)
	shutdown := make(chan bool)
	mc := NewMindControl(broadcast, shutdown)
	mc.SerialDevice.conn = NewMockConn()
	mc.Start()
}
