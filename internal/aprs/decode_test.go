package aprs

import (
	"math"
	"testing"
)

func TestUncompressedPosition(t *testing.T) {
	p := Decode([]byte("!4050.00N/12220.00W>KJ6YWD Mobile"), "APRS")
	if p.Kind != "position" || !p.HasPosition {
		t.Fatalf("bad packet: %+v", p)
	}
	if math.Abs(p.Latitude-40.833333) > 0.00001 || math.Abs(p.Longitude+122.333333) > 0.00001 {
		t.Fatalf("bad coords: %+v", p)
	}
	if p.SymbolTable != "/" || p.SymbolCode != ">" {
		t.Fatalf("bad symbol: %+v", p)
	}
}
func TestTimestampedPosition(t *testing.T) {
	p := Decode([]byte("/092345z4050.00N/12220.00W>test"), "APRS")
	if !p.HasPosition {
		t.Fatalf("bad timestamped packet: %+v", p)
	}
}
func TestMessage(t *testing.T) {
	p := Decode([]byte(":KE6CHO-5 :hello{01"), "APRS")
	if p.Kind != "message" || p.MessageTo != "KE6CHO-5" || p.MessageText != "hello{01" {
		t.Fatalf("bad message: %+v", p)
	}
}
func TestStatus(t *testing.T) {
	p := Decode([]byte(">YWD-APRS test"), "APRS")
	if p.Kind != "status" || p.Comment != "YWD-APRS test" {
		t.Fatalf("bad status: %+v", p)
	}
}
