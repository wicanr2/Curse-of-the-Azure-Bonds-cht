#!/usr/bin/env python3
"""重建 640×480 A6 冒險框；所有接點都以重疊矩形閉合，禁止透明縫隙。"""

from pathlib import Path
from PIL import Image, ImageDraw

ROOT = Path(__file__).resolve().parents[1]
OUT = ROOT / "assets" / "modern-a6" / "ui" / "adventure-frame.png"


def main():
    image = Image.new("RGBA", (640, 480), (0, 0, 0, 0))
    draw = ImageDraw.Draw(image)
    stone, light, shadow = (203, 194, 170, 255), (249, 239, 207, 255), (64, 55, 43, 255)
    gold, glint, dark = (255, 213, 45, 255), (255, 250, 176, 255), (126, 69, 2, 255)

    # 10px 外框先畫成完整實心環；四角與各分隔帶都至少重疊 1px。
    draw.rectangle((0, 0, 639, 479), fill=stone)
    draw.rectangle((10, 10, 629, 469), fill=(0, 0, 0, 0))
    draw.line((2, 2, 637, 2), fill=light, width=2)
    draw.line((2, 477, 637, 477), fill=shadow, width=2)
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

    def band(box, vertical=False):
        draw.rectangle(box, fill=stone)
        if vertical:
            draw.line((box[0] + 1, box[1], box[0] + 1, box[3]), fill=light)
            draw.line((box[2], box[1], box[2], box[3]), fill=shadow)
        else:
            draw.line((box[0], box[1] + 1, box[2], box[1] + 1), fill=light)
            draw.line((box[0], box[3], box[2], box[3]), fill=shadow)

    band((264, 9, 271, 263), vertical=True)
    band((9, 256, 630, 263))
    band((9, 454, 630, 470))

    # 左上 scene／portrait 的細金色內框；完整封閉，不與石框留下單像素缺口。
    draw.rectangle((46, 34, 239, 234), fill=dark)
    draw.rectangle((48, 36, 237, 232), outline=gold, width=2)
    draw.rectangle((50, 38, 235, 230), outline=glint, width=1)
    draw.rectangle((51, 39, 234, 229), fill=(0, 0, 0, 0))
    for x, y in ((46, 34), (231, 34), (46, 226), (231, 226), (264, 256), (614, 454)):
        draw.rectangle((x, y, x + 8, y + 8), fill=dark)
        draw.polygon(((x + 4, y + 1), (x + 7, y + 4), (x + 4, y + 7), (x + 1, y + 4)), fill=(31, 137, 220, 255))

    OUT.parent.mkdir(parents=True, exist_ok=True)
    image.save(OUT, "PNG", optimize=True)
    print(f"wrote {OUT.relative_to(ROOT)}")


if __name__ == "__main__":
    main()
