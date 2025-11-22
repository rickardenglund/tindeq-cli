package tindeq

import (
	"log"

	"tinygo.org/x/bluetooth"
)

var adapter = bluetooth.DefaultAdapter

func Scan() (*Device, error) {
	err := adapter.Enable()
	if err != nil {
		return nil, err
	}

	progressorService3, _ := bluetooth.ParseUUID("7e4e1701-1ea6-40c9-9dcc-13d34ffead57")
	dataChar, _ := bluetooth.ParseUUID("7e4e1702-1ea6-40c9-9dcc-13d34ffead57")
	ctlChar, err := bluetooth.ParseUUID("7e4e1703-1ea6-40c9-9dcc-13d34ffead57")

	deviceCh := make(chan bluetooth.Address, 1)

	err = adapter.Scan(func(adapter *bluetooth.Adapter, device bluetooth.ScanResult) {
		if !device.HasServiceUUID(progressorService3) {
			return
		}

		deviceCh <- device.Address
		err = adapter.StopScan()
		if err != nil {
			log.Fatal(err)
		}
	})

	adr := <-deviceCh

	d, err := adapter.Connect(adr, bluetooth.ConnectionParams{
		ConnectionTimeout: 0, //bluetooth.NewDuration(5 * time.Second),
		MinInterval:       0,
		MaxInterval:       0,
		Timeout:           0, //    bluetooth.NewDuration(10 * time.Second),
	})
	if err != nil {
		return nil, err
	}

	svcs, err := d.DiscoverServices([]bluetooth.UUID{progressorService3})
	if err != nil {
		return nil, err
	}

	res := new(Device)
	res.blDevice = d
	dataC := bluetooth.DeviceCharacteristic{}
	for _, s := range svcs {
		chars, err := s.DiscoverCharacteristics([]bluetooth.UUID{ctlChar, dataChar})
		if err != nil {
			return nil, err
		}
		for _, c := range chars {
			if c.UUID() == ctlChar {
				res.ctlC = c
			}
			if c.UUID() == dataChar {
				dataC = c
			}

		}
	}

	err = dataC.EnableNotifications(res.handleData)
	if err != nil {
		return nil, err
	}

	return res, nil
}
