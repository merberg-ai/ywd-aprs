# YWD-APRS Architecture

This document defines the initial architectural contract for YWD-APRS. It is intentionally stricter than an implementation sketch: the goal is to keep the project lightweight, testable, safe around RF transmission, and usable as a long-running appliance.

## Core rule

**The daemon owns the APRS station. The browser only observes and controls it.**

Closing every browser session must not interrupt receive processing, persistence, messaging state, beacon scheduling, APRS-IS connectivity, or IGate operation once those features exist.

## Layering

The receive path is deliberately split into independent layers:

```text
TCP stream
  -> KISS framing
  -> AX.25 frame
  -> APRS payload
  -> normalized APRS event
  -> station/message/weather/etc. engines
  -> SQLite + event bus
  -> HTTP/WebSocket API
  -> browser UI
```

No layer should need to parse the representation owned by a layer more than one step below it.

### Transport layer

Initial interface: TCP KISS.

Responsibilities:

- establish and maintain TCP connection
- reconnect with bounded backoff
- expose connection state and statistics
- pass byte streams to the KISS layer
- never interpret APRS payloads

Future transports may include serial KISS, AGWPE-compatible input, APRS-IS, replay/test feeds, and other explicitly defined adapters.

### KISS layer

Responsibilities:

- FEND/FESC framing and escaping
- KISS command/port extraction
- data-frame delivery
- malformed-frame accounting
- optional future KISS parameter commands

KISS code must be independently testable with byte fixtures.

### AX.25 layer

Responsibilities:

- decode destination/source/digipeater addresses
- preserve SSIDs and repeated-path state
- decode control/PID fields
- expose UI information payloads to APRS
- retain raw frame bytes for diagnostics

AX.25 parsing must not contain APRS-specific interpretation.

### APRS layer

Responsibilities:

- classify APRS payload type
- decode positions, symbols, paths, messages, weather, telemetry, objects/items, status, capabilities, queries, and later Mic-E/compressed formats
- return normalized events rather than mutate global station state directly
- preserve undecoded/raw payload text for diagnostics

Unknown or partially supported APRS packets should remain visible in the packet monitor instead of being silently discarded.

### Domain engines

Domain state is built from normalized events.

Planned engines include:

- station state and Last Heard
- position history/trails
- APRS messaging and ACK/REJ state
- objects/items
- weather history
- telemetry history
- beacon scheduler
- APRS-IS client
- IGate policy and duplicate suppression

### Persistence

SQLite is the initial persistence layer.

Expected major tables:

- stations
- station_positions
- packets
- messages
- objects
- weather
- telemetry
- interfaces
- events/settings metadata

Schema migrations must be versioned and automatic. A failed migration must not silently destroy or recreate user data.

### API and event delivery

The daemon exposes:

- JSON HTTP endpoints for current state, configuration, history, diagnostics, and commands
- WebSocket event delivery for live stations, packets, interface state, messages, weather, and telemetry

The API should expose normalized domain objects, not internal Go implementation details.

### Browser UI

The UI is a static build embedded into release binaries where practical.

Primary views are expected to include:

- live map
- Last Heard
- selected-station inspector
- raw/decoded packet monitor
- messages/bulletins
- weather
- telemetry
- objects/items
- interfaces/status
- settings
- updates/about

The map is the primary operational view, in the spirit of classic APRS applications such as Xastir, but the visual language should follow the YWD/KJ6YWD dark radio-console style.

## Process model

Initial packaged processes:

```text
ywd-aprsd    long-running service
ywd-aprsctl  local administration/diagnostic CLI
```

`ywd-aprsctl` should communicate with the running daemon when appropriate instead of independently manipulating live database state.

## Configuration

Configuration should have two audiences:

1. operators using the browser setup/settings UI
2. advanced users managing a human-readable configuration file

Planned system paths:

```text
/usr/local/bin/ywd-aprsd
/usr/local/bin/ywd-aprsctl
/etc/ywd-aprs/config.yaml
/var/lib/ywd-aprs/ywd-aprs.db
/etc/systemd/system/ywd-aprs.service
```

Runtime-generated secrets must not be committed to source control or printed into normal logs.

## RF transmit safety

RF transmit capability is treated as a controlled subsystem.

Rules:

- 0A through 0C remain receive-oriented unless a milestone explicitly changes that contract.
- RX-only checkpoints must not accidentally dispatch KISS data frames to RF.
- later transmit support must have a master TX control plus feature-specific controls
- beacon, message, object/item, and IGate-to-RF transmission must be separately controllable
- UI state must clearly show whether TX is enabled
- tests must cover disabled-TX behavior
- configuration migration must never turn TX on implicitly

## Performance target

The project should remain practical on Raspberry Pi-class hardware.

Design implications:

- avoid heavyweight runtime services where a single daemon can do the job
- compile frontend assets ahead of release
- avoid requiring Node.js/Python toolchains on production systems
- use bounded queues and retention policies
- avoid unbounded packet/history growth by default
- prefer streaming/event-driven processing over frequent polling

## Failure behavior

The daemon should degrade visibly rather than fail mysteriously.

Examples:

- KISS disconnect: keep service/UI/database alive and reconnect
- malformed APRS packet: record it and increment decode diagnostics
- database error: surface degraded state prominently
- browser disconnect: no effect on radio processing
- APRS-IS failure: do not affect RF receive
- updater failure: preserve or roll back the previous runnable version

## Testing strategy

Protocol code should be fixture-driven.

`testdata/` will eventually contain captured or synthetic examples for:

- KISS escaping/framing
- AX.25 address/control/PID variations
- APRS positions
- compressed positions
- Mic-E
- messages/acks/rejects
- objects/items
- weather
- telemetry
- malformed and edge-case packets

Real packets encountered during field testing should be converted into regression fixtures whenever practical.

## Branch model

- `main`: stable baseline / releases
- `dev`: active integration branch
- feature/checkpoint branches: derived from `dev`

Development should preserve exact checkpoint commits when physical RF tests qualify a milestone.
