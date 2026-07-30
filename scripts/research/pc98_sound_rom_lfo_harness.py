#!/usr/bin/env python3
"""直接執行 PC-98 SOUND.ROM 的 Timer B 軟體 LFO 路徑。

此 script 不含商業 ROM、driver 或音色資料。輸入必須是使用者本機的 exact
SOUND.ROM 與 MSCDRV.EXE；執行環境需安裝 Unicorn 2.1.4，且依 AGENTS.md
只能在有界 Docker 容器內執行。
"""

import hashlib
import json
import struct
import sys

from unicorn import Uc, UC_ARCH_X86, UC_HOOK_CODE, UC_HOOK_INSN, UC_MODE_16
from unicorn.x86_const import (
    UC_X86_INS_IN,
    UC_X86_INS_OUT,
    UC_X86_REG_AX,
    UC_X86_REG_BX,
    UC_X86_REG_CS,
    UC_X86_REG_CX,
    UC_X86_REG_DI,
    UC_X86_REG_DS,
    UC_X86_REG_DX,
    UC_X86_REG_ES,
    UC_X86_REG_FLAGS,
    UC_X86_REG_IP,
    UC_X86_REG_SI,
    UC_X86_REG_SP,
    UC_X86_REG_SS,
)

ROM_SHA256 = "f05b508d49f31f2a1a61724f013572592abc0833c09c45a72180e84247dc0d0d"
DRIVER_SHA256 = "bddbe63a90078bd9c8c8da5c45417c7ec3afcdf7fd5b724877a83ad9bb7b12f5"
ROM_LINEAR = 0xCC000
ROM_SEGMENT = 0xCEE0
ROM_SEGMENT_LINEAR = ROM_SEGMENT << 4
WORK_SEGMENT = 0x2000
PARAM_SEGMENT = 0x3000
STACK_SEGMENT = 0x4000
SENTINEL_OFFSET = 0x3FFE
SENTINEL_LINEAR = ROM_SEGMENT_LINEAR + SENTINEL_OFFSET
PARAMETER_3_FILE_OFFSET = 0x45A2 + 3 * 100

TARGETS = {
    "set_touch": 0xCF1E8,
    "note": 0xCF239,
    "set_parameter": 0xCF309,
    "modulation_on": 0xCF3E7,
    "set_volume": 0xCF41F,
    "command_return": 0xCF479,
    "lfo_timer_b": 0xCF5F3,
}


class Harness:
    def __init__(self, rom, parameter):
        self.uc = Uc(UC_ARCH_X86, UC_MODE_16)
        self.uc.mem_map(0, 0x100000)
        self.uc.mem_write(ROM_LINEAR, rom)
        self.uc.mem_write(SENTINEL_LINEAR, b"\xF4")
        self.uc.mem_write(PARAM_SEGMENT << 4, parameter)
        self.selected_register = 0
        self.writes = []
        self.tick = -1
        self.uc.hook_add(UC_HOOK_INSN, self._in, None, 1, 0, UC_X86_INS_IN)
        self.uc.hook_add(UC_HOOK_INSN, self._out, None, 1, 0, UC_X86_INS_OUT)
        self.uc.hook_add(UC_HOOK_CODE, self._code)
        self.uc.reg_write(UC_X86_REG_CS, ROM_SEGMENT)
        self.uc.reg_write(UC_X86_REG_DS, WORK_SEGMENT)
        self.uc.reg_write(UC_X86_REG_ES, PARAM_SEGMENT)
        self.uc.reg_write(UC_X86_REG_SS, STACK_SEGMENT)
        self.uc.reg_write(UC_X86_REG_FLAGS, 0x0202)

    def _code(self, uc, address, _size, _data):
        if address == SENTINEL_LINEAR:
            uc.emu_stop()

    def _in(self, _uc, _port, _size, _data):
        return 0

    def _out(self, _uc, port, _size, value, _data):
        if port == 0x188:
            self.selected_register = value & 0xFF
        elif port == 0x18A:
            self.writes.append(
                (self.tick, self.selected_register, value & 0xFF)
            )

    def set_bytes(self, al=None, bl=None, bh=None, dl=None):
        ax = self.uc.reg_read(UC_X86_REG_AX)
        bx = self.uc.reg_read(UC_X86_REG_BX)
        dx = self.uc.reg_read(UC_X86_REG_DX)
        if al is not None:
            ax = (ax & 0xFF00) | al
        if bl is not None:
            bx = (bx & 0xFF00) | bl
        if bh is not None:
            bx = (bx & 0x00FF) | (bh << 8)
        if dl is not None:
            dx = (dx & 0xFF00) | dl
        self.uc.reg_write(UC_X86_REG_AX, ax)
        self.uc.reg_write(UC_X86_REG_BX, bx)
        self.uc.reg_write(UC_X86_REG_DX, dx)

    def call_near(self, absolute):
        self.uc.reg_write(UC_X86_REG_CS, ROM_SEGMENT)
        self.uc.reg_write(UC_X86_REG_IP, absolute - ROM_SEGMENT_LINEAR)
        self.uc.reg_write(UC_X86_REG_DS, WORK_SEGMENT)
        sp = 0xFE00
        self.uc.mem_write(
            (STACK_SEGMENT << 4) + sp, struct.pack("<H", SENTINEL_OFFSET)
        )
        self.uc.reg_write(UC_X86_REG_SP, sp)
        self.uc.emu_start(absolute, SENTINEL_LINEAR + 1, count=2_000_000)

    def timer_b_tick(self, tick):
        self.tick = tick
        for register in (
            UC_X86_REG_AX,
            UC_X86_REG_BX,
            UC_X86_REG_CX,
            UC_X86_REG_DX,
            UC_X86_REG_SI,
            UC_X86_REG_DI,
        ):
            self.uc.reg_write(register, 0)
        self.uc.reg_write(UC_X86_REG_CS, ROM_SEGMENT)
        self.uc.reg_write(
            UC_X86_REG_IP, TARGETS["lfo_timer_b"] - ROM_SEGMENT_LINEAR
        )
        self.uc.reg_write(UC_X86_REG_DS, WORK_SEGMENT)
        self.uc.reg_write(UC_X86_REG_ES, 0)
        sp = 0xFD00
        # Pop order at CF4C3: DS, ES, BP, SI, DI, DX, CX, BX, AX, then IRET.
        frame = [
            WORK_SEGMENT, 0, 0, 0, 0, 0, 0, 0, 0,
            SENTINEL_OFFSET, ROM_SEGMENT, 0x0202,
        ]
        self.uc.mem_write(
            (STACK_SEGMENT << 4) + sp, struct.pack("<12H", *frame)
        )
        self.uc.reg_write(UC_X86_REG_SP, sp)
        self.uc.emu_start(
            TARGETS["lfo_timer_b"], SENTINEL_LINEAR + 1, count=2_000_000
        )

    def state(self):
        raw = self.uc.mem_read(WORK_SEGMENT << 4, 0x20)
        return {
            "note_state": raw[0x14],
            "modulation_flags": raw[0x18],
            "phase_counter": struct.unpack_from("<H", raw, 0x19)[0],
            "effective_pitch_depth": struct.unpack_from("<H", raw, 0x1B)[0],
            "base_fnumber": struct.unpack_from("<H", raw, 0x1D)[0],
        }


