package device

import (
	"fmt"

	goios "github.com/danielpaulus/go-ios/ios"
)

type Device struct {
	Serial      string
	Name        string
	OSVersion   string
	ProductType string
	Entry       goios.DeviceEntry
}

func Get(udid string) (*Device, error) {
	entry, err := goios.GetDevice(udid)
	if err != nil {
		return nil, fmt.Errorf("device %s not found", udid)
	}
	return newDevice(udid, entry), nil
}

func List() ([]Device, error) {
	list, err := goios.ListDevices()
	if err != nil {
		return nil, err
	}

	devices := make([]Device, 0, len(list.DeviceList))
	for _, d := range list.DeviceList {
		devices = append(devices, *newDevice(d.Properties.SerialNumber, d))
	}
	return devices, nil
}

func newDevice(serial string, entry goios.DeviceEntry) *Device {
	d := &Device{
		Serial: serial,
		Entry:  entry,
		Name:   "iOS Device",
	}

	if values, err := goios.GetValues(entry); err == nil {
		if values.Value.DeviceName != "" {
			d.Name = values.Value.DeviceName
		} else if values.Value.ProductType != "" {
			d.Name = values.Value.ProductType
		}
		d.OSVersion = values.Value.ProductVersion
		d.ProductType = values.Value.ProductType
	}

	return d
}
