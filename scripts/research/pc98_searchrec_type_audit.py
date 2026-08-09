#!/usr/bin/env python3
"""唯讀解析 PC-98 Borland legacy debug table 的 SEARCHREC 欄位。

這個工具只驗證 type／member／symbol bytes；它不把 `SEARCH`、`SECRET` 或
`HIDDEN` 等名稱推論成地圖規則，也不修改 executable、IDA database 或輸出檔。
"""

from __future__ import annotations

import argparse
import hashlib
import struct
import sys
from pathlib import Path


EXPECTED_SHA256 = (
    "8bca0b50f47b5a41193584d3d4d1cd7361562ca3daf5360d3691620cc1b752c0"
)


def pascal_names(data: bytes, start: int, size: int) -> list[str]:
    names: list[str] = []
    end = start + size
    cursor = start
    while cursor < end:
        terminator = data.find(b"\0", cursor, end)
        if terminator < 0:
            raise ValueError("name pool 缺少 NUL terminator")
        names.append(data[cursor:terminator].decode("latin1"))
        cursor = terminator + 1
    return names


def name(names: list[str], index: int) -> str:
    return names[index - 1] if 0 < index <= len(names) else ""


def main() -> int:
    parser = argparse.ArgumentParser(description="PC-98 Borland SEARCHREC type audit")
    parser.add_argument("executable", type=Path)
    args = parser.parse_args()
    data = args.executable.read_bytes()
    digest = hashlib.sha256(data).hexdigest()
    print(f"input={args.executable}")
    print(f"sha256={digest}")
    if digest != EXPECTED_SHA256:
        raise ValueError(f"輸入雜湊不符：期待 {EXPECTED_SHA256}")

    last_page_bytes, pages = struct.unpack_from("<HH", data, 2)
    image_size = (pages - 1) * 512 + last_page_bytes if last_page_bytes else pages * 512
    if data[image_size : image_size + 2] != bytes.fromhex("fb52"):
        raise ValueError("MZ image boundary 沒有 Borland 0x52FB header")
    header = data[image_size : image_size + 0x30]
    name_pool_size = struct.unpack_from("<I", header, 4)[0]
    (
        name_count,
        type_count,
        member_count,
        symbol_count,
        _global_count,
        module_count,
        _local_count,
        scope_count,
        line_count,
        source_count,
        segment_count,
        correlation_count,
    ) = struct.unpack_from("<12H", header, 8)
    name_start = len(data) - name_pool_size
    names = pascal_names(data, name_start, name_pool_size)
    symbol_start = image_size + 0x30
    symbol_end = symbol_start + symbol_count * 9
    module_end = symbol_end + module_count * 16
    source_end = module_end + source_count * 6
    scope_end = source_end + scope_count * 12
    line_end = scope_end + line_count * 4
    segment_end = line_end + segment_count * 16
    correlation_end = segment_end + correlation_count * 8
    type_start = correlation_end
    type_end = type_start + type_count * 8
    data_pool_start = name_start - struct.unpack_from("<H", header, 43)[0]
    member_end = data_pool_start
    if member_end - type_end != member_count * 5:
        raise ValueError("member table 長度與 header 不一致")

    types: dict[int, tuple[int, str, int, bytes]] = {}
    for index in range(1, type_count + 1):
        offset = type_start + (index - 1) * 8
        type_id, name_index, size = struct.unpack_from("<BHH", data, offset)
        types[index] = (type_id, name(names, name_index), size, data[offset + 5 : offset + 8])

    members: dict[int, tuple[int, str, int]] = {}
    for index in range(member_count):
        offset = type_end + index * 5
        flags, name_index, type_index = struct.unpack_from("<BHH", data, offset)
        members[index] = (flags, name(names, name_index), type_index)

    searchrec = types[1458]
    if searchrec[2] != 0x2B or searchrec[3] != bytes.fromhex("005602"):
        raise ValueError(f"SEARCHREC type bytes 不符：{searchrec!r}")
    print(
        "exact_type name=SEARCHREC index=1458 id=0x%02X size=0x%04X detail=%s"
        % (searchrec[0], searchrec[2], searchrec[3].hex())
    )

    expected_members = [
        ("FILL", 1459),
        ("ATTR", 8),
        ("TIME", 6),
        ("SIZE", 6),
        ("NAME", 1463),
    ]
    for index, (expected_name, expected_type) in enumerate(expected_members, start=597):
        actual = members[index]
        if actual[1] != expected_name or actual[2] != expected_type:
            raise ValueError(f"SEARCHREC member {index} 不符：{actual!r}")
        field_type = types[actual[2]]
        print(
            f"exact_member index={index} name={actual[1]} type={actual[2]} "
            f"size=0x{field_type[2]:04X} flags=0x{actual[0]:02X}"
        )

    wanted_symbols = {"FINDFIRST": (0x08BD, 0x0112), "FINDNEXT": (0x08BD, 0x0150)}
    found: dict[str, tuple[int, int]] = {}
    for index in range(symbol_count):
        offset = symbol_start + index * 9
        name_index, type_index, symbol_offset, segment, flags = struct.unpack_from(
            "<HHHHB", data, offset
        )
        symbol_name = name(names, name_index)
        if symbol_name in wanted_symbols:
            found[symbol_name] = (segment, symbol_offset)
            print(
                f"exact_symbol name={symbol_name} address={segment:04X}:{symbol_offset:04X} "
                f"type={type_index} flags=0x{flags:02X}"
            )
    if found != wanted_symbols:
        raise ValueError(f"FINDFIRST/FINDNEXT symbols 不完整：{found!r}")
    print("inference_boundary map_search_record=not_proven file_search_record=confirmed_by_assembly")
    return 0


if __name__ == "__main__":
    try:
        raise SystemExit(main())
    except (OSError, ValueError, struct.error) as error:
        print(f"audit_error={error}", file=sys.stderr)
        raise SystemExit(1)
