package tindeq

import (
	"fmt"
	"log"
	"sync"

	"tinygo.org/x/bluetooth"
)

type Device struct {
	ctlC         bluetooth.DeviceCharacteristic
	blDevice     bluetooth.Device
	batteryLevel uint32 // milliVolt

	mut             sync.RWMutex
	lastMeasurement Measurement
	updateCh        chan struct{}
}

func (d *Device) Tare() error {
	_, err := d.ctlC.Write([]byte{0x64, 0})
	return err
}

func (d *Device) SampleBattery() error {
	//TODO: stop measurement while probing
	_, err := d.ctlC.Write([]byte{0x6f, 0})
	return err
}

func (d *Device) StartMeasure() error {
	_, err := d.ctlC.Write([]byte{0x65, 0})
	return err
}

func (d *Device) StopMeasure() error {
	_, err := d.ctlC.Write([]byte{0x66, 0})
	return err
}

func (d *Device) Shutdown() error {
	_, err := d.ctlC.Write([]byte{0x6E, 0})
	return err
}

func (d *Device) Close() error {
	return d.blDevice.Disconnect()
}

func (d *Device) GetBattery() string {
	d.mut.RLock()
	defer d.mut.RUnlock()

	return fmt.Sprintf("%d mV", d.batteryLevel)
}

func (d *Device) GetLatestMeasurement() Measurement {
	d.mut.RLock()
	defer d.mut.RUnlock()

	return d.lastMeasurement
}

func (d *Device) handleData(buf []byte) {
	msg, err := parseData(buf)
	if err != nil {
		log.Fatal(err)
		return
	}

	switch m := msg.(type) {
	case *msgBattery:
		d.mut.Lock()
		d.batteryLevel = m.mV
		d.mut.Unlock()
		d.isUpdated()
	case *msgWeight:
		d.mut.Lock()
		d.lastMeasurement = m.weights[len(m.weights)-1]
		d.mut.Unlock()
		d.isUpdated()
	}
	log.Printf("got: %s\n", msg.String())
}

func (d *Device) isUpdated() {
	select {
	case d.updateCh <- struct{}{}:
	default:
	}
}

func (d *Device) GetUpdateCh() chan struct{} {
	if d.updateCh == nil {
		d.updateCh = make(chan struct{}, 1)
	}

	return d.updateCh
}
