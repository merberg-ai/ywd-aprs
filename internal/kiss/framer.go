package kiss

const (
	FEND  byte = 0xC0
	FESC  byte = 0xDB
	TFEND byte = 0xDC
	TFESC byte = 0xDD
)

type Frame struct {
	Port    int
	Command byte
	Payload []byte
}

type Decoder struct {
	inFrame bool
	escaped bool
	buf     []byte
}

func (d *Decoder) Feed(data []byte) []Frame {
	var out []Frame
	for _, b := range data {
		switch {
		case b == FEND:
			if d.inFrame && len(d.buf) > 0 {
				if f, ok := decodeFrame(d.buf); ok {
					out = append(out, f)
				}
			}
			d.inFrame = true
			d.escaped = false
			d.buf = d.buf[:0]
		case !d.inFrame:
			continue
		case d.escaped:
			switch b {
			case TFEND:
				d.buf = append(d.buf, FEND)
			case TFESC:
				d.buf = append(d.buf, FESC)
			default:
				// Preserve unknown escapes rather than silently deleting data.
				d.buf = append(d.buf, FESC, b)
			}
			d.escaped = false
		case b == FESC:
			d.escaped = true
		default:
			d.buf = append(d.buf, b)
		}
	}
	return out
}

func decodeFrame(buf []byte) (Frame, bool) {
	if len(buf) < 1 {
		return Frame{}, false
	}
	cmd := buf[0]
	payload := append([]byte(nil), buf[1:]...)
	return Frame{Port: int((cmd >> 4) & 0x0f), Command: cmd & 0x0f, Payload: payload}, true
}
