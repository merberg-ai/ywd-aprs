package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	data := []byte(`station:
  callsign: kj6ywd-10
kiss:
  host: 127.0.0.1
  port: 9001
  kiss_port: 2
  reconnect_seconds: 5
web:
  listen: ":9999"
tx:
  enabled: false
`)
	if err := os.WriteFile(path, data, 0600); err != nil {
		t.Fatal(err)
	}
	cfg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Station.Callsign != "KJ6YWD-10" || cfg.KISS.Port != 9001 || cfg.KISS.KISSPort != 2 || cfg.Web.Listen != ":9999" {
		t.Fatalf("unexpected config: %+v", cfg)
	}
}

func TestTXRejected(t *testing.T) {
	cfg := Default()
	cfg.TX.Enabled = true
	if err := Validate(cfg); err == nil {
		t.Fatal("expected TX validation failure")
	}
}
