# YWD-APRS

A lightweight, headless APRS station, mapping, messaging, telemetry, weather, and IGate platform with a modern browser-based interface.

YWD-APRS is inspired by the capabilities of classic APRS applications such as Xastir, but is designed as a daemon-first service for Raspberry Pi and other Linux systems. The radio and APRS engine keep running independently of any browser session; the browser is simply the control and visualization console.

> **Project status:** early development / pre-alpha. The current development target is the RX-only 0A foundation.

## Goals

- Run comfortably on Raspberry Pi-class hardware, including Pi Zero-class systems where practical.
- Use TCP KISS as the first-class modem/TNC interface.
- Keep APRS processing in a persistent daemon rather than in the browser.
- Provide a responsive web UI with a live APRS map, Last Heard, station details, packet monitoring, messaging, weather, telemetry, objects/items, and history.
- Support offline/local maps in addition to online map sources.
- Provide straightforward install, update, backup, restore, diagnostics, and systemd service management.
- Maintain strict separation between KISS transport, AX.25 framing, APRS decoding, station state, persistence, and UI/API layers.
- Keep RF transmit paths explicit and safe; early milestones are RX-only.

## Planned architecture

```text
Browser / phone / tablet
          |
    HTTP + WebSocket
          |
      ywd-aprsd
          |
  +-------+------------------------------+
  | APRS engine                         |
  | station state / messages / WX       |
  | telemetry / objects / history       |
  | SQLite / API / configuration        |
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

The browser is intentionally not part of the radio-processing path. Closing the browser must not stop receiving, decoding, logging, beaconing, messaging, or IGate operation once those features are implemented.

## Initial technology direction

- **Backend:** Go
- **Frontend:** TypeScript with a lightweight component framework
- **Map:** Leaflet-compatible browser mapping
- **Database:** SQLite
- **Live updates:** WebSocket
- **API:** JSON HTTP API
- **Service management:** systemd
- **Primary modem interface:** TCP KISS

The final frontend will be built ahead of release and embedded into the daemon so normal installations do not require Node.js, npm, or a development toolchain.

## Development milestones

### 0A - RX foundation

The first milestone proves the complete receive spine while keeping transmit disabled:

```text
TCP KISS -> KISS decode -> AX.25 decode -> APRS decode -> SQLite -> API/WebSocket
```

Initial 0A work includes:

- daemon and configuration skeleton
- TCP KISS client with reconnect handling
- KISS framing/deframing
- AX.25 UI frame decoding
- initial APRS position decoding
- normalized APRS event model
- SQLite persistence
- HTTP/WebSocket event path
- raw packet monitor foundation
- protocol fixtures and regression tests
- **no RF transmit path**

### 0B - Live map

- browser application shell
- live APRS map
- APRS symbols and overlays
- Last Heard
- station inspector
- station aging
- position history and trails

### 0C - APRS decoding parity

- Mic-E
- compressed positions
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

- APRS-IS client
- filters
- duplicate suppression
- RF-to-IS gating
- policy-controlled IS-to-RF gating
- IGate diagnostics and statistics

### 0F - Productization

- installer and uninstaller
- updater with rollback
- backup/restore
- `ywd-aprsctl` diagnostics and service controls
- first-run setup wizard
- authentication / remote administration controls
- offline map support
- release binaries for common Linux architectures

See [docs/ROADMAP.md](docs/ROADMAP.md) for the working roadmap and [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md) for design rules.

## Branch policy

- `main` - stable project baseline and release-ready code
- `dev` - active development integration branch
- feature/checkpoint branches - short-lived work derived from `dev`

Development should land on `dev` first and move to `main` only at deliberate checkpoints/releases.

## Intended deployment

A normal packaged installation is expected to look approximately like:

```text
/usr/local/bin/ywd-aprsd
/usr/local/bin/ywd-aprsctl
/etc/ywd-aprs/config.yaml
/var/lib/ywd-aprs/ywd-aprs.db
/etc/systemd/system/ywd-aprs.service
```

The long-term goal is a one-command installer followed by browser-based setup.

## RF safety

YWD-APRS will eventually transmit APRS frames. Development milestones must treat transmit capability as an explicit feature, not an incidental side effect. RX-only stages must not contain an active RF transmit path. Later TX functions will have master and per-feature controls and will require deliberate configuration.

Operators are responsible for configuring station identity, frequency, paths, beacon rates, IGate policy, and transmissions in accordance with applicable amateur-radio rules and local network practices.

## License

YWD-APRS is licensed under the GNU General Public License v3.0. See [LICENSE](LICENSE).

## Project

YWD-APRS is part of the KJ6YWD/YWD family of amateur-radio and experimental networking projects.
