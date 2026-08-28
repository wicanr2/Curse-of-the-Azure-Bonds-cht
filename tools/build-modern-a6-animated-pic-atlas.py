#!/usr/bin/env python3
"""Build a traceable reference atlas for first frames of multi-frame PICs."""

from __future__ import annotations

import json
from pathlib import Path

from PIL import Image, ImageDraw

ROOT = Path(__file__).resolve().parents[1]
SOURCE = ROOT / "assets" / "sprites"
ATLAS = ROOT / "workplace" / "modern-a6-animated-pic-reference.png"
INDEX = ROOT / "workplace" / "modern-a6-animated-pic-reference.json"
CELL, COLUMNS, ROWS = 256, 5, 6


def main():
    records = json.loads((SOURCE / "animation.json").read_text())
    groups = {}
    for record in records:
        name = record["name"]
        if name.startswith("pic"):
            groups.setdefault(name.split("-frame-")[0], []).append(record)
    chosen = [sorted(group, key=lambda item: item["name"])[0]
              for _, group in sorted(groups.items()) if len(group) > 1]
    if len(chosen) != 28:
        raise SystemExit(f"expected 28 multi-frame PIC sequences, got {len(chosen)}")
    atlas = Image.new("RGB", (COLUMNS * CELL, ROWS * CELL), "#121722")
    draw = ImageDraw.Draw(atlas)
    index = []
    for number, record in enumerate(chosen):
        path = SOURCE / record["name"]
        image = Image.open(path).convert("RGB").resize((220, 220), Image.Resampling.NEAREST)
        x, y = number % COLUMNS * CELL, number // COLUMNS * CELL
        atlas.paste(image, (x + 18, y + 18))
        draw.rectangle((x, y, x + CELL - 1, y + CELL - 1), outline="#d9bd54", width=3)
        draw.text((x + 8, y + 6), f"{number+1:02d}", fill="white")
        index.append({"cell": number + 1, "key": record["name"].split("-frame-")[0], "first": record["name"]})
    ATLAS.parent.mkdir(parents=True, exist_ok=True)
    atlas.save(ATLAS, "PNG", optimize=True)
    INDEX.write_text(json.dumps(index, indent=2) + "\n")
    print(f"wrote animated PIC atlas with {len(index)} sequence keyframes")


if __name__ == "__main__":
    main()
