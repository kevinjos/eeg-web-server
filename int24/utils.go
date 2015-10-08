package int24

func UnmarshalSLE(c []byte) int32 {
	x := int((int(c[2]) << 16) | (int(c[1]) << 8) | int(c[0]))
	if (x & 8388608) > 0 {
		x |= 4278190080
	} else {
		x &= 16777215
	}
	return int32(x)
}

func MarshalSLE(i int32) []byte {
	out := make([]byte, 3)
	out[0], out[1], out[2] = byte(i), byte(i>>8), byte(i>>16)
	return out
}

func UnmarshalSBE(c []byte) int32 {
	x := int((int(c[0]) << 16) | (int(c[1]) << 8) | int(c[2]))
	if (x & 8388608) > 0 {
		x |= 4278190080
	} else {
		x &= 16777215
	}
	return int32(x)
}

func MarshalSBE(i int32) []byte {
	out := make([]byte, 3)
	out[2], out[1], out[0] = byte(i), byte(i>>8), byte(i>>16)
	return out
}
