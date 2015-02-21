package main

import (
	"testing"
)

type testparsecommandpair struct {
	data   string
	result string
}

var testsparsecommand = []testparsecommandpair{
	{"/x/1/true", channelOn["1"]},
	{"/x/1/false", "1"},
	{"/x/2/false", "2"},
	{"/x/2/x2000000X", "x2000000X"},
	{"/x/3/x3000000X", "x3000000X"},
	{"/x/0/x0000000X", "x1000000Xx2000000Xx3000000Xx4000000Xx5000000Xx6000000Xx7000000Xx8000000X"},
}

func TestParseComman(t *testing.T) {
	handle := NewHandle(&MindControl{})
	for _, pair := range testsparsecommand {
		res := handle.parseCommand(pair.data)
		if res != pair.result {
			t.Error(
				"For", pair.data,
				"Expected", pair.result,
				"Got", res,
			)
		}
	}
}
