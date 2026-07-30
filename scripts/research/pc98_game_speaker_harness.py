#!/usr/bin/env python3
"""直接執行 PC-98 GAME.EXE 的 port 37h speaker pulse routine。

此 script 不含商業 executable bytes。輸入必須是使用者本機的 exact
GAME.EXE；執行環境需安裝 Unicorn 2.1.4，且依 AGENTS.md 只能在有界
Docker 容器內執行。
"""

import hashlib
import json
import struct
import sys

from unicorn import Uc, UC_ARCH_X86, UC_HOOK_CODE, UC_HOOK_INSN, UC_MODE_16
from unicorn.x86_const import (
    UC_X86_INS_OUT,
    UC_X86_REG_CS,
    UC_X86_REG_DS,
    UC_X86_REG_FLAGS,
    UC_X86_REG_IP,
    UC_X86_REG_SP,
    UC_X86_REG_SS,
)

GAME_SHA256 = "8bca0b50f47b5a41193584d3d4d1cd7361562ca3daf5360d3691620cc1b752c0"
IDA_LOAD_BASE = 0x10000
ENTRY = 0x19D1E
STOP = 0x19D5B
LOOP_ON = 0x19D3E
LOOP_OFF = 0x19D4D
SOUND_WORD = 0x280E6
STACK_SEGMENT = 0x4000
STACK_POINTER = 0xFE00


def exact_game(path):
    raw = open(path, "rb").read()
    digest = hashlib.sha256(raw).hexdigest()
    if digest != GAME_SHA256:
        raise SystemExit(f"{path}: SHA-256 {digest}，預期 {GAME_SHA256}")
    if raw[:2] != b"MZ":
        raise SystemExit("GAME.EXE 缺少 MZ signature")
    header_size = struct.unpack_from("<H", raw, 8)[0] * 16
    if header_size <= 0 or header_size >= len(raw):
        raise SystemExit(f"無效 MZ header size：{header_size}")
    return raw, header_size


class Harness:
    def __init__(self, game, header_size, period, pulses):
        self.uc = Uc(UC_ARCH_X86, UC_MODE_16)
        self.uc.mem_map(0, 0x100000)
        self.uc.mem_write(IDA_LOAD_BASE, game[header_size:])
        self.uc.mem_write(SOUND_WORD, struct.pack("<H", period))

        self.events = []
        self.instruction_count = 0
        self.loop_on_count = 0
        self.loop_off_count = 0
        self.last_address = 0

        self.uc.hook_add(UC_HOOK_CODE, self._code)
        self.uc.hook_add(UC_HOOK_INSN, self._out, None, 1, 0, UC_X86_INS_OUT)

        self.uc.reg_write(UC_X86_REG_CS, 0x1000)
        self.uc.reg_write(UC_X86_REG_IP, ENTRY - IDA_LOAD_BASE)
        self.uc.reg_write(UC_X86_REG_DS, 0x1C29)
        self.uc.reg_write(UC_X86_REG_SS, STACK_SEGMENT)
        self.uc.reg_write(UC_X86_REG_SP, STACK_POINTER)
        self.uc.reg_write(UC_X86_REG_FLAGS, 0x0202)

        # FAR CALL entry layout: return IP, return CS, then one WORD argument.
        self.uc.mem_write(
            (STACK_SEGMENT << 4) + STACK_POINTER,
            struct.pack("<3H", 0, 0, pulses),
        )

    def _code(self, uc, address, _size, _data):
        if address == STOP:
            uc.emu_stop()
            return
        self.last_address = address
        self.instruction_count += 1
        if address == LOOP_ON:
            self.loop_on_count += 1
        elif address == LOOP_OFF:
            self.loop_off_count += 1

    def _out(self, _uc, port, size, value, _data):
        self.events.append(
            {
                "address": f"0x{self.last_address:05X}",
                "port": port,
                "size": size,
                "value": value & 0xFF,
                "instruction_index": self.instruction_count,
                "loop_on_total": self.loop_on_count,
                "loop_off_total": self.loop_off_count,
            }
        )

    def run(self):
        self.uc.emu_start(ENTRY, STOP + 1, count=20_000_000)


def main():
    if len(sys.argv) != 4:
        raise SystemExit(
            "用法：pc98_game_speaker_harness.py GAME.EXE PERIOD PULSES"
        )
    game, header_size = exact_game(sys.argv[1])
    period = int(sys.argv[2], 0)
    pulses = int(sys.argv[3], 0)
    if period <= 0 or period > 0xFFFF:
        raise SystemExit("PERIOD 必須介於 1..65535")
    if pulses <= 0 or pulses > 0xFFFF:
        raise SystemExit("PULSES 必須介於 1..65535")

    harness = Harness(game, header_size, period, pulses)
    harness.run()
    expected_outs = pulses * 2 + 2
    if len(harness.events) != expected_outs:
        raise SystemExit(
            f"OUT count {len(harness.events)}，預期 {expected_outs}"
        )
    if harness.loop_on_count != period * pulses:
        raise SystemExit(
            f"on LOOP count {harness.loop_on_count}，預期 {period * pulses}"
        )
    if harness.loop_off_count != period * pulses:
        raise SystemExit(
            f"off LOOP count {harness.loop_off_count}，預期 {period * pulses}"
        )

    print(
        json.dumps(
            {
                "game_sha256": GAME_SHA256,
                "mz_header_size": header_size,
                "entry": f"0x{ENTRY:05X}",
                "stop_before_retf": f"0x{STOP:05X}",
                "period": period,
                "pulses": pulses,
                "instruction_count": harness.instruction_count,
                "loop_on_count": harness.loop_on_count,
                "loop_off_count": harness.loop_off_count,
                "events": harness.events,
            },
            ensure_ascii=False,
            indent=2,
        )
    )


if __name__ == "__main__":
    main()
