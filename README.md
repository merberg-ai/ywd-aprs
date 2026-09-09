# YWD-APRS

A lightweight, headless APRS station, mapping, messaging, telemetry, weather, and IGate platform with a modern browser-based interface.

YWD-APRS is inspired by the capabilities of classic APRS applications such as Xastir, but is designed as a daemon-first service for Raspberry Pi and other Linux systems. The radio and APRS engine keep running independently of any browser session; the browser is simply the control and visualization console.

> **Project status:** active pre-alpha development on `dev`. The current `0A / early-0B` checkpoint is RX-only and ready for TCP KISS physical receive testing.

## Current working receive slice

The current development branch implements:

```text
TCP KISS
   ↓
KISS framing
   ↓
AX.25 decoding
   ↓
APRS classification / position decoding
   ↓
station + packet state
   ↓
append-only normalized JSONL RX log
   ↓
HTTP API
   ↓
YWD browser map / Last Heard / packet monitor
```

The current gate is **receive-only by construction**:

- the KISS client exposes no transmit/write API;
- no beacon or messaging scheduler exists;
- `tx.enabled: true` is rejected during configuration validation;
- the browser UI explicitly reports `RX ONLY // TX PATH ABSENT`.

## Quick development install

On Debian or Raspberry Pi OS:

```bash
curl -fsSL https://raw.githubusercontent.com/merberg-ai/ywd-aprs/dev/scripts/install-dev.sh | sudo bash
```

The default development configuration targets the current YWD-MMDVM-TNC TCP KISS endpoint:

```yaml
kiss:
  host: 192.168.1.11
  port: 8001
  kiss_port: 0
```

Then:

```bash
ywd-aprsctl doctor
journalctl -fu ywd-aprs
```

Open `http://<host-ip>:8080/` in a browser. See [docs/PHYSICAL-TEST-0A.md](docs/PHYSICAL-TEST-0A.md) for the physical acceptance procedure and included synthetic KISS smoke test.

## Goals

- Run comfortably on Raspberry Pi-class hardware, including Pi Zero-class systems where practical.
- Use TCP KISS as the first-class modem/TNC interface.
- Keep APRS processing in a persistent daemon rather than in the browser.
- Provide a responsive web UI with a live APRS map, Last Heard, station details, packet monitoring, messaging, weather, telemetry, objects/items, and history.
- Support offline/local maps in addition to online map sources.
- Provide straightforward install, update, backup, restore, diagnostics, and systemd service management.
- Maintain strict separation between KISS transport, AX.25 framing, APRS decoding, station state, persistence, and UI/API layers.
- Keep RF transmit paths explicit and safe; early milestones are RX-only.

## Architecture direction

```text
Browser / phone / tablet
          |
        HTTP
          |
      ywd-aprsd
          |
  +-------+------------------------------+
  | APRS engine                         |
  | station state / packet history      |
  | API / configuration / logging       |
  +------------------+------------------+
                     |
                  AX.25
                     |
                    KISS
                     |
                 TCP KISS
                     |
        TNC / Direwolf / YWD-MMDVM-TNC
                     |
                     RF
```

The browser is intentionally not part of the radio-processing path. Closing it does not stop the daemon.

## Current technology

- **Backend:** Go, standard library only at the current gate
- **Frontend:** embedded HTML/CSS/JavaScript
- **Map:** Leaflet + OpenStreetMap tiles during development
- **Current persistence:** normalized JSONL RX log plus in-memory station state
- **Planned persistence:** SQLite
- **API:** JSON HTTP API
- **Service management:** systemd
- **Primary modem interface:** TCP KISS

The frontend is embedded into the daemon, so normal installations do not require Node.js or npm.

## Development milestones

### 0A - RX foundation

- daemon and configuration
- TCP KISS client with reconnect handling
- KISS framing/deframing
- AX.25 decoding and TNC2 rendering
- initial APRS decoding
- receive packet logging
- HTTP/API state path
- packet monitor foundation
- protocol fixtures and regression tests
- **no RF transmit path**

### 0B - Live map and persistent domain state

- richer map and APRS symbols/overlays
- station inspector and aging
- SQLite station/packet/position history
- trails and historical playback
- live event transport instead of polling

### 0C - APRS decoding parity

- full Mic-E decode
- compressed-position edge cases
- objects and items
- weather
- telemetry
- status and capabilities
- bulletins and announcements
- APRS queries and edge cases

### 0D - Transmit features

- controlled station beaconing
- APRS messaging
- ACK/REJ and retry state
- object/item transmission
- explicit master and per-feature TX controls

### 0E - APRS-IS and IGate

- APRS-IS client and filters
- duplicate suppression
- RF-to-IS gating
- policy-controlled IS-to-RF gating
- IGate diagnostics and statistics

### 0F - Productization

- stable installer and updater with rollback
- backup/restore
- first-run browser setup wizard
- authentication / remote administration controls
- offline map support
- release binaries for common Linux architectures

See [docs/ROADMAP.md](docs/ROADMAP.md) and [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md).

## Branch policy

- `main` - stable project baseline and release-ready code
- `dev` - active development integration branch
- feature/checkpoint branches - short-lived work derived from `dev`

Development lands on `dev` first and moves to `main` only at deliberate checkpoints/releases.

## RF safety

YWD-APRS will eventually transmit APRS frames. Development milestones treat transmit capability as an explicit feature, not an incidental side effect. RX-only stages must not contain an active RF transmit path. Later TX functions will have master and per-feature controls and require deliberate configuration.

Operators are responsible for configuring station identity, frequency, paths, beacon rates, IGate policy, and transmissions in accordance with applicable amateur-radio rules and local network practices.

## License

YWD-APRS is licensed under the GNU General Public License v3.0. See [LICENSE](LICENSE).

## Project

YWD-APRS is part of the KJ6YWD/YWD family of amateur-radio and experimental networking projects.
