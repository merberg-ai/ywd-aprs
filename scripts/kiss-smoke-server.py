#!/usr/bin/env python3
"""Tiny RX test fixture: emits one known APRS AX.25 UI frame over TCP KISS."""
import socket,time
FEND=0xC0;FESC=0xDB;TFEND=0xDC;TFESC=0xDD

def addr(call,ssid,last=False):
    call=call.ljust(6)[:6]
    b=bytearray((ord(c)<<1)&0xfe for c in call)
    ss=0x60|((ssid&0xf)<<1)|(1 if last else 0)
    b.append(ss);return bytes(b)
def kiss(payload):
    body=bytes([0])+payload
    body=body.replace(bytes([FESC]),bytes([FESC,TFESC])).replace(bytes([FEND]),bytes([FESC,TFEND]))
    return bytes([FEND])+body+bytes([FEND])
raw=addr('APRS',0)+addr('KJ6YWD',9,last=True)+bytes([0x03,0xF0])+b'!4035.19N/12223.47W>YWD-APRS KISS SMOKE TEST'
frame=kiss(raw)
with socket.socket() as s:
    s.setsockopt(socket.SOL_SOCKET,socket.SO_REUSEADDR,1);s.bind(('0.0.0.0',8001));s.listen(1)
    print('KISS smoke server listening on :8001')
    while True:
        c,a=s.accept();print('client',a)
        with c:
            try:
                while True:c.sendall(frame);print('sent KJ6YWD-9 APRS position');time.sleep(3)
            except (BrokenPipeError,ConnectionResetError):pass
