#!/usr/bin/env python3
"""以原始 bytes 稽核 PC-98 overlay-14 的 MOVEPARTY map-buffer writer。

這是 IDA Pro 報告的可重現交叉驗證，不取代 IDA 的函式邊界／xref graph；工具只
讀取輸入，不寫入 binary、IDA database 或報告檔案。語意名稱只標示已由 Borland
symbol／raw operand 對上的邊界，不能把這個 writer 直接命名成秘密門規則。
"""

from __future__ import annotations

import argparse
import hashlib
import struct
import sys
from pathlib import Path


EXPECTED_SHA256 = (
    "a8e03ba9a5381c3a9f7ab411ced3262b21e0b65b948160d614386d677610e7b9"
)
MOVEPARTY = 0x0BCC
MAP_WRITER = 0x003E


def require_bytes(data: bytes, offset: int, expected: bytes, label: str) -> None:
    actual = data[offset : offset + len(expected)]
    if actual != expected:
        raise ValueError(
            f"{label}: offset=0x{offset:04X} expected={expected.hex()} actual={actual.hex()}"
        )
    print(f"exact_bytes label={label} local=0x{offset:04X} bytes={actual.hex()}")


def main() -> int:
    parser = argparse.ArgumentParser(description="PC-98 overlay-14 raw map writer audit")
    parser.add_argument("overlay", type=Path)
    args = parser.parse_args()
    data = args.overlay.read_bytes()
    digest = hashlib.sha256(data).hexdigest()
    print(f"input={args.overlay}")
    print(f"sha256={digest}")
    if digest != EXPECTED_SHA256:
        raise ValueError(f"輸入雜湊不符：期待 {EXPECTED_SHA256}")
    print(f"exact_file_hash sha256={digest}")

    require_bytes(data, MOVEPARTY, bytes.fromhex("5589e5"), "MOVEPARTY_entry")
    require_bytes(data, 0x0CD3, bytes.fromhex("0ee867f3"), "MOVEPARTY_to_local_writer")
    displacement = struct.unpack_from("<h", data, 0x0CD5)[0]
    call_target = 0x0CD7 + displacement
    print(
        "exact_near_call caller=overlay14:0x0CD4 "
        f"target=overlay14:0x{call_target:04X} displacement={displacement:+d}"
    )
    if call_target != MAP_WRITER:
        raise ValueError(f"near-call target 不符：0x{call_target:04X}")

    require_bytes(data, 0x0059, bytes.fromhex("c43ea0a2"), "THE3DMAP_pointer_load")
    require_bytes(data, 0x0061, bytes.fromhex("268a850003"), "third_plane_read")
    require_bytes(data, 0x0080, bytes.fromhex("26889d0003"), "third_plane_write_direction_6")
    require_bytes(data, 0x0066, bytes.fromhex("243f"), "direction_6_mask")
    require_bytes(data, 0x00A7, bytes.fromhex("24cf"), "direction_4_mask")
    require_bytes(data, 0x00E7, bytes.fromhex("24f3"), "direction_2_mask")
    require_bytes(data, 0x0127, bytes.fromhex("24fc"), "direction_0_mask")
    require_bytes(data, 0x0141, bytes.fromhex("26889d0003"), "third_plane_write_direction_0")
    require_bytes(data, 0x0146, bytes.fromhex("89ec5dca0600"), "map_writer_return")

    print(
        "exact_shape writer=overlay14:0x003E "
        "address=THE3DMAP+(arg_0A<<4)+arg_08+0x300 "
        "effect=selected_2bit_field_is_and_masked_to_zero"
    )
    print(
        "inference_boundary action_label=unknown "
        "secret_door_mapping=unknown "
        "runtime_persistence=unknown"
    )
    return 0


if __name__ == "__main__":
    try:
        raise SystemExit(main())
    except (OSError, ValueError) as error:
        print(f"audit_error={error}", file=sys.stderr)
        raise SystemExit(1)
