package openbci

import (
	"github.com/tarm/serial"
	"io"
	"time"
)

var command map[string][]byte = map[string][]byte{
	"stop":  []byte{'\x73'},
	"start": []byte{'\x61'},
	"reset": []byte{'\x76'},
}

func NewDevice(location string, baud int, readTimeout time.Duration) (io.ReadWriteCloser, error) {
	conf := &serial.Config{
		Name:        location,
		Baud:        baud,
		ReadTimeout: readTimeout,
	}
	conn, err := serial.OpenPort(conf)
	if err != nil {
		return nil, err
	}
	return &Device{
		r: conn,
		w: conn,
		c: conn,
	}, nil
}

type Device struct {
	r io.Reader
	w io.Writer
	c io.Closer
}

func (d *Device) Read(buf []byte) (int, error) {
	n, err := d.r.Read(buf)
	if err != nil {
		return 0, err
	}
	return n, nil

}

func isReset(buf []byte) bool {
	for _, val := range buf {
		if val == command["reset"][0] {
			return true
		}
	}
	return false
}

func (d *Device) Write(buf []byte) (int, error) {
	if isReset(buf) {
		n, err := d.reset(buf)
		if err != nil {
			return 0, err
		}
		return n, nil
	}
	n, err := d.w.Write(buf)
	if err != nil {
		return 0, err
	}
	return n, nil
}

func (d *Device) reset(buf []byte) (n int, err error) {
	var (
		n0, n1, n2, idx int
		init_array      [3]byte
		scrolling       [3]byte
	)
	buf = make([]byte, 1)
	n0, err = d.w.Write(command["stop"])
	if err != nil {
		return 0, err
	}
	n += n0
	time.Sleep(10 * time.Millisecond)
	n1, err = d.w.Write(command["reset"])
	if err != nil {
		return n, err
	}
	n += n1
	time.Sleep(10 * time.Millisecond) // wait

	init_array = [3]byte{'\x24', '\x24', '\x24'}

	for {
		_, err := d.Read(buf)
		if err != nil {
			return n, err
		}
		scrolling[idx%3] = buf[0]
		idx++
		if scrolling == init_array {
			n2, err = d.w.Write(command["start"])
			if err != nil {
				return n, err
			}
			n += n2
			return n, nil
		}
	}
}

func (d *Device) Close() error {
	err := d.c.Close()
	if err != nil {
		return err
	}
	return nil
}
