package webui

import (
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/merberg-ai/ywd-aprs/internal/buildinfo"
	"github.com/merberg-ai/ywd-aprs/internal/config"
	"github.com/merberg-ai/ywd-aprs/internal/state"
)

//go:embed static/*
var assets embed.FS

type Server struct {
	cfg   config.Config
	state *state.Store
}

func New(cfg config.Config, st *state.Store) *Server { return &Server{cfg: cfg, state: st} }

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	sub, err := fs.Sub(assets, "static")
	if err != nil {
		panic(err)
	}
	mux.Handle("/", http.FileServer(http.FS(sub)))
	mux.HandleFunc("/healthz", s.health)
	mux.HandleFunc("/api/v1/state", s.getState)
	mux.HandleFunc("/api/v1/public-config", s.publicConfig)
	return withHeaders(mux)
}

func (s *Server) getState(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	limit := 100
	if v := r.URL.Query().Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 && n <= 500 {
			limit = n
		}
	}
	writeJSON(w, s.state.Snapshot(limit))
}
func (s *Server) publicConfig(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, map[string]any{"title": s.cfg.Web.Title, "callsign": s.cfg.Station.Callsign, "latitude": s.cfg.Station.Latitude, "longitude": s.cfg.Station.Longitude, "tile_url": s.cfg.Maps.TileURL, "attribution": s.cfg.Maps.Attribution, "version": buildinfo.Version, "commit": buildinfo.Commit})
}
func (s *Server) health(w http.ResponseWriter, r *http.Request) {
	snap := s.state.Snapshot(0)
	status := "ok"
	code := http.StatusOK
	if !snap.Runtime.KISS.Connected {
		status = "degraded"
		code = http.StatusServiceUnavailable
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	if err := json.NewEncoder(w).Encode(map[string]any{"status": status, "version": buildinfo.Version, "kiss_connected": snap.Runtime.KISS.Connected, "rx_frames": snap.Runtime.KISS.RXFrames, "parse_errors": snap.Runtime.ParseErrors, "started_at": snap.Runtime.StartedAt, "now": time.Now()}); err != nil {
		log.Printf("web json: %v", err)
	}
}
func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("web json: %v", err)
	}
}
func withHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Referrer-Policy", "no-referrer")
		w.Header().Set("Cache-Control", "no-store")
		next.ServeHTTP(w, r)
	})
}

func ListenAndServe(addr string, h http.Handler) error {
	srv := &http.Server{Addr: addr, Handler: h, ReadHeaderTimeout: 5 * time.Second}
	return srv.ListenAndServe()
}
func URLForListen(addr string) string {
	if len(addr) > 0 && addr[0] == ':' {
		return fmt.Sprintf("http://0.0.0.0%s", addr)
	}
	return "http://" + addr
}
