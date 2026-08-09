#!/usr/bin/env python3
"""唯讀稽核 PC-98 overlay-14 的 MOVEPARTY action／map-field writer 邊界。

本工具只驗證既有 overlay raw bytes 與 near-call 目標，不修改 binary、IDA
database 或輸出檔。它保留 overlay-local 位址，不能把 raw 第三平面欄位命名成
秘密門、detail 或任何 ECL 旗標。
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


def require_bytes(data: bytes, offset: int, expected: bytes, label: str) -> None:
    actual = data[offset : offset + len(expected)]
    if actual != expected:
        raise ValueError(
            f"{label}: local=0x{offset:04X} expected={expected.hex()} actual={actual.hex()}"
        )
    print(f"exact_bytes label={label} local=0x{offset:04X} bytes={actual.hex()}")


def require_near_call(
    data: bytes, call_offset: int, target: int, label: str
) -> None:
    require_bytes(data, call_offset, b"\xE8", f"{label}_opcode")
    displacement = struct.unpack_from("<h", data, call_offset + 1)[0]
    actual_target = call_offset + 3 + displacement
    print(
        f"exact_near_call label={label} caller=overlay14:0x{call_offset:04X} "
        f"target=overlay14:0x{actual_target:04X} displacement={displacement:+d}"
    )
    if actual_target != target:
        raise ValueError(
            f"{label}: target=0x{actual_target:04X}, expected=0x{target:04X}"
        )


def main() -> int:
    parser = argparse.ArgumentParser(
        description="PC-98 overlay-14 MOVEPARTY action writer audit"
    )
    parser.add_argument("overlay", type=Path)
    args = parser.parse_args()

    data = args.overlay.read_bytes()
    digest = hashlib.sha256(data).hexdigest()
    print(f"input={args.overlay}")
    print(f"sha256={digest}")
    if digest != EXPECTED_SHA256:
        raise ValueError(f"輸入雜湊不符：期待 {EXPECTED_SHA256}")
    print(f"exact_file_hash sha256={digest}")

    # local 003E: clear the selected 2-bit field in THE3DMAP+300h.
    require_bytes(data, 0x003E, bytes.fromhex("5589e5"), "clear_writer_entry")
    require_bytes(data, 0x0059, bytes.fromhex("c43ea0a2"), "clear_writer_the3dmap")
    for offset, mask, direction in (
        (0x0066, "3f", "6"),
        (0x00A7, "cf", "4"),
        (0x00E7, "f3", "2"),
        (0x0127, "fc", "0"),
    ):
        require_bytes(data, offset, bytes.fromhex("24" + mask), f"clear_mask_dir_{direction}")
    require_bytes(data, 0x0146, bytes.fromhex("89ec5dca0600"), "clear_writer_return")

    # local 014C: set the same selected field to raw two-bit value 01.
    require_bytes(data, 0x014C, bytes.fromhex("5589e5"), "set_writer_entry")
    require_bytes(data, 0x017F, bytes.fromhex("c43ea0a2"), "set_writer_the3dmap")
    for offset, mask, value, direction in (
        (0x018C, "3f", "40", "6"),
        (0x01CF, "cf", "10", "4"),
        (0x0212, "f3", "04", "2"),
        (0x0254, "fc", "01", "0"),
    ):
        require_bytes(data, offset, bytes.fromhex("24" + mask), f"set_mask_dir_{direction}")
        require_bytes(data, offset + 2, bytes.fromhex("0c" + value), f"set_value_dir_{direction}")
    require_bytes(data, 0x0275, bytes.fromhex("89ec5dca0600"), "set_writer_return")
    print(
        "exact_shape set_writer=overlay14:0x014C "
        "address=THE3DMAP+(arg_0A<<4)+arg_08+0x300 "
        "effect=selected_2bit_field_raw_value_01"
    )

    # MOVEPARTY's visible action bytes and helper dispatches.
    require_bytes(data, 0x0BCC, bytes.fromhex("5589e5"), "moveparty_entry")
    for offset, token, label in (
        (0x0D1A, "42", "B_branch"),
        (0x0D27, "50", "P_branch"),
        (0x0D50, "4B", "K_branch"),
    ):
        require_bytes(data, offset, bytes.fromhex("3c" + token), f"{label}_token")
    require_near_call(data, 0x0D1F, 0x02F5, "B_helper")
    require_near_call(data, 0x0D41, 0x05B4, "P_helper")
    require_near_call(data, 0x0D55, 0x0714, "K_helper")

    # The movement-flow clear call is guarded by the local result flag at bp-3.
    require_near_call(data, 0x0CD4, 0x003E, "movement_result_clear")
    for offset, label in (
        (0x0566, "B_helper_first_set"),
        (0x05A4, "B_helper_second_set"),
        (0x062F, "P_helper_first_set"),
        (0x066D, "P_helper_second_set"),
    ):
        require_near_call(data, offset, 0x014C, label)
    print(
        "exact_shape action_helpers=B/P/K "
        "B_and_P_each_have_two_direct_set_writer_calls "
        "movement_flow_has_one_direct_clear_writer_call"
    )
    print(
        "inference_boundary action_success_predicate=unknown "
        "secret_door_mapping=unknown "
        "detail_semantics=unknown "
        "runtime_persistence=unknown"
    )
    return 0


if __name__ == "__main__":
    try:
        raise SystemExit(main())
    except (OSError, ValueError, struct.error) as error:
        print(f"audit_error={error}", file=sys.stderr)
        raise SystemExit(1)
