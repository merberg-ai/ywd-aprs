package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/merberg-ai/ywd-aprs/internal/aprs"
	"github.com/merberg-ai/ywd-aprs/internal/ax25"
	"github.com/merberg-ai/ywd-aprs/internal/buildinfo"
	"github.com/merberg-ai/ywd-aprs/internal/config"
	"github.com/merberg-ai/ywd-aprs/internal/kiss"
	"github.com/merberg-ai/ywd-aprs/internal/packetlog"
	"github.com/merberg-ai/ywd-aprs/internal/state"
	"github.com/merberg-ai/ywd-aprs/internal/webui"
)

func main() {
	configPath := flag.String("config", "/etc/ywd-aprs/config.yaml", "configuration file")
	check := flag.Bool("check", false, "validate configuration and exit")
	showVersion := flag.Bool("version", false, "print version and exit")
	flag.Parse()
	if *showVersion {
		fmt.Printf("ywd-aprsd %s commit=%s built=%s\n", buildinfo.Version, buildinfo.Commit, buildinfo.Date)
		return
	}

	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("config: %v", err)
	}
	if *check {
		fmt.Printf("OK: %s (KISS %s, web %s, TX absent)\n", *configPath, cfg.KISSAddress(), cfg.Web.Listen)
		return
	}

	log.SetPrefix("ywd-aprsd: ")
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)
	log.Printf("starting %s commit=%s", buildinfo.Version, buildinfo.Commit)
	log.Printf("station=%s kiss=%s/%d web=%s", cfg.Station.Callsign, cfg.KISSAddress(), cfg.KISS.KISSPort, cfg.Web.Listen)
	log.Printf("RX-only development gate: no RF transmit path exists in this build")

	pktlog, err := packetlog.Open(cfg.Logging.PacketFile)
	if err != nil {
		log.Fatalf("packet log: %v", err)
	}
	defer pktlog.Close()
	st := state.New(500)
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	var seenKISSStatus bool
	var lastKISSConnected bool
	var lastKISSError string
	client := kiss.NewClient(kiss.ClientConfig{Address: cfg.KISSAddress(), KISSPort: cfg.KISS.KISSPort, ReconnectDelay: cfg.ReconnectDelay()}, func(port int, payload []byte) {
		frame, err := ax25.Decode(payload)
		if err != nil {
			st.AddParseError()
			log.Printf("AX25 decode: %v raw=%x", err, payload)
			return
		}
		decoded := aprs.Decode(frame.Info, frame.Destination.String())
		packet := st.Ingest(port, frame, decoded)
		if err := pktlog.Write(packet); err != nil {
			log.Printf("packet log write: %v", err)
		}
		log.Printf("RX %s [%s]", frame.TNC2(), decoded.Kind)
	}, func(status kiss.Status) {
		st.SetKISSStatus(status)
		if status.Connected {
			if !seenKISSStatus || !lastKISSConnected {
				log.Printf("KISS connected: %s", cfg.KISSAddress())
			}
		} else if status.LastError != "" && (!seenKISSStatus || lastKISSConnected || status.LastError != lastKISSError) {
			log.Printf("KISS offline: %s", status.LastError)
		}
		seenKISSStatus = true
		lastKISSConnected = status.Connected
		lastKISSError = status.LastError
	})
	go client.Run(ctx)

	httpServer := &http.Server{Addr: cfg.Web.Listen, Handler: webui.New(cfg, st).Handler(), ReadHeaderTimeout: 5 * time.Second}
	go func() {
		log.Printf("web UI listening on %s", cfg.Web.Listen)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("web server: %v", err)
			cancel()
		}
	}()

	<-ctx.Done()
	log.Printf("shutting down")
	shutdownCtx, stop := context.WithTimeout(context.Background(), 5*time.Second)
	defer stop()
	_ = httpServer.Shutdown(shutdownCtx)
}
