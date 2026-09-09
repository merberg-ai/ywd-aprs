package kiss

import "testing"

func TestDecoderEscapesAndChunking(t *testing.T) {
	d := &Decoder{}
	part1 := []byte{FEND, 0x20, 0x01, FESC}
	part2 := []byte{TFEND, 0x02, FESC, TFESC, 0x03, FEND}
	if got := d.Feed(part1); len(got) != 0 {
		t.Fatalf("unexpected frame in first chunk: %v", got)
	}
	got := d.Feed(part2)
	if len(got) != 1 {
		t.Fatalf("got %d frames", len(got))
	}
	if got[0].Port != 2 || got[0].Command != 0 {
		t.Fatalf("bad command: %+v", got[0])
	}
	want := []byte{0x01, FEND, 0x02, FESC, 0x03}
	if string(got[0].Payload) != string(want) {
		t.Fatalf("payload %x want %x", got[0].Payload, want)
	}
}
