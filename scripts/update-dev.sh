#!/usr/bin/env bash
set -euo pipefail
if [[ ${EUID:-$(id -u)} -ne 0 ]]; then echo 'Run with sudo.' >&2; exit 1; fi
SRC=/opt/ywd-aprs-src
if [[ ! -d "$SRC/.git" ]]; then echo 'Development source tree not found; run install-dev.sh first.' >&2; exit 1; fi
cd "$SRC"
git fetch --prune origin dev
git checkout -f dev
git reset --hard origin/dev
go test ./...
commit="$(git rev-parse --short=12 HEAD)"; build_date="$(date -u +%Y-%m-%dT%H:%M:%SZ)"
ldflags="-s -w -X github.com/merberg-ai/ywd-aprs/internal/buildinfo.Commit=$commit -X github.com/merberg-ai/ywd-aprs/internal/buildinfo.Date=$build_date"
go build -trimpath -ldflags "$ldflags" -o /usr/local/bin/ywd-aprsd.new ./cmd/ywd-aprsd
go build -trimpath -ldflags "$ldflags" -o /usr/local/bin/ywd-aprsctl.new ./cmd/ywd-aprsctl
/usr/local/bin/ywd-aprsd.new -config /etc/ywd-aprs/config.yaml -check
[[ -f /usr/local/bin/ywd-aprsd ]] && cp -f /usr/local/bin/ywd-aprsd /usr/local/bin/ywd-aprsd.previous
install -m 0755 /usr/local/bin/ywd-aprsd.new /usr/local/bin/ywd-aprsd
install -m 0755 /usr/local/bin/ywd-aprsctl.new /usr/local/bin/ywd-aprsctl
rm -f /usr/local/bin/ywd-aprsd.new /usr/local/bin/ywd-aprsctl.new
systemctl restart ywd-aprs
sleep 1
ywd-aprsctl doctor || true
