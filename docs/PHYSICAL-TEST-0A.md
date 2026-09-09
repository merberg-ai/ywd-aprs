# 0A/0B Physical Receive Test

This development checkpoint exercises the live receive chain:

`TCP KISS -> KISS framing -> AX.25 -> APRS -> station state -> packet log -> HTTP API -> browser map/monitor`

## Safety gate

This checkpoint is receive-only by construction:

- the KISS client exposes no transmit/write API;
- no beacon or message scheduler exists;
- `tx.enabled: true` is rejected during configuration validation;
- the UI explicitly reports `RX ONLY // TX PATH ABSENT`.

## Install on a Debian/Raspberry Pi OS machine

```bash
curl -fsSL https://raw.githubusercontent.com/merberg-ai/ywd-aprs/dev/scripts/install-dev.sh | sudo bash
```

The installer preserves an existing `/etc/ywd-aprs/config.yaml` on later runs.

Edit the TCP KISS endpoint if required:

```bash
sudo nano /etc/ywd-aprs/config.yaml
sudo ywd-aprsd -config /etc/ywd-aprs/config.yaml -check
sudo systemctl restart ywd-aprs
```

For the current YWD-MMDVM-TNC test host the expected settings are:

```yaml
kiss:
  host: 192.168.1.11
  port: 8001
  kiss_port: 0
```

## Observe

```bash
ywd-aprsctl doctor
journalctl -fu ywd-aprs
```

Open `http://<pi-address>:8080/` in a browser.

A successful live test should show:

1. `KISS ONLINE` in the header;
2. the RX frame counter increasing when 145.050 MHz traffic is received;
3. decoded source callsigns in **Last Heard**;
4. TNC2-formatted frames in **Packet Monitor**;
5. APRS position packets appearing on the map when a position frame is heard;
6. `/var/lib/ywd-aprs/packets.jsonl` accumulating normalized received packets;
7. zero transmitted frames, because no transmit path exists.

## Local smoke test without a TNC

A fixture is included to prove the complete software receive/UI path before attaching hardware. Stop YWD-APRS, run the fixture on port 8001, point `kiss.host` at the fixture machine, then start YWD-APRS again:

```bash
cd /opt/ywd-aprs-src
python3 scripts/kiss-smoke-server.py
```

It emits a synthetic `KJ6YWD-9` APRS position every three seconds. The browser should show the station near Redding and the RX counter should increment.

## Useful diagnostics

```bash
systemctl status ywd-aprs --no-pager
journalctl -u ywd-aprs -n 100 --no-pager
curl -s http://127.0.0.1:8080/api/v1/state | python3 -m json.tool
curl -s http://127.0.0.1:8080/api/v1/public-config | python3 -m json.tool
```
