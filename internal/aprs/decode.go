package aprs

import (
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
)

type Packet struct {
	Kind        string  `json:"kind"`
	HasPosition bool    `json:"has_position"`
	Latitude    float64 `json:"latitude,omitempty"`
	Longitude   float64 `json:"longitude,omitempty"`
	SymbolTable string  `json:"symbol_table,omitempty"`
	SymbolCode  string  `json:"symbol_code,omitempty"`
	Comment     string  `json:"comment,omitempty"`
	MessageTo   string  `json:"message_to,omitempty"`
	MessageText string  `json:"message_text,omitempty"`
	ObjectName  string  `json:"object_name,omitempty"`
	Alive       bool    `json:"alive,omitempty"`
}

func Decode(info []byte, destination string) Packet {
	s := string(info)
	if s == "" {
		return Packet{Kind: "empty"}
	}
	p := Packet{Kind: "other"}
	switch s[0] {
	case '!', '=':
		if pos, err := parsePosition(s[1:]); err == nil {
			return pos
		}
	case '/', '@':
		if len(s) > 8 {
			if pos, err := parsePosition(s[8:]); err == nil {
				return pos
			}
		}
	case '>':
		p.Kind = "status"
		p.Comment = strings.TrimSpace(s[1:])
		return p
	case ':':
		if len(s) >= 11 && s[10] == ':' {
			p.Kind = "message"
			p.MessageTo = strings.TrimSpace(s[1:10])
			p.MessageText = s[11:]
			p.Comment = p.MessageText
			return p
		}
	case ';':
		if len(s) >= 19 {
			p.Kind = "object"
			p.ObjectName = strings.TrimSpace(s[1:10])
			p.Alive = s[10] == '*'
			if pos, err := parsePosition(s[18:]); err == nil {
				p.HasPosition = pos.HasPosition
				p.Latitude = pos.Latitude
				p.Longitude = pos.Longitude
				p.SymbolTable = pos.SymbolTable
				p.SymbolCode = pos.SymbolCode
				p.Comment = pos.Comment
			}
			return p
		}
	case ')':
		// Item: )NAME!position or )NAME_position
		if i := strings.IndexAny(s[1:], "!_"); i >= 0 {
			i++
			p.Kind = "item"
			p.ObjectName = strings.TrimSpace(s[1:i])
			p.Alive = s[i] == '!'
			if pos, err := parsePosition(s[i+1:]); err == nil {
				p.HasPosition = pos.HasPosition
				p.Latitude = pos.Latitude
				p.Longitude = pos.Longitude
				p.SymbolTable = pos.SymbolTable
				p.SymbolCode = pos.SymbolCode
				p.Comment = pos.Comment
			}
			return p
		}
	case '_':
		p.Kind = "weather"
		p.Comment = s[1:]
		return p
	case 'T':
		if strings.HasPrefix(s, "T#") {
			p.Kind = "telemetry"
			p.Comment = s
			return p
		}
	case '?':
		p.Kind = "query"
		p.Comment = s
		return p
	case '}':
		p.Kind = "third-party"
		p.Comment = s[1:]
		return p
	case '`', '\'':
		p.Kind = "mic-e"
		p.Comment = s
		return p
	}
	_ = destination // reserved for Mic-E destination decoding in the parity phase
	p.Comment = s
	return p
}

func parsePosition(s string) (Packet, error) {
	if len(s) < 13 {
		return Packet{}, errors.New("position too short")
	}
	// Uncompressed: DDMM.mmN/DDDMM.mmW<symbol>
	if len(s) >= 19 && isDigit(s[0]) && isDigit(s[1]) {
		lat, err := parseLat(s[0:8])
		if err != nil {
			return Packet{}, err
		}
		lon, err := parseLon(s[9:18])
		if err != nil {
			return Packet{}, err
		}
		p := Packet{Kind: "position", HasPosition: true, Latitude: lat, Longitude: lon, SymbolTable: string(s[8]), SymbolCode: string(s[18])}
		if len(s) > 19 {
			p.Comment = strings.TrimSpace(s[19:])
		}
		return p, nil
	}
	// Compressed APRS: table + 4 base91 latitude + 4 base91 longitude + symbol.
	if len(s) >= 10 && validBase91(s[1:9]) {
		y, err := base91(s[1:5])
		if err != nil {
			return Packet{}, err
		}
		x, err := base91(s[5:9])
		if err != nil {
			return Packet{}, err
		}
		lat := 90.0 - float64(y)/380926.0
		lon := -180.0 + float64(x)/190463.0
		if math.Abs(lat) > 90 || math.Abs(lon) > 180 {
			return Packet{}, errors.New("compressed position out of range")
		}
		p := Packet{Kind: "position", HasPosition: true, Latitude: lat, Longitude: lon, SymbolTable: string(s[0]), SymbolCode: string(s[9])}
		// Bytes 10..12 may be compressed course/speed/altitude/type. Keep any later text as comment.
		if len(s) > 13 {
			p.Comment = strings.TrimSpace(s[13:])
		}
		return p, nil
	}
	return Packet{}, errors.New("unsupported position encoding")
}

func parseLat(s string) (float64, error) {
	if len(s) != 8 {
		return 0, errors.New("bad latitude length")
	}
	deg, err := strconv.Atoi(s[:2])
	if err != nil {
		return 0, err
	}
	mins, err := strconv.ParseFloat(s[2:7], 64)
	if err != nil {
		return 0, err
	}
	if deg > 90 || mins >= 60 {
		return 0, errors.New("latitude out of range")
	}
	v := float64(deg) + mins/60
	switch s[7] {
	case 'N':
	case 'S':
		v = -v
	default:
		return 0, fmt.Errorf("bad latitude hemisphere %q", s[7])
	}
	return v, nil
}
func parseLon(s string) (float64, error) {
	if len(s) != 9 {
		return 0, errors.New("bad longitude length")
	}
	deg, err := strconv.Atoi(s[:3])
	if err != nil {
		return 0, err
	}
	mins, err := strconv.ParseFloat(s[3:8], 64)
	if err != nil {
		return 0, err
	}
	if deg > 180 || mins >= 60 {
		return 0, errors.New("longitude out of range")
	}
	v := float64(deg) + mins/60
	switch s[8] {
	case 'E':
	case 'W':
		v = -v
	default:
		return 0, fmt.Errorf("bad longitude hemisphere %q", s[8])
	}
	return v, nil
}
func isDigit(b byte) bool { return b >= '0' && b <= '9' }
func validBase91(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] < '!' || s[i] > '{' {
			return false
		}
	}
	return true
}
func base91(s string) (int, error) {
	v := 0
	for i := 0; i < len(s); i++ {
		if s[i] < '!' || s[i] > '{' {
			return 0, errors.New("invalid base91")
		}
		v = v*91 + int(s[i]-33)
	}
	return v, nil
}
