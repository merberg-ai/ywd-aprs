package state

import (
	"sort"
	"sync"
	"time"

	"github.com/merberg-ai/ywd-aprs/internal/aprs"
	"github.com/merberg-ai/ywd-aprs/internal/ax25"
	"github.com/merberg-ai/ywd-aprs/internal/kiss"
)

type Packet struct {
	ID          uint64    `json:"id"`
	ReceivedAt  time.Time `json:"received_at"`
	Port        int       `json:"port"`
	Source      string    `json:"source"`
	Destination string    `json:"destination"`
	Path        []string  `json:"path,omitempty"`
	TNC2        string    `json:"tnc2"`
	Info        string    `json:"info"`
	RawHex      string    `json:"raw_hex"`
	Kind        string    `json:"kind"`
	FrameType   string    `json:"frame_type"`
	Control     byte      `json:"control"`
	PID         byte      `json:"pid,omitempty"`
	HasPosition bool      `json:"has_position"`
	Latitude    float64   `json:"latitude,omitempty"`
	Longitude   float64   `json:"longitude,omitempty"`
	SymbolTable string    `json:"symbol_table,omitempty"`
	SymbolCode  string    `json:"symbol_code,omitempty"`
	Comment     string    `json:"comment,omitempty"`
}

type Station struct {
	Callsign    string    `json:"callsign"`
	LastHeard   time.Time `json:"last_heard"`
	Packets     uint64    `json:"packets"`
	Destination string    `json:"destination"`
	LastPath    []string  `json:"last_path,omitempty"`
	Kind        string    `json:"kind"`
	HasPosition bool      `json:"has_position"`
	Latitude    float64   `json:"latitude,omitempty"`
	Longitude   float64   `json:"longitude,omitempty"`
	SymbolTable string    `json:"symbol_table,omitempty"`
	SymbolCode  string    `json:"symbol_code,omitempty"`
	Comment     string    `json:"comment,omitempty"`
	LastTNC2    string    `json:"last_tnc2"`
}

type Runtime struct {
	StartedAt   time.Time   `json:"started_at"`
	ParseErrors uint64      `json:"parse_errors"`
	KISS        kiss.Status `json:"kiss"`
}

type Snapshot struct {
	Runtime  Runtime   `json:"runtime"`
	Stations []Station `json:"stations"`
	Packets  []Packet  `json:"packets"`
}

type Store struct {
	mu          sync.RWMutex
	startedAt   time.Time
	parseErrors uint64
	kiss        kiss.Status
	stations    map[string]*Station
	packets     []Packet
	nextID      uint64
	maxPackets  int
}

func New(maxPackets int) *Store {
	if maxPackets < 1 {
		maxPackets = 500
	}
	return &Store{startedAt: time.Now(), stations: map[string]*Station{}, maxPackets: maxPackets}
}

func (s *Store) SetKISSStatus(st kiss.Status) { s.mu.Lock(); s.kiss = st; s.mu.Unlock() }
func (s *Store) AddParseError()               { s.mu.Lock(); s.parseErrors++; s.mu.Unlock() }

func (s *Store) Ingest(port int, f ax25.Frame, a aprs.Packet) Packet {
	now := time.Now()
	path := make([]string, 0, len(f.Path))
	for _, p := range f.Path {
		v := p.String()
		if p.Repeated {
			v += "*"
		}
		path = append(path, v)
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.nextID++
	p := Packet{ID: s.nextID, ReceivedAt: now, Port: port, Source: f.Source.String(), Destination: f.Destination.String(), Path: path, TNC2: f.TNC2(), Info: string(f.Info), RawHex: f.RawHex, Kind: a.Kind, FrameType: f.FrameType(), Control: f.Control, PID: f.PID, HasPosition: a.HasPosition, Latitude: a.Latitude, Longitude: a.Longitude, SymbolTable: a.SymbolTable, SymbolCode: a.SymbolCode, Comment: a.Comment}
	s.packets = append(s.packets, p)
	if len(s.packets) > s.maxPackets {
		s.packets = append([]Packet(nil), s.packets[len(s.packets)-s.maxPackets:]...)
	}
	st := s.stations[p.Source]
	if st == nil {
		st = &Station{Callsign: p.Source}
		s.stations[p.Source] = st
	}
	st.LastHeard = now
	st.Packets++
	st.Destination = p.Destination
	st.LastPath = append([]string(nil), path...)
	st.Kind = a.Kind
	st.Comment = a.Comment
	st.LastTNC2 = p.TNC2
	if a.HasPosition {
		st.HasPosition = true
		st.Latitude = a.Latitude
		st.Longitude = a.Longitude
		st.SymbolTable = a.SymbolTable
		st.SymbolCode = a.SymbolCode
	}
	return p
}

func (s *Store) Snapshot(limit int) Snapshot {
	s.mu.RLock()
	defer s.mu.RUnlock()
	stations := make([]Station, 0, len(s.stations))
	for _, st := range s.stations {
		cp := *st
		cp.LastPath = append([]string(nil), st.LastPath...)
		stations = append(stations, cp)
	}
	sort.Slice(stations, func(i, j int) bool { return stations[i].LastHeard.After(stations[j].LastHeard) })
	if limit <= 0 || limit > len(s.packets) {
		limit = len(s.packets)
	}
	packets := append([]Packet(nil), s.packets[len(s.packets)-limit:]...)
	for i, j := 0, len(packets)-1; i < j; i, j = i+1, j-1 {
		packets[i], packets[j] = packets[j], packets[i]
	}
	return Snapshot{Runtime: Runtime{StartedAt: s.startedAt, ParseErrors: s.parseErrors, KISS: s.kiss}, Stations: stations, Packets: packets}
}
