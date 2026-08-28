#!/usr/bin/env python3
"""Build reusable README overview sheets for the completed A6 redraw families."""

from __future__ import annotations

import hashlib
from pathlib import Path

from PIL import Image, ImageDraw


ROOT = Path(__file__).resolve().parents[1]
MODERN = ROOT / "assets" / "modern-a6"
REFERENCE = ROOT / "docs" / "reference"


def unique(paths):
    result, seen = [], set()
    for path in sorted(paths):
        digest = hashlib.sha256(path.read_bytes()).digest()
        if digest not in seen:
            seen.add(digest)
            result.append(path)
    return result


def labelled_sheet(paths, output, columns=12, cell=(132, 112)):
    paths = unique(paths)
    rows = (len(paths) + columns - 1) // columns
    sheet = Image.new("RGB", (columns * cell[0], rows * cell[1]), "#0d121a")
    draw = ImageDraw.Draw(sheet)
    for index, path in enumerate(paths):
        image = Image.open(path).convert("RGBA")
        image.thumbnail((cell[0] - 12, cell[1] - 30), Image.Resampling.NEAREST)
        x = index % columns * cell[0] + (cell[0] - image.width) // 2
        y = index // columns * cell[1] + 2
        sheet.paste(image, (x, y), image)
        draw.text((index % columns * cell[0] + 3, y + image.height + 3), path.stem[:20], fill="white")
    sheet.save(output, "PNG", optimize=True)


def symbol_sheet(paths, output, columns=40):
    paths = unique(paths)
    cell = 20
    rows = (len(paths) + columns - 1) // columns
    sheet = Image.new("RGB", (columns * cell, rows * cell), "#111722")
    for index, path in enumerate(paths):
        image = Image.open(path).convert("RGBA")
        image.thumbnail((16, 16), Image.Resampling.NEAREST)
        x, y = index % columns * cell + 2, index // columns * cell + 2
        sheet.paste(image, (x, y), image)
    sheet.save(output, "PNG", optimize=True)


def main():
    labelled_sheet(list((MODERN / "sprites").glob("pic*-frame-*.png")),
                   REFERENCE / "modern-a6-picture-animation-overview.png")
    labelled_sheet(list((MODERN / "sprites").glob("sprit*-frame-*.png")),
                   REFERENCE / "modern-a6-sprite-animation-overview.png")
    labelled_sheet(list((MODERN / "combat").glob("*.png")),
                   REFERENCE / "modern-a6-combat-terrain-overview.png", columns=10)
    symbol_sheet(list((MODERN / "symbols").glob("*.png")),
                 REFERENCE / "modern-a6-first-person-symbol-overview.png")
    print("built PIC, SPRIT, combat-terrain, and first-person-symbol overview sheets")


if __name__ == "__main__":
    main()
