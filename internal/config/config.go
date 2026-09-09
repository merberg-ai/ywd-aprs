package config

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Station StationConfig
	KISS    KISSConfig
	Web     WebConfig
	Logging LoggingConfig
	Maps    MapsConfig
	TX      TXConfig
}

type StationConfig struct {
	Callsign  string
	Latitude  float64
	Longitude float64
}

type KISSConfig struct {
	Host             string
	Port             int
	KISSPort         int
	ReconnectSeconds int
}

type WebConfig struct {
	Listen string
	Title  string
}

type LoggingConfig struct {
	PacketFile string
}

type MapsConfig struct {
	TileURL     string
	Attribution string
}

type TXConfig struct {
	Enabled bool
}

func Default() Config {
	return Config{
		Station: StationConfig{Callsign: "KJ6YWD-10"},
		KISS: KISSConfig{
			Host:             "192.168.1.11",
			Port:             8001,
			KISSPort:         0,
			ReconnectSeconds: 3,
		},
		Web: WebConfig{
			Listen: ":8080",
			Title:  "YWD-APRS",
		},
		Logging: LoggingConfig{PacketFile: "/var/lib/ywd-aprs/packets.jsonl"},
		Maps: MapsConfig{
			TileURL:     "https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png",
			Attribution: "OpenStreetMap contributors",
		},
		TX: TXConfig{Enabled: false},
	}
}

func Load(path string) (Config, error) {
	cfg := Default()
	f, err := os.Open(path)
	if err != nil {
		return cfg, err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	section := ""
	lineNo := 0
	for scanner.Scan() {
		lineNo++
		raw := scanner.Text()
		line := strings.TrimSpace(stripComment(raw))
		if line == "" {
			continue
		}
		if strings.HasSuffix(line, ":") && !strings.Contains(strings.TrimSuffix(line, ":"), ":") {
			section = strings.TrimSpace(strings.TrimSuffix(line, ":"))
			continue
		}
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			return cfg, fmt.Errorf("config line %d: expected key: value", lineNo)
		}
		key := strings.TrimSpace(parts[0])
		value := unquote(strings.TrimSpace(parts[1]))
		if err := apply(&cfg, section, key, value); err != nil {
			return cfg, fmt.Errorf("config line %d: %w", lineNo, err)
		}
	}
	if err := scanner.Err(); err != nil {
		return cfg, err
	}
	if err := Validate(cfg); err != nil {
		return cfg, err
	}
	return cfg, nil
}

func Validate(cfg Config) error {
	if strings.TrimSpace(cfg.Station.Callsign) == "" {
		return errors.New("station.callsign must not be empty")
	}
	if cfg.KISS.Host == "" {
		return errors.New("kiss.host must not be empty")
	}
	if cfg.KISS.Port < 1 || cfg.KISS.Port > 65535 {
		return errors.New("kiss.port must be between 1 and 65535")
	}
	if cfg.KISS.KISSPort < 0 || cfg.KISS.KISSPort > 15 {
		return errors.New("kiss.kiss_port must be between 0 and 15")
	}
	if cfg.KISS.ReconnectSeconds < 1 {
		return errors.New("kiss.reconnect_seconds must be at least 1")
	}
	if cfg.Web.Listen == "" {
		return errors.New("web.listen must not be empty")
	}
	if cfg.TX.Enabled {
		return errors.New("tx.enabled=true is rejected: RF transmit is intentionally not implemented in 0A")
	}
	return nil
}

func (c Config) KISSAddress() string {
	return fmt.Sprintf("%s:%d", c.KISS.Host, c.KISS.Port)
}

func (c Config) ReconnectDelay() time.Duration {
	return time.Duration(c.KISS.ReconnectSeconds) * time.Second
}

func apply(cfg *Config, section, key, value string) error {
	full := section + "." + key
	switch full {
	case "station.callsign":
		cfg.Station.Callsign = strings.ToUpper(value)
	case "station.latitude":
		v, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return err
		}
		cfg.Station.Latitude = v
	case "station.longitude":
		v, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return err
		}
		cfg.Station.Longitude = v
	case "kiss.host":
		cfg.KISS.Host = value
	case "kiss.port":
		v, err := strconv.Atoi(value)
		if err != nil {
			return err
		}
		cfg.KISS.Port = v
	case "kiss.kiss_port":
		v, err := strconv.Atoi(value)
		if err != nil {
			return err
		}
		cfg.KISS.KISSPort = v
	case "kiss.reconnect_seconds":
		v, err := strconv.Atoi(value)
		if err != nil {
			return err
		}
		cfg.KISS.ReconnectSeconds = v
	case "web.listen":
		cfg.Web.Listen = value
	case "web.title":
		cfg.Web.Title = value
	case "logging.packet_file":
		cfg.Logging.PacketFile = value
	case "maps.tile_url":
		cfg.Maps.TileURL = value
	case "maps.attribution":
		cfg.Maps.Attribution = value
	case "tx.enabled":
		v, err := strconv.ParseBool(value)
		if err != nil {
			return err
		}
		cfg.TX.Enabled = v
	default:
		return fmt.Errorf("unknown setting %s", strings.TrimPrefix(full, "."))
	}
	return nil
}

func stripComment(s string) string {
	inSingle, inDouble := false, false
	for i, r := range s {
		switch r {
		case '\'':
			if !inDouble {
				inSingle = !inSingle
			}
		case '"':
			if !inSingle {
				inDouble = !inDouble
			}
		case '#':
			if !inSingle && !inDouble {
				return s[:i]
			}
		}
	}
	return s
}

func unquote(s string) string {
	if len(s) >= 2 && ((s[0] == '"' && s[len(s)-1] == '"') || (s[0] == '\'' && s[len(s)-1] == '\'')) {
		return s[1 : len(s)-1]
	}
	return s
}
