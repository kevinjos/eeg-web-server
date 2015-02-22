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

type testparsecommandpair struct {
	data   string
	result string
}

var testsparsecommand = []testparsecommandpair{
	{"/x/1/true", "!"},
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