def exact_input(path, expected):
    raw = open(path, "rb").read()
    digest = hashlib.sha256(raw).hexdigest()
    if digest != expected:
        raise SystemExit(f"{path}: SHA-256 {digest}，預期 {expected}")
    return raw


def main():
    if len(sys.argv) != 3:
        raise SystemExit("用法：pc98_sound_rom_lfo_harness.py SOUND.ROM MSCDRV.EXE")
    rom = exact_input(sys.argv[1], ROM_SHA256)
    driver = exact_input(sys.argv[2], DRIVER_SHA256)
    if rom[TARGETS["command_return"] - ROM_LINEAR] != 0xC3:
        raise SystemExit("SOUND.ROM command return anchor 不是 RET")
    parameter = driver[PARAMETER_3_FILE_OFFSET:PARAMETER_3_FILE_OFFSET + 100]
    harness = Harness(rom, parameter)

    harness.set_bytes(al=0, dl=0)
    harness.uc.reg_write(UC_X86_REG_BX, 0)
    harness.call_near(TARGETS["set_parameter"])
    harness.set_bytes(al=0, bl=112)
    harness.call_near(TARGETS["set_volume"])
    harness.set_bytes(al=0, bl=7)
    harness.call_near(TARGETS["set_touch"])
    harness.set_bytes(al=0, bh=0x30, bl=0x40)
    harness.call_near(TARGETS["note"])
    harness.set_bytes(al=0)
    harness.call_near(TARGETS["modulation_on"])
    initial = harness.state()

    start = len(harness.writes)
    for tick in range(1, 81):
        harness.timer_b_tick(tick)
    dynamic = harness.writes[start:]
    pitch = [event for event in dynamic if event[1] in (0xA4, 0xA0)]
    levels = [event for event in dynamic if event[1] in (0x40, 0x44, 0x48, 0x4C)]
    result = {
        "state_after_note": initial,
        "state_after_80_timer_b_ticks": harness.state(),
        "pitch_write_count": len(pitch),
        "level_write_count": len(levels),
        "first_pitch_tick": pitch[0][0] if pitch else None,
        "first_level_tick": levels[0][0] if levels else None,
    }
    print(json.dumps(result, ensure_ascii=False, indent=2))
    if (
        result["first_pitch_tick"] != 30
        or result["first_level_tick"] != 30
        or result["pitch_write_count"] != 102
        or result["level_write_count"] != 204
        or result["state_after_80_timer_b_ticks"]["phase_counter"] != 51
    ):
        raise SystemExit("ROM LFO regression mismatch")


if __name__ == "__main__":
    main()
