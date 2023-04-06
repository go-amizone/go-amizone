package validator

import "net"

func ValidateHardwareAddr(addr net.HardwareAddr) error {
	string_repr := addr.String()
	_, err := net.ParseMAC(string_repr)
	return err
}
