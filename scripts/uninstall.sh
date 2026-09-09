#!/usr/bin/env bash
set -euo pipefail
if [[ ${EUID:-$(id -u)} -ne 0 ]]; then echo 'Run with sudo.' >&2; exit 1; fi
systemctl disable --now ywd-aprs.service 2>/dev/null || true
rm -f /etc/systemd/system/ywd-aprs.service /usr/local/bin/ywd-aprsd /usr/local/bin/ywd-aprsctl /usr/local/bin/ywd-aprsd.previous
systemctl daemon-reload
echo 'YWD-APRS binaries/service removed. Configuration and /var/lib/ywd-aprs were preserved.'
