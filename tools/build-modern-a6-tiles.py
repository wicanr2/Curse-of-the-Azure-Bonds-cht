#!/usr/bin/env python3
"""Redraw all AREA-map tiles at 4x while preserving their indexed glyphs."""

from __future__ import annotations

import hashlib
from pathlib import Path

from PIL import Image, ImageDraw, ImageFilter


ROOT = Path(__file__).resolve().parents[1]
SOURCE = ROOT / "assets" / "runtime-images" / "tiles"
OUTPUT = ROOT / "assets" / "modern-a6" / "tiles"
OVERVIEW = ROOT / "docs" / "reference" / "modern-a6-tiles-overview.png"
SCALE = 4


def noise(name: str, x: int, y: int) -> int:
    value = hashlib.sha256(f"{name}:{x}:{y}".encode()).digest()[0]
    return value % 13 - 6


def material_color(red: int, green: int, blue: int) -> tuple[int, int, int, int]:
    if green > 150 and red < 80:
        return 38, 128, 48, 255       # moss/grass field
    if red > 210 and green > 210 and blue > 180:
        return 255, 238, 172, 255     # pale carved highlight
    if red > 210 and green > 150 and blue < 100:
        return 246, 190, 50, 255      # gold route/glyph
    if red > 160 and green < 120:
        return 156, 42, 32, 255       # red inset/shadow
    if red < 80 and green < 80 and blue < 80:
        return 35, 24, 20, 255        # deep outline
    if blue > 100 and green > 90 and red < 100:
        return 38, 158, 164, 255      # turquoise edge register
    if abs(red - green) < 35 and abs(green - blue) < 35:
        shade = max(70, min(205, (red + green + blue) // 3))
        return shade, shade, shade, 255
    return red, green, blue, 255


def redraw(path: Path) -> Image.Image:
    source = Image.open(path).convert("RGBA")
    base = Image.new("RGBA", source.size)
    for y in range(source.height):
        for x in range(source.width):
            red, green, blue, alpha = source.getpixel((x, y))
            color = material_color(red, green, blue)
            if color[1] == 128 and color[0] == 38:
                delta = noise(path.name, x, y)
                color = (color[0] + delta, color[1] + delta, color[2] + delta, alpha)
            else:
                color = (*color[:3], alpha)
            base.putpixel((x, y), color)

    # LANCZOS softens the old stair steps; a restrained sharpen restores a
    # readable symbol edge without reverting to nearest-neighbour pixels.
    enlarged = base.resize(
        (source.width * SCALE, source.height * SCALE), Image.Resampling.LANCZOS
    )
    return enlarged.filter(ImageFilter.UnsharpMask(radius=0.8, percent=85, threshold=3))


def main() -> None:
    OUTPUT.mkdir(parents=True, exist_ok=True)
    paths = sorted(SOURCE.glob("tiles-*.png"))
    if len(paths) != 48:
        raise SystemExit(f"expected 48 source tiles, got {len(paths)}")
    for path in paths:
        target = OUTPUT / path.name
        redraw(path).save(target, "PNG", optimize=True)
    sheet = Image.new("RGB", (8 * 150, 6 * 150), "#11151a")
    draw = ImageDraw.Draw(sheet)
    for index, path in enumerate(paths):
        tile = Image.open(OUTPUT / path.name).convert("RGBA").resize(
            (96, 96), Image.Resampling.LANCZOS
        )
        x = (index % 8) * 150 + 27
        y = (index // 8) * 150 + 5
        sheet.paste(tile, (x, y), tile)
        draw.text(
            ((index % 8) * 150 + 5, (index // 8) * 150 + 108),
            path.stem.replace("tiles-", ""), fill="white"
        )
    sheet.save(OVERVIEW, "PNG", optimize=True)
    print(f"redrew {len(paths)} modern A6 tiles")


if __name__ == "__main__":
    main()
