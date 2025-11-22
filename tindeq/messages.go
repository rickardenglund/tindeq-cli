package tindeq

import (
	"encoding/binary"
	"fmt"
	"math"
)

type msgBattery struct {
	mV uint32
}

func (b *msgBattery) String() string {
	return fmt.Sprintf("bat: %d", b.mV)
}

type msg interface {
	String() string
}

type msgWeight struct {
	weights []Measurement
}

func (m *msgWeight) String() string {
	first := m.weights[0]
	secs := float64(first.Millis) / 1000
	return fmt.Sprintf("#%d: %.2f (%.3f s)", len(m.weights), first.Weight, secs)
}

type msgBatteryWarning struct{}

func (m msgBatteryWarning) String() string {
	return "battery waning"
}

func parseData(buf []byte) (msg, error) {
	if len(buf) < 2 {
		return nil, fmt.Errorf("invalid msg, %v", buf)
	}
	const (
		battery        = 0
		measurement    = 1
		batteryWarning = 4
	)
	switch buf[0] {
	case battery:
		if buf[1] != 4 {
			return nil, fmt.Errorf("unexpected length: %d\n", buf[1])
		}
		v := binary.LittleEndian.Uint32(buf[2:6])
		return &msgBattery{mV: v}, nil
	case measurement:
		ms, err := parseMeasurements(buf[2:])
		if err != nil {
			return nil, err
		}

		return &msgWeight{weights: ms}, nil
	case batteryWarning:
		return &msgBatteryWarning{}, nil
	default:
		return nil, fmt.Errorf("unknown opcode: %d, len(%d)", buf[0], buf[1])
	}

}

type Measurement struct {
	Weight float32
	Millis uint32
}

func parseMeasurements(buf []byte) ([]Measurement, error) {
	if len(buf)%8 != 0 {
		return nil, fmt.Errorf("invalid length %d", len(buf))
	}

	res := make([]Measurement, len(buf)/8)
	for i := range res {
		offset := i * 8
		wInt := binary.LittleEndian.Uint32(buf[offset : offset+4])
		w := math.Float32frombits(wInt)
		t := binary.LittleEndian.Uint32(buf[offset+4 : offset+8])
		res[i] = Measurement{Weight: w, Millis: t}
	}

	return res, nil
}
