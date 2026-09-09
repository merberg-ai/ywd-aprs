package state

import (
	"testing"

	"github.com/merberg-ai/ywd-aprs/internal/aprs"
	"github.com/merberg-ai/ywd-aprs/internal/ax25"
)

func TestIngestStation(t *testing.T) {
	s := New(2)
	f := ax25.Frame{Source: ax25.Address{Callsign: "KJ6YWD", SSID: 9}, Destination: ax25.Address{Callsign: "APRS"}, Info: []byte("!4050.00N/12220.00W>test")}
	a := aprs.Decode(f.Info, "APRS")
	s.Ingest(0, f, a)
	snap := s.Snapshot(10)
	if len(snap.Stations) != 1 || !snap.Stations[0].HasPosition {
		t.Fatalf("bad snapshot: %+v", snap)
	}
}
