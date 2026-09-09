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
	PID         byte      `json:"pid"`
	Info        []byte    `json:"-"`
	RawHex      string    `json:"raw_hex"`
}

func Decode(raw []byte) (Frame, error) {
	if len(raw) < 16 {
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
	// APRS normally uses UI (0x03) with PID 0xF0. For other U/I frames,
	// retaining the next byte as PID is still useful to the monitor.
	if offset < len(raw) {
		f.PID = raw[offset]
		offset++
	}
	if offset <= len(raw) {
		f.Info = append([]byte(nil), raw[offset:]...)
	}
	return f, nil
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
