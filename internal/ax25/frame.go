package ax25

import (
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
)

type Address struct {
	Callsign string `json:"callsign"`
	SSID     int    `json:"ssid"`
	Repeated bool   `json:"repeated,omitempty"`
}

func (a Address) String() string {
	if a.SSID == 0 {
		return a.Callsign
	}
	return fmt.Sprintf("%s-%d", a.Callsign, a.SSID)
}

type Frame struct {
	Destination Address   `json:"destination"`
	Source      Address   `json:"source"`
	Path        []Address `json:"path,omitempty"`
	Control     byte      `json:"control"`
	PID         byte      `json:"pid,omitempty"`
	HasPID      bool      `json:"has_pid"`
	Info        []byte    `json:"-"`
	RawHex      string    `json:"raw_hex"`
}

func Decode(raw []byte) (Frame, error) {
	// Two addresses (14 bytes) + one control byte is a valid short S/U frame.
	if len(raw) < 15 {
		return Frame{}, errors.New("AX.25 frame too short")
	}
	var addrs []Address
	offset := 0
	for {
		if offset+7 > len(raw) {
			return Frame{}, errors.New("truncated AX.25 address field")
		}
		a, last, err := decodeAddress(raw[offset : offset+7])
		if err != nil {
			return Frame{}, err
		}
		addrs = append(addrs, a)
		offset += 7
		if last {
			break
		}
		if len(addrs) > 10 {
			return Frame{}, errors.New("too many AX.25 addresses")
		}
	}
	if len(addrs) < 2 {
		return Frame{}, errors.New("AX.25 frame requires destination and source")
	}
	if offset >= len(raw) {
		return Frame{}, errors.New("missing AX.25 control field")
	}
	f := Frame{Destination: addrs[0], Source: addrs[1], Control: raw[offset], RawHex: hex.EncodeToString(raw)}
	if len(addrs) > 2 {
		f.Path = append([]Address(nil), addrs[2:]...)
	}
	offset++
	if controlHasPID(f.Control) {
		if offset >= len(raw) {
			return Frame{}, fmt.Errorf("%s frame missing PID", f.FrameType())
		}
		f.PID = raw[offset]
		f.HasPID = true
		offset++
		if offset < len(raw) {
			f.Info = append([]byte(nil), raw[offset:]...)
		}
	}
	return f, nil
}

func controlHasPID(control byte) bool {
	if control&0x01 == 0 {
		return true
	} // I frame, modulo-8 control
	return control&0xEF == 0x03 // UI frame, ignore P/F bit
}

func (f Frame) IsAPRSUI() bool { return f.Control&0xEF == 0x03 && f.HasPID && f.PID == 0xF0 }

func (f Frame) FrameType() string {
	c := f.Control
	if c&0x01 == 0 {
		return "I"
	}
	if c&0x03 == 0x01 {
		switch (c >> 2) & 0x03 {
		case 0:
			return "RR"
		case 1:
			return "RNR"
		case 2:
			return "REJ"
		case 3:
			return "SREJ"
		}
	}
	switch c & 0xEF {
	case 0x03:
		return "UI"
	case 0x2F:
		return "SABM"
	case 0x43:
		return "DISC"
	case 0x63:
		return "UA"
	case 0x0F:
		return "DM"
	case 0x87:
		return "FRMR"
	case 0xAF:
		return "XID"
	case 0xE3:
		return "TEST"
	default:
		return fmt.Sprintf("U-%02X", c&0xEF)
	}
}

func decodeAddress(b []byte) (Address, bool, error) {
	if len(b) != 7 {
		return Address{}, false, errors.New("address must be 7 bytes")
	}
	var sb strings.Builder
	for i := 0; i < 6; i++ {
		ch := b[i] >> 1
		if ch < 0x20 || ch > 0x7e {
			return Address{}, false, fmt.Errorf("invalid callsign byte %02x", b[i])
		}
		sb.WriteByte(ch)
	}
	call := strings.TrimSpace(sb.String())
	if call == "" {
		return Address{}, false, errors.New("empty callsign")
	}
	ssid := int((b[6] >> 1) & 0x0f)
	last := b[6]&0x01 != 0
	repeated := b[6]&0x80 != 0
	return Address{Callsign: call, SSID: ssid, Repeated: repeated}, last, nil
}

func (f Frame) TNC2() string {
	var b strings.Builder
	b.WriteString(f.Source.String())
	b.WriteByte('>')
	b.WriteString(f.Destination.String())
	for _, p := range f.Path {
		b.WriteByte(',')
		b.WriteString(p.String())
		if p.Repeated {
			b.WriteByte('*')
		}
	}
	b.WriteByte(':')
	b.Write(f.Info)
	return b.String()
}
