#!/usr/bin/env python3
"""唯讀列出指定 GEO set／block 的可行走元件與邊界候選。

這個工具只使用目前已核對的雙側 wall/detail movement contract；不會把
wall=09、detail=0 或任何邊界自動命名成秘密門。輸出供後續正常玩家路徑
測試選擇座標，語意仍需回到 ECL／runtime 驗證。
"""

from __future__ import annotations

import argparse
import hashlib
import struct
import zipfile
from collections import deque
from pathlib import Path

WIDTH = 16
HEIGHT = 16
GEO_BLOCK_SIZE = 0x402


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
            output.extend(packed[cursor : cursor + 1] * count)
            cursor += 1
    if len(output) != expected_size:
        raise ValueError(f"decoded={len(output)} expected={expected_size}")
    return bytes(output)


def read_block(archive: Path, geo_set: int, block_id: int) -> bytes:
    with zipfile.ZipFile(archive) as source:
        member = f"GEO{geo_set}.DAX"
        raw = source.read(member)
    header_size = struct.unpack_from("<H", raw)[0]
    data_offset = header_size + 2
    for offset in range(2, data_offset, 9):
        candidate, block_offset, raw_size, packed_size = struct.unpack_from(
            "<BIHH", raw, offset
        )
        if candidate != block_id:
            continue
        start = data_offset + block_offset
        decoded = decode_rle(raw[start : start + packed_size], raw_size)
        if len(decoded) != GEO_BLOCK_SIZE:
            raise ValueError(f"block={block_id} size={len(decoded)}")
        return decoded[2:]
    raise ValueError(f"{member} block {block_id} not found")


def edge(payload: bytes, x: int, y: int, direction: int) -> tuple[int, int]:
    index = y * WIDTH + x
    wall_index = {0: 0, 2: 1, 4: 2, 6: 3}[direction]
    first = payload[index]
    second = payload[0x100 + index]
    extra = payload[0x300 + index]
    walls = ((first >> 4) & 0x0F, first & 0x0F, (second >> 4) & 0x0F, second & 0x0F)
    details = (extra & 3, (extra >> 2) & 3, (extra >> 4) & 3, (extra >> 6) & 3)
    return walls[wall_index], details[wall_index]


def delta(direction: int) -> tuple[int, int]:
    return {0: (0, -1), 2: (1, 0), 4: (0, 1), 6: (-1, 0)}[direction]


def neighbors(payload: bytes, source: tuple[int, int], open_doors: bool = False):
    x, y = source
    for direction in (0, 2, 4, 6):
        dx, dy = delta(direction)
        target = ((x + dx) % WIDTH, (y + dy) % HEIGHT)
        wall, detail = edge(payload, x, y, direction)
        other_wall, other_detail = edge(payload, *target, (direction + 4) % 8)
        source_open = wall == 0 or detail == 1 or (open_doors and detail in (2, 3))
        target_open = other_wall == 0 or other_detail == 1 or (
            open_doors and other_detail in (2, 3)
        )
        if source_open and target_open:
            yield target, direction, wall, detail, other_wall, other_detail


def component(
    payload: bytes, start: tuple[int, int], open_doors: bool = False
) -> set[tuple[int, int]]:
    visited = {start}
    queue = deque([start])
    while queue:
        source = queue.popleft()
        for target, *_ in neighbors(payload, source, open_doors):
            if target not in visited:
                visited.add(target)
                queue.append(target)
    return visited


def shortest_path(
    payload: bytes,
    start: tuple[int, int],
    target: tuple[int, int],
    avoid_special: bool,
    open_doors: bool = False,
):
    queue = deque([start])
    previous: dict[tuple[int, int], tuple[tuple[int, int], tuple[int, int, int, int, int, int]]] = {}
    visited = {start}
    while queue and target not in visited:
        source = queue.popleft()
        for candidate in neighbors(payload, source, open_doors):
            destination, direction, wall, detail, other_wall, other_detail = candidate
            if avoid_special and destination != target:
                terrain = payload[0x200 + destination[1] * WIDTH + destination[0]]
                if terrain not in (0x00, 0x80, 0xC0):
                    continue
            if destination in visited:
                continue
            visited.add(destination)
            previous[destination] = (source, candidate)
            queue.append(destination)
    if target not in visited:
        return None
    path = []
    cursor = target
    while cursor != start:
        source, candidate = previous[cursor]
        path.append((source, candidate))
        cursor = source
    path.reverse()
    return path


def main() -> int:
    parser = argparse.ArgumentParser()
    parser.add_argument("archive", type=Path)
    parser.add_argument("--set", type=int, default=2, dest="geo_set")
    parser.add_argument("--block", type=int, required=True)
    parser.add_argument("--start", type=int, nargs=2, required=True, metavar=("X", "Y"))
    parser.add_argument("--target", type=int, nargs=2, metavar=("X", "Y"))
    parser.add_argument("--avoid-special", action="store_true")
    parser.add_argument("--open-doors", action="store_true")
    args = parser.parse_args()
    digest = hashlib.sha256(args.archive.read_bytes()).hexdigest()
    payload = read_block(args.archive, args.geo_set, args.block)
    start = tuple(args.start)
    cells = component(payload, start, args.open_doors)
    print(f"input={args.archive}")
    print(f"sha256={digest}")
    print(
        f"set={args.geo_set} block={args.block} start={start} "
        f"reachable_cells={len(cells)}"
    )
    if args.target is not None:
        target = tuple(args.target)
        path = shortest_path(
            payload, start, target, args.avoid_special, args.open_doors
        )
        print(f"target={target} path_length={len(path) if path is not None else 'unreachable'}")
        if path is not None:
            for index, (source, candidate) in enumerate(path, start=1):
                destination, direction, wall, detail, other_wall, other_detail = candidate
                print(
                    f"step={index:02d} ({source[0]},{source[1]})--{direction} "
                    f"wall={wall:02X}/{detail} other={other_wall:02X}/{other_detail} "
                    f"-->({destination[0]},{destination[1]}) "
                    f"terrain={payload[0x200 + destination[1] * WIDTH + destination[0]]:02X}"
                )
    terrain_cells = []
    for y in range(HEIGHT):
        for x in range(WIDTH):
            if (x, y) in cells and payload[0x200 + y * WIDTH + x] != 0:
                terrain_cells.append((x, y, payload[0x200 + y * WIDTH + x]))
    print("terrain_cells=" + ", ".join(f"({x},{y})={terrain:02X}" for x, y, terrain in terrain_cells))
    for y in range(HEIGHT):
        for x in range(WIDTH):
            if (x, y) not in cells:
                continue
            for direction in (0, 2, 4, 6):
                dx, dy = delta(direction)
                if 0 <= x + dx < WIDTH and 0 <= y + dy < HEIGHT:
                    continue
                wall, detail = edge(payload, x, y, direction)
                other_wall, other_detail = edge(payload, (x + dx) % WIDTH, (y + dy) % HEIGHT, (direction + 4) % 8)
                print(
                    f"boundary=({x},{y},{direction}) wall={wall:02X} detail={detail} "
                    f"other_wall={other_wall:02X} other_detail={other_detail}"
                )
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
