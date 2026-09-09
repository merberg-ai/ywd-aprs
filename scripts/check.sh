#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")/.."

echo '[1/5] gofmt'
unformatted="$(gofmt -l cmd internal)"
if [[ -n "$unformatted" ]]; then
  echo "$unformatted"
  echo 'Run gofmt on the files above.' >&2
  exit 1
fi

echo '[2/5] go vet'
go vet ./...

echo '[3/5] go test'
go test ./...

echo '[4/5] native build'
go build ./cmd/ywd-aprsd ./cmd/ywd-aprsctl

echo '[5/5] Raspberry Pi cross-builds'
tmp="$(mktemp -d)"; trap 'rm -rf "$tmp"' EXIT
GOOS=linux GOARCH=arm GOARM=6 go build -o "$tmp/ywd-aprsd-armv6" ./cmd/ywd-aprsd
GOOS=linux GOARCH=arm GOARM=7 go build -o "$tmp/ywd-aprsd-armv7" ./cmd/ywd-aprsd
GOOS=linux GOARCH=arm64 go build -o "$tmp/ywd-aprsd-arm64" ./cmd/ywd-aprsd

echo 'YWD-APRS checks PASS'
