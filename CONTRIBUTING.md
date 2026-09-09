# Contributing to YWD-APRS

YWD-APRS is an early-stage amateur-radio software project. Contributions, testing, protocol fixtures, documentation fixes, and hardware interoperability reports are welcome.

## Branch workflow

The repository uses a simple stability model:

- `main` is the stable project baseline and release branch.
- `dev` is the active integration branch.
- feature/checkpoint branches should normally branch from `dev`.

Do not develop directly on `main` unless a deliberate maintenance/release change requires it.

A typical contribution should:

1. start from current `dev`
2. use a focused branch
3. include tests or fixtures for protocol behavior where practical
4. document externally visible configuration/API changes
5. merge back to `dev`
6. reach `main` only as part of a deliberate checkpoint/release

## Architectural expectations

Please read `docs/ARCHITECTURE.md` before making substantial changes.

In particular:

- keep TCP transport, KISS, AX.25, APRS, domain state, persistence, and web/UI concerns separated
- do not hide malformed/unsupported packets when they can be preserved for diagnostics
- avoid browser-dependent radio processing
- avoid heavyweight runtime dependencies on production Raspberry Pi systems
- prefer bounded queues, bounded retry behavior, and configurable retention
- do not introduce an RF transmit path into milestones declared RX-only

## RF-related changes

Changes capable of transmitting over amateur radio require extra care.

A TX-capable change should include:

- a clear enable/disable state
- disabled-by-default behavior unless a documented existing configuration explicitly says otherwise
- tests proving disabled TX does not dispatch frames
- useful logging/diagnostics without leaking sensitive configuration
- documentation of the RF behavior being added

Do not use uncontrolled live RF as a substitute for unit tests. Protocol generation should be testable without a radio attached.

## Protocol fixtures

Real-world packet examples are especially valuable. When adding a decoder fix for an APRS or AX.25 edge case, add a regression fixture/test whenever possible so the same packet cannot silently break later.

Fixtures should avoid unnecessary personal/private information. Callsigns and packet data that were publicly transmitted on amateur radio may still be normalized or replaced when the exact identity is irrelevant to the test.

## Code style

For Go code:

- `gofmt` is authoritative
- keep packages focused
- avoid global mutable state when a component can own its dependencies explicitly
- errors should retain enough context to diagnose interface/protocol failures
- exported APIs should be documented when they form part of a package contract

For the browser application:

- keep production dependencies modest
- do not move APRS protocol/state ownership into frontend code
- make mobile/tablet use a first-class concern
- preserve accessibility and legibility even with the dark YWD visual style

## Commits

Prefer short, descriptive commit subjects such as:

```text
feat(kiss): add TCP reconnect state
fix(ax25): preserve repeated digi bit
feat(aprs): decode uncompressed positions
test(aprs): add Mic-E regression fixtures
docs: document 0A acceptance gate
```

Exact checkpoint commits used for physical qualification should not be rewritten after qualification.

## Issues and bug reports

Useful bug reports include:

- YWD-APRS version/commit
- operating system and architecture
- relevant interface type/configuration with secrets removed
- expected behavior
- actual behavior
- logs around the failure
- raw KISS/AX.25/APRS fixture when the issue is protocol-specific

Please do not post passwords, APRS-IS credentials, private keys, or other secrets in issues.
