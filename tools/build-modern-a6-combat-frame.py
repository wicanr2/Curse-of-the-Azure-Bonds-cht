#!/usr/bin/env python3
"""Render the approved A6 18px carved-stone combat overlay."""

from pathlib import Path
from PIL import Image, ImageDraw

ROOT = Path(__file__).resolve().parents[1]
OUT = ROOT / "assets" / "modern-a6" / "ui" / "combat-frame.png"


def main():
    image = Image.new("RGBA", (640, 480), (0, 0, 0, 0))
    draw = ImageDraw.Draw(image)
    stone, light, shadow = (203, 194, 170, 255), (249, 239, 207, 255), (64, 55, 43, 255)
    gold, glint, dark = (255, 213, 45, 255), (255, 250, 176, 255), (126, 69, 2, 255)
    blue, blue_glint = (31, 137, 220, 255), (170, 232, 255, 255)
    draw.rectangle((0, 0, 639, 479), fill=stone)
    draw.rectangle((18, 18, 621, 461), fill=(0, 0, 0, 0))
    draw.line((2, 2, 637, 2), fill=light, width=2)
    draw.line((2, 477, 637, 477), fill=shadow, width=2)
    # Sparse chisel marks preserve the approved refined-stone rhythm.
    for x in range(8, 632, 18):
        draw.line((x, 7, x + 6, 3), fill=shadow)
        draw.line((x + 2, 8, x + 8, 4), fill=light)
        draw.line((x, 472, x + 6, 476), fill=light)
        draw.line((x + 2, 471, x + 8, 475), fill=shadow)
    for y in range(10, 466, 18):
        draw.line((7, y, 3, y + 6), fill=shadow)
        draw.line((8, y + 2, 4, y + 8), fill=light)
        draw.line((632, y, 636, y + 6), fill=light)
        draw.line((631, y + 2, 635, y + 8), fill=shadow)
    for box in ((14, 14, 625, 465),):
        draw.rectangle(box, outline=dark, width=1)
        draw.rectangle((box[0]+1, box[1]+1, box[2]-1, box[3]-1), outline=gold, width=2)
        draw.rectangle((box[0]+3, box[1]+3, box[2]-3, box[3]-3), outline=glint, width=1)
    # Internal bands match the native combat partition without covering content.
    for orientation, pos, start, end in (("v", 358, 18, 359), ("h", 358, 18, 621), ("h", 446, 18, 621)):
        if orientation == "v":
            draw.rectangle((pos, start, pos + 7, end), fill=stone)
            draw.line((pos + 1, start, pos + 1, end), fill=light, width=1)
            draw.line((pos + 3, start, pos + 3, end), fill=glint, width=1)
            draw.line((pos + 4, start, pos + 4, end), fill=gold, width=2)
            draw.line((pos + 7, start, pos + 7, end), fill=shadow, width=1)
        else:
            draw.rectangle((start, pos, end, pos + 7), fill=stone)
            draw.line((start, pos + 1, end, pos + 1), fill=light, width=1)
            draw.line((start, pos + 3, end, pos + 3), fill=glint, width=1)
            draw.line((start, pos + 4, end, pos + 4), fill=gold, width=2)
            draw.line((start, pos + 7, end, pos + 7), fill=shadow, width=1)
    for x, y in ((14, 14), (621, 14), (14, 461), (621, 461), (357, 357), (357, 445)):
        draw.rectangle((x, y, x + 8, y + 8), fill=dark)
        draw.polygon(((x+4, y+1), (x+7, y+4), (x+4, y+7), (x+1, y+4)), fill=blue)
        draw.point((x+4, y+2), fill=blue_glint)
    OUT.parent.mkdir(parents=True, exist_ok=True)
    image.save(OUT, "PNG", optimize=True)
    print(f"wrote {OUT.relative_to(ROOT)}")


if __name__ == "__main__":
    main()
