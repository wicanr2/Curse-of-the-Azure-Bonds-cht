#!/usr/bin/env python3
"""Build a square, row-major reference sheet for one multi-frame PIC sequence."""

from __future__ import annotations

import json
import math
import sys
from pathlib import Path

from PIL import Image

ROOT = Path(__file__).resolve().parents[1]
SOURCE = ROOT / "assets" / "sprites"


def main():
    if len(sys.argv) != 2:
        raise SystemExit("usage: build-modern-a6-pic-sequence-reference.py picN-block-XX")
    key = sys.argv[1]
    paths = sorted(SOURCE.glob(key + "-frame-*.png"))
    if len(paths) < 2:
        raise SystemExit(f"{key} is not a multi-frame PIC sequence")
    columns = math.ceil(math.sqrt(len(paths)))
    rows = math.ceil(len(paths) / columns)
    cell = 256
    sheet = Image.new("RGB", (columns * cell, rows * cell), "#080b10")
    for index, path in enumerate(paths):
        image = Image.open(path).convert("RGB").resize((cell - 8, cell - 8), Image.Resampling.NEAREST)
        x, y = index % columns * cell + 4, index // columns * cell + 4
        sheet.paste(image, (x, y))
    output = ROOT / "workplace" / f"modern-a6-{key}-reference.png"
    metadata = output.with_suffix(".json")
    sheet.save(output, "PNG", optimize=True)
    metadata.write_text(json.dumps({"key": key, "columns": columns, "rows": rows,
                                    "files": [path.name for path in paths]}, indent=2) + "\n")
    print(f"wrote {output.relative_to(ROOT)} ({columns}x{rows}, {len(paths)} frames)")


if __name__ == "__main__":
    main()
