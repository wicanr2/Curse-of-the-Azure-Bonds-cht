#!/usr/bin/env python3
"""Build a traceable 5x6 reference atlas for the 28 one-frame PIC sequences."""

from __future__ import annotations

import json
from pathlib import Path

from PIL import Image, ImageDraw

ROOT = Path(__file__).resolve().parents[1]
SOURCE = ROOT / "assets" / "sprites"
ATLAS = ROOT / "workplace" / "modern-a6-static-pic-reference.png"
INDEX = ROOT / "workplace" / "modern-a6-static-pic-reference.json"
CELL, COLUMNS, ROWS = 256, 5, 6


def main():
    records = json.loads((SOURCE / "animation.json").read_text())
    counts = {}
    for record in records:
        name = record["name"]
        if name.startswith("pic"):
            key = name.split("-frame-")[0]
            counts[key] = counts.get(key, 0) + 1
    paths = sorted(SOURCE / record["name"] for record in records
                   if record["name"].startswith("pic") and counts[record["name"].split("-frame-")[0]] == 1)
    if len(paths) != 28:
        raise SystemExit(f"expected 28 single-frame PIC sequences, got {len(paths)}")
    atlas = Image.new("RGB", (COLUMNS * CELL, ROWS * CELL), "#121722")
    draw = ImageDraw.Draw(atlas)
    index = []
    for number, path in enumerate(paths):
        image = Image.open(path).convert("RGB")
        image = image.resize((220, 220), Image.Resampling.NEAREST)
        x, y = number % COLUMNS * CELL, number // COLUMNS * CELL
        atlas.paste(image, (x + 18, y + 18))
        draw.rectangle((x, y, x + CELL - 1, y + CELL - 1), outline="#d9bd54", width=3)
        draw.text((x + 8, y + 6), f"{number+1:02d}", fill="white")
        index.append({"cell": number + 1, "file": path.name})
    ATLAS.parent.mkdir(parents=True, exist_ok=True)
    atlas.save(ATLAS, "PNG", optimize=True)
    INDEX.write_text(json.dumps(index, indent=2) + "\n")
    print(f"wrote {ATLAS.relative_to(ROOT)} and {INDEX.relative_to(ROOT)}")


if __name__ == "__main__":
    main()
