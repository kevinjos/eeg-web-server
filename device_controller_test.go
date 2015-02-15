package main

import (
	"testing"
)

func TestMCOpen(t *testing.T) {
	mc := NewMindController()
	c := make(chan bool)
	go func(c chan bool) {
		reset := <- mc.ResetButton
		c <- reset
	}(c)
	mc.Open()
	reset := <-c
	if reset == false {
		t.Error(
			"For: MC.Open()",
			"Expected: true",
			"Got: false",
		)
	}
}
