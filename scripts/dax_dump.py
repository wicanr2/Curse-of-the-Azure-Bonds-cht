#!/usr/bin/env python3
"""Inspect SSI Gold Box DAX containers and decode their RLE blocks.

This is a research tool, not the remake runtime.  It intentionally reports
malformed/truncated blocks instead of silently accepting them.
"""

from __future__ import annotations

import argparse
import pathlib
import struct
import zipfile
from dataclasses import dataclass


@dataclass(frozen=True)
class HeaderEntry:
    block_id: int
    offset: int
    raw_size: int
    packed_size: int


def read_source(path: pathlib.Path, member: str | None) -> bytes:
    if member is None:
        return path.read_bytes()
    with zipfile.ZipFile(path) as archive:
        return archive.read(member)


def parse_entries(data: bytes) -> tuple[int, list[HeaderEntry]]:
    if len(data) < 2:
        raise ValueError("file is shorter than DAX header")
    header_size = struct.unpack_from("<H", data)[0]
    data_offset = header_size + 2
    if data_offset > len(data) or (data_offset - 2) % 9:
        raise ValueError(f"invalid DAX header size: {header_size}")
    entries: list[HeaderEntry] = []
    for pos in range(2, data_offset, 9):
        block_id, offset, raw_size, packed_size = struct.unpack_from("<BIHH", data, pos)
        end = data_offset + offset + packed_size
        if end > len(data):
            raise ValueError(f"block {block_id} exceeds file boundary")
        entries.append(HeaderEntry(block_id, offset, raw_size, packed_size))
    return data_offset, entries


def decode_rle(packed: bytes, expected_size: int) -> bytes:
    output = bytearray()
    pos = 0
    while pos < len(packed) and len(output) < expected_size:
        control = struct.unpack_from("<b", packed, pos)[0]
        pos += 1
        if control >= 0:
            count = control + 1
            if pos + count > len(packed):
                raise ValueError("literal run exceeds packed block")
            output.extend(packed[pos : pos + count])
            pos += count
        else:
            count = -control
            if pos >= len(packed):
                raise ValueError("repeat run has no value byte")
            output.extend(packed[pos : pos + 1] * count)
            pos += 1
    if len(output) != expected_size:
        raise ValueError(f"decoded {len(output)} bytes, expected {expected_size}")
    return bytes(output)


def decode_ecl_text(data: bytes) -> list[str]:
    """Extract the 6-bit packed strings used inside ECL blocks."""
    alphabet = lambda value: chr(value + 0x40 if value <= 0x1F else value)
    strings: list[str] = []
    for pos in range(len(data) - 2):
        if data[pos] != 0x80:
            continue
        length = data[pos + 1]
        payload = data[pos + 2 : pos + 2 + length]
        if len(payload) != length:
            continue
        chars: list[str] = []
        state = 1
        previous = 0
        for current in payload:
            if state == 1:
                value = (current >> 2) & 0x3F
                if value:
                    chars.append(alphabet(value))
                state = 2
            elif state == 2:
                value = ((previous << 4) | (current >> 4)) & 0x3F
                if value:
                    chars.append(alphabet(value))
                state = 3
            else:
                value = ((previous << 2) | (current >> 6)) & 0x3F
                if value:
                    chars.append(alphabet(value))
                value = current & 0x3F
                if value:
                    chars.append(alphabet(value))
                state = 1
            previous = current
        text = "".join(chars).strip()
        if len(text) > 3 and text[0].isalpha() and any(char.isspace() for char in text):
            strings.append(text)
    return strings


def dump(path: pathlib.Path, member: str | None) -> None:
    data = read_source(path, member)
    data_offset, entries = parse_entries(data)
    print(f"source={path}{':' + member if member else ''}")
    print(f"data_offset={data_offset} blocks={len(entries)}")
    for entry in entries:
        start = data_offset + entry.offset
        packed = data[start : start + entry.packed_size]
        decoded = decode_rle(packed, entry.raw_size)
        texts = decode_ecl_text(decoded)
        print(
            f"block={entry.block_id} offset={entry.offset} "
            f"raw={entry.raw_size} packed={entry.packed_size} texts={len(texts)}"
        )
        for text in texts[:5]:
            print(f"  text={text!r}")


def main() -> int:
    parser = argparse.ArgumentParser()
    parser.add_argument("source", type=pathlib.Path)
    parser.add_argument("--member", help="read a member from a ZIP source")
    args = parser.parse_args()
    dump(args.source, args.member)
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
