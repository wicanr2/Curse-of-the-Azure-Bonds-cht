#!/usr/bin/env python3
"""唯讀稽核 PC-98 overlay-14 MOVEPARTY 的 result/action continuation。

本工具只檢查可回查的 raw bytes、far-call 位址與 near-call 位址，不修改
executable、overlay 或 IDA database。它刻意不把 result、B/P/K 或第三平面
field 命名成秘密門、開門、技能、ECL flag 或正式地圖語意。
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


def require_near_call(data: bytes, offset: int, target: int, label: str) -> None:
    require_bytes(data, offset, b"\xE8", f"{label}_opcode")
    displacement = struct.unpack_from("<h", data, offset + 1)[0]
    actual_target = offset + 3 + displacement
    print(
        f"exact_near_call label={label} caller=overlay14:0x{offset:04X} "
        f"target=overlay14:0x{actual_target:04X} displacement={displacement:+d}"
    )
    if actual_target != target:
        raise ValueError(
            f"{label}: target=0x{actual_target:04X}, expected=0x{target:04X}"
        )


def require_far_call(
    data: bytes, offset: int, segment: int, target: int, label: str
) -> None:
    require_bytes(data, offset, b"\x9A", f"{label}_opcode")
    actual_target = struct.unpack_from("<H", data, offset + 1)[0]
    actual_segment = struct.unpack_from("<H", data, offset + 3)[0]
    print(
        f"exact_far_call label={label} caller=overlay14:0x{offset:04X} "
        f"target={actual_segment:04X}:{actual_target:04X}"
    )
    if (actual_segment, actual_target) != (segment, target):
        raise ValueError(
            f"{label}: target={actual_segment:04X}:{actual_target:04X}, "
            f"expected={segment:04X}:{target:04X}"
        )


def main() -> int:
    parser = argparse.ArgumentParser(
        description="PC-98 overlay-14 MOVEPARTY action transaction audit"
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

    # The pre-action map-cell service is a far call in a separate address space.
    # Do not equate 017C:0039 with overlay-30 local 0039 or BLOCKCODE.
    require_far_call(data, 0x0C06, 0x017C, 0x0039, "first_map_cell_service")
    require_bytes(data, 0x0C0B, bytes.fromhex("3c01"), "result_one_compare")
    require_bytes(data, 0x0C0F, bytes.fromhex("c646ff01"), "result_one_success_flag")
    require_bytes(data, 0x0C16, bytes.fromhex("3c02"), "result_two_compare")
    require_bytes(data, 0x0C1A, bytes.fromhex("3c03"), "result_three_compare")
    require_far_call(data, 0x0C21, 0x0164, 0x002F, "result_two_three_followup")

    # Dynamic action input and the three visible token comparisons.
    require_far_call(data, 0x0D0F, 0x0164, 0x0039, "action_input_service")
    for offset, token, label in (
        (0x0D1A, "42", "B_token"),
        (0x0D27, "50", "P_token"),
        (0x0D50, "4B", "K_token"),
    ):
        require_bytes(data, offset, bytes.fromhex("3c" + token), label)
    require_near_call(data, 0x0D1F, 0x02F5, "B_helper")

    # P performs another map-cell service call and only continues to its helper
    # when that service returns AL=2; the semantic name of the return is unknown.
    require_far_call(data, 0x0D37, 0x017C, 0x0039, "P_map_cell_service")
    require_bytes(data, 0x0D3C, bytes.fromhex("3c02"), "P_result_two_compare")
    require_near_call(data, 0x0D41, 0x05B4, "P_helper")
    require_near_call(data, 0x0D55, 0x0714, "K_helper")

    # All action outcomes converge here. A nonzero helper result calls local
    # 0807; the far call at 014A:00DE follows the branch regardless. Neither
    # call is assigned a product-level name by this audit.
    require_bytes(data, 0x0D5B, bytes.fromhex("807eff007404"), "common_result_test")
    require_near_call(data, 0x0D62, 0x0807, "nonzero_result_followup")
    require_far_call(data, 0x0D65, 0x014A, 0x00DE, "common_action_exit_service")

    print(
        "exact_shape result=1_sets_local_success; result=2_or_3_enters_action_path; "
        "P_path_requires_second_service_result_2; nonzero_helper_calls_local_0807; "
        "all_paths_reach_014A:00DE"
    )
    print(
        "inference_boundary action_success_predicate=unknown "
        "secret_door_mapping=unknown detail_semantics=unknown "
        "ECL_effect=unknown runtime_persistence=unknown"
    )
    return 0


if __name__ == "__main__":
    try:
        raise SystemExit(main())
    except (OSError, ValueError, struct.error) as error:
        print(f"audit_error={error}", file=sys.stderr)
        raise SystemExit(1)
