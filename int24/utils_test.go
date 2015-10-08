package int24

import "testing"

type test24pair struct {
	data   []byte
	result int32
}

var tests24 = []test24pair{
	{[]byte{0, 0, 0}, 0},
	{[]byte{255, 255, 255}, -1},
	{[]byte{0, 0, 128}, -8388608},
	{[]byte{255, 255, 127}, 8388607},
}

func TestConvert24bitTo32bit(t *testing.T) {
	for _, pair := range tests24 {
		res := Unmarshal(pair.data)
		if res != pair.result {
			t.Error(
				"For", pair.data,
				"expected", pair.result,
				"got", res,
			)
		}
	}
}

type test24revpair struct {
	data   int32
	result []byte
}

var tests24rev = []test24revpair{
	{0, []byte{0, 0, 0}},
	{-1, []byte{255, 255, 255}},
	{-8388608, []byte{0, 0, 128}},
	{8388607, []byte{255, 255, 127}},
}

func TestConvert32IntTo3ByteArray(t *testing.T) {
	for _, pair := range tests24rev {
		res := Marshal(pair.data)
		for idx, b := range res {
			if b != pair.result[idx] {
				t.Error(
					"For", pair.data,
					"expected", pair.result,
					"got", res,
				)
			}
		}
	}
}

func Test24bitToIntAndBack(t *testing.T) {
	for i := -8388608; i < 8388608; i++ {
		a := Marshal(int32(i))
		b := Unmarshal(a)
		if i != int(b) {
			t.Error(
				"For", i,
				"Expected", i,
				"Got", b,
			)
		}
	}
}
