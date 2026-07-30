#!/usr/bin/env python3
"""建立執行 exact GAME.EXE speaker routine 的暫存 PC-98 D88。

輸出只供 Docker 內 NP2kai 動態研究。程式不內嵌商業 executable bytes；
輸入 GAME.EXE 與 D88 皆保持唯讀，只有指定輸出副本會被建立。
"""

import hashlib
import struct
import sys


GAME_SHA256 = "8bca0b50f47b5a41193584d3d4d1cd7361562ca3daf5360d3691620cc1b752c0"
GAME_ROUTINE_FILE_OFFSET = 0xA6BE
GAME_ROUTINE_SIZE = 64
GAME_PERIOD_OFFSET = 0xBE56

D88_HEADER_SIZE = 0x2B0
D88_TRACK_ZERO_OFFSET_FIELD = 0x20
D88_SECTOR_HEADER_SIZE = 16
D88_EXPECTED_SECTOR_SIZE = 1024
PC98_IPL_ENTRY_OFFSET = 0x110


def read_exact_game(path):
    with open(path, "rb") as source:
        game = source.read()
    digest = hashlib.sha256(game).hexdigest()
    if digest != GAME_SHA256:
        raise SystemExit(f"{path}: SHA-256 {digest}，預期 {GAME_SHA256}")
    end = GAME_ROUTINE_FILE_OFFSET + GAME_ROUTINE_SIZE
    if end > len(game):
        raise SystemExit("GAME.EXE speaker routine 超出檔案")
    routine = game[GAME_ROUTINE_FILE_OFFSET:end]
    if routine[:10] != bytes.fromhex("56 55 8b ec 83 7e 08 00 74 31"):
        raise SystemExit("GAME.EXE speaker routine 開頭不符")
    if routine[-3:] != bytes.fromhex("ca 02 00"):
        raise SystemExit("GAME.EXE speaker routine 結尾不符")
    return routine


def first_sector_payload(disk):
    if len(disk) < D88_HEADER_SIZE:
        raise SystemExit("D88 小於固定 header")
    track_offset = struct.unpack_from("<I", disk, D88_TRACK_ZERO_OFFSET_FIELD)[0]
    if track_offset < D88_HEADER_SIZE or track_offset + D88_SECTOR_HEADER_SIZE > len(disk):
        raise SystemExit("D88 track 0 offset 無效")
    header = disk[track_offset : track_offset + D88_SECTOR_HEADER_SIZE]
    if tuple(header[:4]) != (0, 0, 1, 3):
        raise SystemExit(
            f"D88 第一磁區 CHRN={tuple(header[:4])}，預期 (0, 0, 1, 3)"
        )
    size = struct.unpack_from("<H", header, 14)[0]
    if size != D88_EXPECTED_SECTOR_SIZE:
        raise SystemExit(f"D88 第一磁區大小 {size}，預期 {D88_EXPECTED_SECTOR_SIZE}")
    payload = track_offset + D88_SECTOR_HEADER_SIZE
    if payload + size > len(disk):
        raise SystemExit("D88 第一磁區 payload 超出檔案")
    return payload, size


def build_boot_sector(base_sector, routine, period, pulses):
    if not 1 <= period <= 0xFFFF:
        raise SystemExit("PERIOD 必須介於 1..65535")
    if not 1 <= pulses <= 0xFFFF:
        raise SystemExit("PULSES 必須介於 1..65535")

    # BIOS 已把第一磁區載入 CS:0000。DS 設成 CS-0x0BC0，讓原 routine 的
    # DS:BE56 指到同一 sector 的 CS:0256；routine 本身不需改一個 byte。
    code = bytearray(
        bytes.fromhex(
            "fa "          # cli
            "b0 04 e6 37 " # probe wrapper entry marker
            "0e 58 "       # push cs / pop ax
            "8e d0 "       # mov ss,ax
            "bc f0 03 "    # mov sp,03f0
            "2d c0 0b "    # sub ax,0bc0
            "8e d8 "       # mov ds,ax
            "c7 06 56 be"  # mov word ptr ds:be56,period
        )
    )
    code.extend(struct.pack("<H", period))
    code.extend(bytes.fromhex("b8"))
    code.extend(struct.pack("<H", pulses))
    code.extend(
        bytes.fromhex(
            "50 "          # push ax: FAR argument
            "b0 05 e6 37 " # marker immediately before exact routine
            "0e "          # push cs: FAR return segment
            "b8"
        )
    )
    return_immediate = len(code)
    code.extend(b"\x00\x00")
    code.extend(bytes.fromhex("50 e9"))
    jump_displacement = len(code)
    code.extend(b"\x00\x00")

    routine_offset = PC98_IPL_ENTRY_OFFSET + len(code)
    code.extend(routine)
    return_offset = PC98_IPL_ENTRY_OFFSET + len(code)
    code.extend(bytes.fromhex("fa f4 eb fd"))

    struct.pack_into("<H", code, return_immediate, return_offset)
    relative = routine_offset - (
        PC98_IPL_ENTRY_OFFSET + jump_displacement + 2
    )
    struct.pack_into("<h", code, jump_displacement, relative)

    if len(base_sector) != D88_EXPECTED_SECTOR_SIZE:
        raise SystemExit("D88 base IPL sector 大小不符")
    if base_sector[:3] != bytes.fromhex("e9 0d 01"):
        raise SystemExit("D88 base IPL 未跳到 0x0110")
    sector = bytearray(base_sector)
    end = PC98_IPL_ENTRY_OFFSET + len(code)
    if end > len(sector):
        raise SystemExit("speaker probe 超出 IPL sector")
    sector[PC98_IPL_ENTRY_OFFSET:end] = code
    struct.pack_into("<H", sector, 0x256, period)
    return sector, routine_offset, return_offset


def main():
    if len(sys.argv) != 6:
        raise SystemExit(
            "用法：pc98_speaker_probe_disk.py "
            "GAME.EXE BASE.d88 OUTPUT.d88 PERIOD PULSES"
        )
    routine = read_exact_game(sys.argv[1])
    with open(sys.argv[2], "rb") as source:
        disk = bytearray(source.read())
    payload, size = first_sector_payload(disk)
    sector, routine_offset, return_offset = build_boot_sector(
        disk[payload : payload + size],
        routine,
        int(sys.argv[4], 0),
        int(sys.argv[5], 0),
    )
    disk[payload : payload + size] = sector
    with open(sys.argv[3], "wb") as output:
        output.write(disk)

    print(
        f"輸出={sys.argv[3]} "
        f"routine=CS:{routine_offset:04X} "
        f"return=CS:{return_offset:04X} "
        f"sha256={hashlib.sha256(disk).hexdigest()}"
    )


if __name__ == "__main__":
    main()
