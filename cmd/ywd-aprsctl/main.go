package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/merberg-ai/ywd-aprs/internal/buildinfo"
)

func main() {
	url := flag.String("url", "http://127.0.0.1:8080", "YWD-APRS web/API URL")
	flag.Parse()
	cmd := "status"
	if flag.NArg() > 0 {
		cmd = flag.Arg(0)
	}
	switch cmd {
	case "version":
		fmt.Printf("ywd-aprsctl %s commit=%s built=%s\n", buildinfo.Version, buildinfo.Commit, buildinfo.Date)
	case "status":
		status(*url, false)
	case "doctor":
		status(*url, true)
	default:
		fmt.Fprintf(os.Stderr, "usage: ywd-aprsctl [-url URL] {status|doctor|version}\n")
		os.Exit(2)
	}
}

func status(base string, doctor bool) {
	client := &http.Client{Timeout: 3 * time.Second}
	resp, err := client.Get(base + "/api/v1/state?limit=1")
	if err != nil {
		fmt.Printf("[FAIL] API %s: %v\n", base, err)
		os.Exit(1)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("[FAIL] API read: %v\n", err)
		os.Exit(1)
	}
	var v struct {
		Runtime struct {
			StartedAt   time.Time `json:"started_at"`
			ParseErrors uint64    `json:"parse_errors"`
			KISS        struct {
				Connected     bool      `json:"connected"`
				RXFrames      uint64    `json:"rx_frames"`
				Reconnects    uint64    `json:"reconnects"`
				IgnoredFrames uint64    `json:"ignored_frames"`
				LastRX        time.Time `json:"last_rx"`
				LastError     string    `json:"last_error"`
			} `json:"kiss"`
		} `json:"runtime"`
		Stations []any `json:"stations"`
	}
	if err := json.Unmarshal(body, &v); err != nil {
		fmt.Printf("[FAIL] invalid API response: %v\n", err)
		os.Exit(1)
	}
	state := "OFFLINE"
	if v.Runtime.KISS.Connected {
		state = "ONLINE"
	}
	fmt.Printf("YWD-APRS %s\nAPI:   %s\nKISS:  %s\nRX:    %d frames\nSTN:   %d stations\nERR:   %d parse errors\n", buildinfo.Version, base, state, v.Runtime.KISS.RXFrames, len(v.Stations), v.Runtime.ParseErrors)
	if !v.Runtime.KISS.LastRX.IsZero() {
		fmt.Printf("LAST:  %s\n", v.Runtime.KISS.LastRX.Local().Format(time.RFC3339))
	}
	if v.Runtime.KISS.LastError != "" {
		fmt.Printf("KISS ERROR: %s\n", v.Runtime.KISS.LastError)
	}
	if doctor {
		fmt.Printf("RECONNECTS: %d\nIGNORED:    %d\nUPTIME:     %s\nTX:         NOT IMPLEMENTED (RX-only gate)\n", v.Runtime.KISS.Reconnects, v.Runtime.KISS.IgnoredFrames, time.Since(v.Runtime.StartedAt).Round(time.Second))
	}
	if !v.Runtime.KISS.Connected {
		os.Exit(1)
	}
}
