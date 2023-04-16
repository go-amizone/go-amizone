package models

import (
	"net"

	"github.com/samber/lo"
)

type WifiMacInfo struct {
	RegisteredAddresses []net.HardwareAddr
	Slots               int
	FreeSlots           int

	// requestVerificationToken is used when submitting the form to register macs
	// It is not exported to keep it from being serialized with requests, as it is only (ostensibly) useful when not stale.
	// TODO: We could export this and instead use custom Stringer/JSON Marshaller to omit it from serialization.
	requestVerificationToken string
}

func (i *WifiMacInfo) GetRequestVerificationToken() string {
	return i.requestVerificationToken
}

func (i *WifiMacInfo) SetRequestVerificationToken(token string) {
	i.requestVerificationToken = token
}

func (i *WifiMacInfo) HasFreeSlot() bool {
	return i.FreeSlots > 0
}

func (i *WifiMacInfo) IsRegistered(mac net.HardwareAddr) bool {
	return lo.SomeBy(i.RegisteredAddresses, func(addr net.HardwareAddr) bool {
		return addr.String() == mac.String()
	})
}
