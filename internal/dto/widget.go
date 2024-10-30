package dto

import (
	"fmt"
	"strings"
)

// WidgetPortBitmap represents single or bitmasked list of widget ports.
type WidgetPortBitmap int8

const (
	P WidgetPortBitmap = 1 << iota
	Q
	R
)

func (w WidgetPortBitmap) String() string {
	switch w {
	case P:
		return "P"
	case Q:
		return "Q"
	case R:
		return "R"
	}
	return "Unknown"
}

// ToBitString returns the binary representation of the WidgetPortBitmap. Since
// as 4 bit bitstring.
func (w WidgetPortBitmap) ToBitString() string {
	return fmt.Sprintf("%04b", w)
}

func WidgetPortFromString(str string) WidgetPortBitmap {
	str = strings.ToUpper(str)
	switch str {
	case "P":
		return P
	case "Q":
		return Q
	case "R":
		return R
	}
	return 0
}

func WidgetPortFromStrings(strs []string) WidgetPortBitmap {
	var ports WidgetPortBitmap
	for _, str := range strs {
		ports |= WidgetPortFromString(str)
	}
	return ports
}

type Widget struct {
	ID           uint64           `json:"id"`
	Name         string           `json:"name"`
	SerialNumber string           `json:"serial_number"`
	PortBitmap   WidgetPortBitmap `json:"port_bitmap"`
}

type WidgetConnections struct {
	// Owner widget
	SerialNumber string `json:"serial_number" redis:"-"`
	// Peer widgets. Empty strings mean that there is no connection to that port
	P_PeerSerialNumber string `json:"p_peer_serial_num" redis:"p_peer_sn"`
	R_PeerSerialNumber string `json:"r_peer_serial_num" redis:"r_peer_sn"`
	Q_PeerSerialNumber string `json:"q_peer_serial_num" redis:"q_peer_sn"`
}
