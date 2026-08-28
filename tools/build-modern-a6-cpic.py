#!/usr/bin/env python3
"""Edge-aware 2x redraw of every Gold Box tactical CPIC sprite."""

from __future__ import annotations

import colorsys
import hashlib
from pathlib import Path

from PIL import Image, ImageDraw


ROOT = Path(__file__).resolve().parents[1]
SOURCE = ROOT / "assets" / "sprites"
OUTPUT = ROOT / "assets" / "modern-a6" / "sprites"
OVERVIEW = ROOT / "docs" / "reference" / "modern-a6-cpic-overview.png"
CURATED = {"cpic1-block-01-item-00.png"}

EGA_MATERIAL = {
    (0, 0, 0): (10, 13, 18), (0, 0, 173): (28, 46, 94),
    (0, 173, 0): (42, 112, 62), (0, 173, 173): (43, 123, 132),
    (173, 0, 0): (132, 47, 38), (173, 0, 173): (117, 52, 111),
    (173, 82, 0): (139, 83, 42), (173, 173, 173): (157, 153, 142),
    (82, 82, 82): (73, 78, 84), (82, 82, 255): (70, 108, 188),
    (82, 255, 82): (92, 174, 101), (82, 255, 255): (97, 191, 194),
    (255, 82, 82): (210, 91, 69), (255, 82, 255): (190, 101, 174),
    (255, 255, 82): (231, 198, 85), (255, 255, 255): (235, 231, 215),
}


def modern_color(pixel: tuple[int, int, int, int]) -> tuple[int, int, int, int]:
    red, green, blue, alpha = pixel
    if alpha == 0:
        return 0, 0, 0, 0
    if (red, green, blue) in EGA_MATERIAL:
        return (*EGA_MATERIAL[(red, green, blue)], alpha)
    hue, saturation, value = colorsys.rgb_to_hsv(red / 255, green / 255, blue / 255)
    saturation = min(1.0, saturation * 1.08)
    if 0.12 < value < 0.78:
        value = min(1.0, value * 1.06)
    rr, gg, bb = colorsys.hsv_to_rgb(hue, saturation, value)
    return round(rr * 255), round(gg * 255), round(bb * 255), alpha


def scale2x(source: Image.Image) -> Image.Image:
    source = source.convert("RGBA")
    width, height = source.size
    target = Image.new("RGBA", (width * 2, height * 2))
    pixels = source.load()
    out = target.load()

    def at(x: int, y: int):
        return pixels[min(width - 1, max(0, x)), min(height - 1, max(0, y))]

    for y in range(height):
        for x in range(width):
            above, left, center = at(x, y - 1), at(x - 1, y), at(x, y)
            right, below = at(x + 1, y), at(x, y + 1)
            if left != right and above != below:
                values = (
                    left if left == above else center,
                    right if above == right else center,
                    left if left == below else center,
                    right if below == right else center,
                )
            else:
                values = (center, center, center, center)
            out[x * 2, y * 2] = modern_color(values[0])
            out[x * 2 + 1, y * 2] = modern_color(values[1])
            out[x * 2, y * 2 + 1] = modern_color(values[2])
            out[x * 2 + 1, y * 2 + 1] = modern_color(values[3])
    return target


def add_material_depth(image: Image.Image) -> Image.Image:
    image = image.convert("RGBA")
    source = image.copy().load()
    output = image.load()
    width, height = image.size
    for y in range(height):
        for x in range(width):
            red, green, blue, alpha = source[x, y]
            if alpha == 0:
                continue
            delta = ((x * 17 + y * 29) % 7) - 3
            for dx, dy, light in ((0, -1, 13), (-1, 0, 8), (0, 1, -11), (1, 0, -7)):
                nx, ny = x + dx, y + dy
                if nx < 0 or ny < 0 or nx >= width or ny >= height or source[nx, ny][3] == 0:
                    delta += light
            output[x, y] = tuple(max(0, min(255, value + delta)) for value in (red, green, blue)) + (alpha,)
    return image


def make_overview(paths: list[Path]) -> None:
    representatives: dict[str, Path] = {}
    for path in paths:
        digest = hashlib.sha256(path.read_bytes()).hexdigest()
        representatives.setdefault(digest, path)
    items = list(representatives.values())
    columns, cell_w, cell_h = 10, 144, 130
    rows = (len(items) + columns - 1) // columns
    sheet = Image.new("RGB", (columns * cell_w, rows * cell_h), "#10141a")
    draw = ImageDraw.Draw(sheet)
    for index, source in enumerate(items):
        sprite = Image.open(OUTPUT / source.name).convert("RGBA")
        sprite.thumbnail((104, 96), Image.Resampling.NEAREST)
        x = (index % columns) * cell_w + (cell_w - sprite.width) // 2
        y = (index // columns) * cell_h + 2
        sheet.paste(sprite, (x, y), sprite)
        draw.text(
            ((index % columns) * cell_w + 3, (index // columns) * cell_h + 101),
            source.stem.replace("-item-00", ""), fill="white"
        )
    sheet.save(OVERVIEW, "PNG", optimize=True)


def main() -> None:
    OUTPUT.mkdir(parents=True, exist_ok=True)
    paths = sorted(SOURCE.glob("cpic*-item-*.png"))
    if len(paths) != 156:
        raise SystemExit(f"expected 156 source CPIC files, got {len(paths)}")
    for path in paths:
        target = OUTPUT / path.name
        if path.name in CURATED and target.exists():
            continue
        add_material_depth(scale2x(Image.open(path))).save(target, "PNG", optimize=True)
    make_overview(paths)
    print(f"redrew {len(paths)} CPIC files; {len(CURATED)} curated override retained")


if __name__ == "__main__":
    main()
