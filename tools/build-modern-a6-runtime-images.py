#!/usr/bin/env python3
"""Build the complete A6 combat, first-person symbol, and sky image layer."""

from __future__ import annotations

import colorsys
from pathlib import Path

from PIL import Image


ROOT = Path(__file__).resolve().parents[1]
SOURCE = ROOT / "assets" / "runtime-images"
OUTPUT = ROOT / "assets" / "modern-a6"
FAMILIES = ("combat", "symbols", "sky")

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


def tint(pixel):
    red, green, blue, alpha = pixel
    if alpha == 0:
        return 0, 0, 0, 0
    if (red, green, blue) in EGA_MATERIAL:
        return (*EGA_MATERIAL[(red, green, blue)], alpha)
    hue, saturation, value = colorsys.rgb_to_hsv(red / 255, green / 255, blue / 255)
    saturation = min(1.0, saturation * 1.06)
    value = min(1.0, value * (1.08 if 0.1 < value < 0.8 else 1.0))
    rr, gg, bb = colorsys.hsv_to_rgb(hue, saturation, value)
    return round(rr * 255), round(gg * 255), round(bb * 255), alpha


def scale2x(source: Image.Image) -> Image.Image:
    source = source.convert("RGBA")
    width, height = source.size
    result = Image.new("RGBA", (width * 2, height * 2))
    src, dst = source.load(), result.load()
    def at(x, y):
        return src[min(width - 1, max(0, x)), min(height - 1, max(0, y))]
    for y in range(height):
        for x in range(width):
            n, w, p, e, s = at(x, y - 1), at(x - 1, y), at(x, y), at(x + 1, y), at(x, y + 1)
            values = ((w if w == n else p, e if n == e else p, w if w == s else p, e if s == e else p)
                      if w != e and n != s else (p, p, p, p))
            dst[2*x, 2*y], dst[2*x+1, 2*y] = tint(values[0]), tint(values[1])
            dst[2*x, 2*y+1], dst[2*x+1, 2*y+1] = tint(values[2]), tint(values[3])
    return result


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


def mask_wall_transparency(source: Image.Image) -> Image.Image:
    source = source.convert("RGBA")
    # runtimeImageCatalog.maskedSymbolImages uses the fixed EGA palette index
    # 13, not the top-left pixel, as the DOS masked-blit colour.
    key = (255, 82, 255)
    pixels = source.load()
    for y in range(source.height):
        for x in range(source.width):
            red, green, blue, alpha = pixels[x, y]
            if (red, green, blue) == key:
                pixels[x, y] = (red, green, blue, 0)
    return source


def main() -> None:
    counts = {}
    for family in FAMILIES:
        paths = sorted((SOURCE / family).glob("*.png"))
        target = OUTPUT / family
        target.mkdir(parents=True, exist_ok=True)
        for path in paths:
            source = Image.open(path)
            if family == "symbols":
                source = mask_wall_transparency(source)
            add_material_depth(scale2x(source)).save(target / path.name, "PNG", optimize=True)
        counts[family] = len(paths)
    if counts != {"combat": 65, "symbols": 1625, "sky": 3}:
        raise SystemExit(f"unexpected runtime-image counts: {counts}")
    print(f"redrew {sum(counts.values())} runtime images: {counts}")


if __name__ == "__main__":
    main()
