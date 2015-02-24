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
	"time"
	"math/rand"
	"io"
)

type MockConn struct {
}

func NewMockConn() *MockConn {
	return &MockConn{}
}

func (MockConn) Close() error {
	return nil
}

func (MockConn) Read(b []byte) (int, error) {
	a := []byte{1, 1, 1, 1, 1, 1, 1, 1}
	copy(b, a)
	i := rand.Int31()
	if i > (1<<30) {
		return len(b), nil
	} else {
		return len(b), io.EOF
	}
}

func (MockConn) Write(b []byte) (int, error) {
	i := rand.Int31()
	if i > (1<<30) {
		return len(b), nil
	} else {
		return len(b), io.ErrShortWrite
	}
}

func TestNewOpenBCI(t *testing.T) {
	openbci := NewOpenBCI()
	go func() {
		<-openbci.writeChan
		<-openbci.readChan
		<-openbci.timeoutChan
		<-openbci.resetChan
		<-openbci.pauseReadChan
		<-openbci.quitCommand
		<-openbci.quitRead
	}()
	openbci.writeChan <- "abc"
	openbci.readChan <- '\x00'
	openbci.timeoutChan <- true
	openbci.resetChan <- make(chan bool)
	openbci.pauseReadChan <- make(chan bool)
	openbci.quitCommand <- true
	openbci.quitRead <- true
	if openbci.conn != nil {
		t.Error(
			"For openbci.conn",
			"expected", nil,
			"got", openbci.conn,
		)
	}
}

func TestClose(t *testing.T) {
	d := NewOpenBCI()
	d.conn = NewMockConn()
	go func() {
		<-d.quitCommand
		<-d.quitRead
	}()
	d.Close()
}

func TestCommand(t *testing.T) {
	d := NewOpenBCI()
	d.conn = NewMockConn()
	mockReadChan := make(chan byte)
	mockResumeChan := make(chan bool)
	d.readChan = mockReadChan
	go d.command()
	for n := 0; n < 10; n++ {
		d.writeChan <- "abc"
	}
	d.resetChan <- mockResumeChan
	d.quitCommand <- true
}

func TestRead(t *testing.T) {
	d := NewOpenBCI()
	d.conn = NewMockConn()
	go d.read()
	go func() {
		for {
			select {
			case b := <-d.readChan:
				if b != 1 {
					t.Error(
						"For device read",
						"Expected 1",
						"Got", b,
					)
				}
			case m := <-d.timeoutChan:
				if m != true {
					t.Error(
						"For device readtimeout",
						"Expected true",
						"Got", m,
					)
				}
			}
		}
	}()
	mockResumeChan := make(chan bool)
	time.Sleep(100 * time.Millisecond)
	go func() { 
		mockResumeChan <- true 
	}()
	d.pauseReadChan <- mockResumeChan
	d.quitRead <- true
}

func TestReset(t *testing.T) {
	d := NewOpenBCI()
	d.conn = NewMockConn()
	resumeChan := make(chan bool)
	go d.reset(resumeChan)
	go func() {
		b := <-d.writeChan
		if b != "s" {
			t.Error(
				"For reset connection",
				"Expected s",
				"Got", b,
			)
		}
		go func() {
			resumeRead := <-d.pauseReadChan
			<-resumeRead
		}()
		b = <-d.writeChan
		if b != "v" {
			t.Error(
				"For reset connection",
				"Expected v",
				"Got", b,
			)
		}
	}()
	d.readChan <- '\x24'
	d.readChan <- '\x24'
	d.readChan <- '\x24'
	<-resumeChan
}
