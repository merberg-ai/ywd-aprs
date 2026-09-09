package ax25

import "testing"

func encAddr(call string, ssid int, last, repeated bool) []byte {
	out := make([]byte, 7)
	for i := 0; i < 6; i++ {
		ch := byte(' ')
		if i < len(call) {
			ch = call[i]
		}
		out[i] = ch << 1
	}
	out[6] = 0x60 | byte((ssid&0xf)<<1)
	if repeated {
		out[6] |= 0x80
	}
	if last {
		out[6] |= 1
	}
	return out
}

func TestDecodeUI(t *testing.T) {
	raw := append(encAddr("APRS", 0, false, false), encAddr("KJ6YWD", 9, false, false)...)
	raw = append(raw, encAddr("WIDE1", 1, true, true)...)
	raw = append(raw, 0x03, 0xF0)
	raw = append(raw, []byte("!4050.00N/12220.00W>test")...)
	f, err := Decode(raw)
	if err != nil {
		t.Fatal(err)
	}
	if f.Source.String() != "KJ6YWD-9" || f.Destination.String() != "APRS" || len(f.Path) != 1 || !f.Path[0].Repeated {
		t.Fatalf("unexpected frame: %+v", f)
	}
	if !f.IsAPRSUI() || f.FrameType() != "UI" {
		t.Fatalf("not APRS UI: %+v", f)
	}
	if f.TNC2() != "KJ6YWD-9>APRS,WIDE1-1*:!4050.00N/12220.00W>test" {
		t.Fatalf("tnc2=%q", f.TNC2())
	}
}
func TestDecodeShortRR(t *testing.T) {
	raw := append(encAddr("KJ6YWD", 5, false, false), encAddr("KE6CHO", 5, true, false)...)
	raw = append(raw, 0x81) // RR, N(R)=4, P/F=0
	f, err := Decode(raw)
	if err != nil {
		t.Fatal(err)
	}
	if f.FrameType() != "RR" || f.HasPID || len(f.Info) != 0 {
		t.Fatalf("bad RR: %+v type=%s", f, f.FrameType())
	}
}
func TestDecodeUA(t *testing.T) {
	raw := append(encAddr("KJ6YWD", 10, false, false), encAddr("KJ6YWD", 5, true, false)...)
	raw = append(raw, 0x63)
	f, err := Decode(raw)
	if err != nil {
		t.Fatal(err)
	}
	if f.FrameType() != "UA" || f.HasPID {
		t.Fatalf("bad UA: %+v", f)
	}
}
