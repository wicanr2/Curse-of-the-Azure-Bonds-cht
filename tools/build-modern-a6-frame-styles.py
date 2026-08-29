#!/usr/bin/env python3
"""由核准的 A 框重生 B／C 色材；幾何與透明安全區逐像素不變。"""

from pathlib import Path
from PIL import Image, ImageDraw

ROOT = Path(__file__).resolve().parents[1]
UI = ROOT / "assets" / "modern-a6" / "ui"

PALETTES = {
    "b": {(203, 194, 170): (89, 84, 82), (249, 239, 207): (151, 143, 132), (64, 55, 43): (30, 27, 29)},
    "c": {(203, 194, 170): (221, 204, 164), (249, 239, 207): (255, 239, 196), (64, 55, 43): (112, 87, 50)},
}


def add_ornament(image, style):
    draw = ImageDraw.Draw(image)
    if style == "b":
        ink, glint = (30, 27, 29, 255), (151, 143, 132, 255)
        for x in range(14, 626, 20):
            draw.polygon(((x, 5), (x + 4, 2), (x + 8, 5), (x + 4, 8)), outline=ink)
            draw.point((x + 4, 3), fill=glint)
            draw.polygon(((x, 474), (x + 4, 471), (x + 8, 474), (x + 4, 477)), outline=ink)
    elif style == "c":
        ink, glint = (112, 87, 50, 255), (255, 239, 196, 255)
        for x in range(12, 628, 24):
            draw.line((x, 7, x, 3, x + 8, 3, x + 8, 7, x + 4, 7), fill=ink, width=1)
            draw.line((x, 472, x, 476, x + 8, 476, x + 8, 472, x + 4, 472), fill=glint, width=1)
        for x, y in ((1, 1), (629, 1), (1, 469), (629, 469)):
            draw.rectangle((x, y, x + 9, y + 9), outline=ink, width=1)
            draw.rectangle((x + 3, y + 3, x + 6, y + 6), outline=glint, width=1)


def recolour(source: Path, target: Path, palette, style):
    image = Image.open(source).convert("RGBA")
    pixels = image.load()
    for y in range(image.height):
        for x in range(image.width):
            red, green, blue, alpha = pixels[x, y]
            replacement = palette.get((red, green, blue))
            if replacement is not None:
                pixels[x, y] = (*replacement, alpha)
    add_ornament(image, style)
    image.save(target, "PNG", optimize=True)
    print(f"wrote {target.relative_to(ROOT)}")


def main():
    for stem in ("adventure-frame", "combat-frame"):
        source = UI / f"{stem}.png"
        for style, palette in PALETTES.items():
            recolour(source, UI / f"{stem}-{style}.png", palette, style)


if __name__ == "__main__":
    main()
