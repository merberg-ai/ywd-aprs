# YWD-APRS Roadmap

YWD-APRS is being developed in small checkpoints that each prove a useful vertical slice. The roadmap is intentionally receive-first so protocol correctness and observability are established before RF transmit features are introduced.

## 0A - RX foundation

**Objective:** prove the complete headless receive spine from TCP KISS through an API/event output while keeping RF TX absent.

Target flow:

```text
TCP KISS
  -> KISS decoder
  -> AX.25 decoder
  -> APRS decoder
  -> normalized event
  -> SQLite
  -> HTTP/WebSocket
```

Planned deliverables:

- Go module and daemon skeleton
- configuration loader and validation
- structured logging
- TCP KISS interface with reconnect state
- KISS framing/escaping parser
- AX.25 UI frame decoder
- initial APRS uncompressed position decoder
- APRS symbol extraction
- normalized packet/station event types
- SQLite schema and migration foundation
- packet persistence
- station/current-position persistence
- health/status HTTP endpoint
- WebSocket live-event stream
- unit tests and protocol fixtures
- CI on supported host architectures
- explicit no-TX contract

### 0A acceptance target

Using a real TCP KISS source, received APRS traffic should travel through the full stack and be observable through the service API/WebSocket with decoded AX.25/APRS fields and persisted packet/station state.

No browser map is required to call 0A complete.

## 0B - Live map console

**Objective:** turn the receive daemon into the first useful browser APRS console.

Planned deliverables:

- browser app shell
- KJ6YWD/YWD visual system
- live Leaflet map
- APRS symbol tables and overlays
- Last Heard list
- station inspector
- raw/decoded packet monitor
- interface connection/status view
- station aging
- position history and trails
- map filters and label modes

### 0B acceptance target

Live RF stations received through TCP KISS appear on the map and update without page reload. A station can be selected to inspect decoded/current information and packet history.

## 0C - APRS receive parity

**Objective:** substantially cover the APRS data types expected from a classic full-feature APRS station application.

Planned work:

- timestamped positions
- compressed positions
- Mic-E
- NMEA/GPS-style APRS payloads where applicable
- objects and items
- status
- capabilities
- messages as receive-only domain state
- bulletins/announcements
- weather
- telemetry
- queries
- PHG/range-related metadata
- third-party/network encapsulation as needed
- malformed/unsupported packet diagnostics
- expanded APRS regression corpus

### 0C acceptance target

Common real-world APRS packets on an active channel decode into normalized domain state and render appropriately in the browser, with unknown packets still visible for debugging.

## 0D - Controlled RF transmit

**Objective:** introduce RF transmit deliberately, with safety controls built into the architecture rather than bolted on later.

Planned deliverables:

- master RF TX enable
- per-feature TX controls
- station beacon generation
- beacon scheduling
- APRS message transmission
- ACK/REJ tracking
- bounded message retry state
- object/item transmission
- TX packet logging
- explicit UI TX indicators
- TX disabled-by-default migration rules

### 0D acceptance target

A configured station can intentionally beacon and exchange APRS messages through TCP KISS while all TX functions remain disabled by default on a fresh or upgraded installation.

## 0E - APRS-IS and IGate

**Objective:** make YWD-APRS useful as a network-connected APRS station/IGate.

Planned deliverables:

- APRS-IS connection management
- login/passcode configuration
- server rotation/failover
- APRS-IS filters
- RF-to-IS gating
- duplicate suppression
- configurable IS-to-RF gating policy
- gate statistics and diagnostics
- source/interface attribution throughout the UI

### 0E acceptance target

RF and APRS-IS traffic can coexist in one station database/map with clearly identified sources and policy-controlled gating.

## 0F - Productization

**Objective:** make installation and long-term operation appliance-like.

Planned deliverables:

- systemd unit
- one-command installer
- uninstaller
- `ywd-aprsctl`
- update checker
- signed/checksummed release artifacts
- automatic pre-update backup
- failed-update rollback
- backup/restore tooling
- `doctor` diagnostics
- first-run browser wizard
- authentication and remote-admin settings
- retention controls
- offline MBTiles/local map support
- release builds for amd64, arm64, armv7 and armv6 where dependencies permit

## 1.0 readiness

The first stable release should not merely mean “the map works.” Before 1.0, the project should have:

- an explicit Xastir-inspired feature parity matrix
- protocol regression coverage for supported APRS formats
- upgrade/migration testing
- documented backup/restore behavior
- documented RF TX safety behavior
- stable installation/update procedure
- stable configuration schema
- useful diagnostics when radio/network interfaces fail
- successful extended operation on Raspberry Pi hardware

## Deferred ideas

These are intentionally outside the early critical path but fit the architecture:

- multiple simultaneous RF KISS interfaces
- serial KISS
- AGWPE-compatible interface support
- packet replay/time-machine mode
- map heatmaps and RF coverage visualization
- station alerts/watch lists
- plugin/event-hook API
- GPS device integration for the local station
- weather alert overlays
- optional read-only public dashboard mode
- multi-user administration
