package marshaller

import (
	"net"
	"strings"
)

func Mac(mac net.HardwareAddr) string {
	return strings.Replace(mac.String(), ":", "-", -1)
}
