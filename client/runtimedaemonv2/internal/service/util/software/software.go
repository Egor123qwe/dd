package software

import (
	"errors"
	"net"

	"github.com/google/uuid"
)

func UniqueMachineID() (string, error) {
	macAddress, err := MacAddress()
	if err != nil {
		return "", err
	}

	return uuid.NewMD5(uuid.Nil, macAddress).String(), nil
}

func MacAddress() (net.HardwareAddr, error) {
	interfaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}

	for _, iface := range interfaces {
		if iface.HardwareAddr.String() != "" {
			return iface.HardwareAddr, nil
		}
	}

	return nil, errors.New("failed to get mac address")
}
