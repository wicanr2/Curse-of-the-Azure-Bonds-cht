#!/usr/bin/env python3
"""唯讀稽核 GEO2 block 3 的 E2 候選路徑。

這不是 remake runtime，也不把 wall=09 自動命名成秘密門。工具只解碼指定
GEO record，分別列出「目前 movement contract」與「允許 wall=09 候選橋接」的
最短路徑，讓研究報告可以在 Docker 內重生同一組座標／位址空間證據。
"""

from __future__ import annotations

import argparse
import hashlib
import struct
import zipfile
from collections import deque
from dataclasses import dataclass
from pathlib import Path


WIDTH = 16
HEIGHT = 16
PAYLOAD_SIZE = 0x400
GEO_BLOCK_SIZE = 2 + PAYLOAD_SIZE
EXPECTED_ARCHIVE_SHA256 = (
    "c98698a6271c17177dfdb27f34b0389b7d34f58ef206e92575393f4655f5b26d"
)


@dataclass(frozen=True)
class Cell:
    walls: tuple[int, int, int, int]
    details: tuple[int, int, int, int]
    terrain: int


@dataclass(frozen=True)
class Step:
    source: tuple[int, int]
    target: tuple[int, int]
    direction: int
    wall: int
    detail: int
    other_wall: int
    other_detail: int


def decode_rle(packed: bytes, expected_size: int) -> bytes:
    output = bytearray()
    cursor = 0
    while cursor < len(packed) and len(output) < expected_size:
        control = struct.unpack_from("<b", packed, cursor)[0]
        cursor += 1
        if control >= 0:
            count = control + 1
            output.extend(packed[cursor : cursor + count])
            cursor += count
        else:
            count = -control
            if cursor >= len(packed):
                raise ValueError("GEO RLE repeat 缺少值 byte")
            output.extend(packed[cursor : cursor + 1] * count)
            cursor += 1
    if len(output) != expected_size:
        raise ValueError(f"GEO RLE decoded={len(output)} expected={expected_size}")
    return bytes(output)


def read_geo_block(archive: Path, block_id: int) -> bytes:
    with zipfile.ZipFile(archive) as source:
        raw = source.read("GEO2.DAX")
    header_size = struct.unpack_from("<H", raw)[0]
    data_offset = header_size + 2
    for offset in range(2, data_offset, 9):
        candidate_id, block_offset, raw_size, packed_size = struct.unpack_from(
            "<BIHH", raw, offset
        )
        if candidate_id != block_id:
            continue
        start = data_offset + block_offset
        packed = raw[start : start + packed_size]
        decoded = decode_rle(packed, raw_size)
        if len(decoded) != GEO_BLOCK_SIZE:
            raise ValueError(
                f"GEO2 block {block_id} size={len(decoded)} expected={GEO_BLOCK_SIZE}"
            )
        return decoded[2:]
    raise ValueError(f"GEO2.DAX 缺少 block {block_id}")


def parse_grid(payload: bytes) -> list[list[Cell]]:
    rows: list[list[Cell]] = []
    for y in range(HEIGHT):
        row: list[Cell] = []
        for x in range(WIDTH):
            index = y * WIDTH + x
            first = payload[index]
            second = payload[0x100 + index]
            extra = payload[0x300 + index]
            row.append(
                Cell(
                    walls=(
                        (first >> 4) & 0x0F,
                        first & 0x0F,
                        (second >> 4) & 0x0F,
                        second & 0x0F,
                    ),
                    details=(
                        extra & 0x03,
                        (extra >> 2) & 0x03,
                        (extra >> 4) & 0x03,
                        (extra >> 6) & 0x03,
                    ),
                    terrain=payload[0x200 + index],
                )
            )
        rows.append(row)
    return rows


def delta(direction: int) -> tuple[int, int]:
    return {0: (0, -1), 2: (1, 0), 4: (0, 1), 6: (-1, 0)}[direction]


def direction_name(direction: int) -> str:
    return {0: "N", 2: "E", 4: "S", 6: "W"}[direction]


def edge(grid: list[list[Cell]], x: int, y: int, direction: int) -> tuple[int, int, int]:
    index = {0: 0, 2: 1, 4: 2, 6: 3}[direction]
    cell = grid[y % HEIGHT][x % WIDTH]
    return cell.walls[index], cell.details[index], index


def shortest_path(
    grid: list[list[Cell]],
    start: tuple[int, int],
    target: tuple[int, int],
    allow_wall09: bool,
) -> list[Step] | None:
    queue = deque([start])
    previous: dict[tuple[int, int], tuple[tuple[int, int], Step]] = {}
    visited = {start}
    directions = (0, 2, 4, 6)
    while queue and target not in visited:
        source = queue.popleft()
        for direction in directions:
            dx, dy = delta(direction)
            destination = ((source[0] + dx) % WIDTH, (source[1] + dy) % HEIGHT)
            wall, detail, _ = edge(grid, *source, direction)
            other_wall, other_detail, _ = edge(grid, *destination, (direction + 4) % 8)
            passable = wall == 0 or detail == 1
            if not passable and allow_wall09 and (wall == 9 or other_wall == 9):
                passable = True
            if not passable or destination in visited:
                continue
            visited.add(destination)
            previous[destination] = (
                source,
                Step(source, destination, direction, wall, detail, other_wall, other_detail),
            )
            queue.append(destination)
    if target not in visited:
        return None
    result: list[Step] = []
    cursor = target
    while cursor != start:
        source, step = previous[cursor]
        result.append(step)
        cursor = source
    result.reverse()
    return result


def main() -> int:
    parser = argparse.ArgumentParser()
    parser.add_argument("archive", type=Path)
    parser.add_argument("--start", type=int, nargs=2, default=(13, 10), metavar=("X", "Y"))
    parser.add_argument("--target", type=int, nargs=2, default=(8, 15), metavar=("X", "Y"))
    args = parser.parse_args()
    digest = hashlib.sha256(args.archive.read_bytes()).hexdigest()
    print(f"input={args.archive}")
    print(f"sha256={digest}")
    if digest != EXPECTED_ARCHIVE_SHA256:
        raise ValueError(f"輸入雜湊不符：期待 {EXPECTED_ARCHIVE_SHA256}")
    grid = parse_grid(read_geo_block(args.archive, 3))
    start = tuple(args.start)
    target = tuple(args.target)
    for allow_wall09 in (False, True):
        path = shortest_path(grid, start, target, allow_wall09)
        print(f"allow_wall_09={str(allow_wall09).lower()} reachable={path is not None}")
        if path is None:
            continue
        for index, step in enumerate(path, start=1):
            print(
                f"{index:02d}: ({step.source[0]},{step.source[1]}) "
                f"--{direction_name(step.direction)} wall={step.wall:02X} "
                f"detail={step.detail} other_wall={step.other_wall:02X} "
                f"other_detail={step.other_detail}--> "
                f"({step.target[0]},{step.target[1]})"
            )
    boundary_wall, boundary_detail, _ = edge(grid, *target, 4)
    other_wall, other_detail, _ = edge(grid, target[0], 0, 0)
    print(
        f"e2_candidate_boundary=({target[0]},{target[1]},S) "
        f"wall={boundary_wall:02X} detail={boundary_detail} "
        f"other=( {target[0]},0,N ) wall={other_wall:02X} detail={other_detail}"
    )
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
