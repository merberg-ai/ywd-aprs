package state

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/merberg-ai/ywd-aprs/internal/aprs"
	"github.com/merberg-ai/ywd-aprs/internal/ax25"
)

func TestEmptySnapshotUsesArrays(t *testing.T) {
	snap := New(2).Snapshot(10)
	if snap.Stations == nil || snap.Packets == nil {
		t.Fatalf("empty snapshot must use non-nil arrays: %+v", snap)
	}
	b, err := json.Marshal(snap)
	if err != nil {
		t.Fatal(err)
	}
	got := string(b)
	if !strings.Contains(got, `"stations":[]`) || !strings.Contains(got, `"packets":[]`) {
		t.Fatalf("empty snapshot JSON must contain arrays, got %s", got)
	}
}

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
