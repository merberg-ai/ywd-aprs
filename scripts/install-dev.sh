#!/usr/bin/env bash
set -euo pipefail

REPO="https://github.com/merberg-ai/ywd-aprs.git"
BRANCH="dev"
SRC="/opt/ywd-aprs-src"
CONF_DIR="/etc/ywd-aprs"
DATA_DIR="/var/lib/ywd-aprs"

if [[ ${EUID:-$(id -u)} -ne 0 ]]; then
  echo "Run as root, e.g.: curl -fsSL https://raw.githubusercontent.com/merberg-ai/ywd-aprs/dev/scripts/install-dev.sh | sudo bash" >&2
  exit 1
fi

export DEBIAN_FRONTEND=noninteractive
echo '================================================'
echo ' YWD-APRS DEV INSTALL // RX-ONLY PHYSICAL GATE'
echo '================================================'
apt-get update
apt-get install -y --no-install-recommends ca-certificates git golang

if id ywd-aprs >/dev/null 2>&1; then :; else useradd --system --home "$DATA_DIR" --shell /usr/sbin/nologin ywd-aprs; fi
install -d -o ywd-aprs -g ywd-aprs -m 0750 "$DATA_DIR"
install -d -o root -g ywd-aprs -m 0750 "$CONF_DIR"

if [[ -d "$SRC/.git" ]]; then
  git -C "$SRC" fetch --prune origin "$BRANCH"
  git -C "$SRC" checkout -f "$BRANCH"
  git -C "$SRC" reset --hard "origin/$BRANCH"
else
  rm -rf "$SRC"
  git clone --depth 1 --branch "$BRANCH" "$REPO" "$SRC"
fi

cd "$SRC"
echo '[BUILD] tests'
go test ./...
commit="$(git rev-parse --short=12 HEAD)"
build_date="$(date -u +%Y-%m-%dT%H:%M:%SZ)"
ldflags="-s -w -X github.com/merberg-ai/ywd-aprs/internal/buildinfo.Commit=$commit -X github.com/merberg-ai/ywd-aprs/internal/buildinfo.Date=$build_date"
echo '[BUILD] ywd-aprsd + ywd-aprsctl'
go build -trimpath -ldflags "$ldflags" -o /usr/local/bin/ywd-aprsd ./cmd/ywd-aprsd
go build -trimpath -ldflags "$ldflags" -o /usr/local/bin/ywd-aprsctl ./cmd/ywd-aprsctl
chmod 0755 /usr/local/bin/ywd-aprsd /usr/local/bin/ywd-aprsctl

if [[ ! -f "$CONF_DIR/config.yaml" ]]; then
  install -o root -g ywd-aprs -m 0640 config/config.example.yaml "$CONF_DIR/config.yaml"
  echo '[CONFIG] installed default /etc/ywd-aprs/config.yaml'
else
  echo '[CONFIG] existing configuration preserved'
fi

install -o root -g root -m 0644 packaging/systemd/ywd-aprs.service /etc/systemd/system/ywd-aprs.service
/usr/local/bin/ywd-aprsd -config "$CONF_DIR/config.yaml" -check
systemctl daemon-reload
systemctl enable ywd-aprs.service
# Always restart after replacing the embedded daemon binary. `enable --now` alone
# does not restart an already-running service, which can leave the old daemon/UI
# serving even though ywd-aprsctl has been updated.
systemctl restart ywd-aprs.service
sleep 1

echo
echo '================================================'
echo ' YWD-APRS INSTALLED'
echo '================================================'
systemctl --no-pager --full status ywd-aprs.service | sed -n '1,12p' || true
ip="$(hostname -I 2>/dev/null | awk '{print $1}')"
echo "Web UI: http://${ip:-THIS-HOST}:8080"
echo 'Config: /etc/ywd-aprs/config.yaml'
echo 'Logs:   journalctl -fu ywd-aprs'
echo 'Status: ywd-aprsctl status'
echo
echo 'NOTE: This development gate is RX ONLY. No RF transmit path exists.'
