#!/usr/bin/env python3
"""Build the README overview for the four modern A6 BIGPIC redraws."""

from pathlib import Path
from PIL import Image, ImageDraw

ROOT = Path(__file__).resolve().parents[1]
paths = sorted((ROOT / "assets" / "modern-a6" / "pictures").glob("bigpic*.png"))
if len(paths) != 4:
    raise SystemExit(f"expected 4 modern BIGPIC files, got {len(paths)}")
sheet = Image.new("RGB", (1280, 570), "#10141a")
draw = ImageDraw.Draw(sheet)
for index, path in enumerate(paths):
    image = Image.open(path).convert("RGB").resize((608, 240), Image.Resampling.LANCZOS)
    x = 20 + (index % 2) * 630
    y = 20 + (index // 2) * 275
    sheet.paste(image, (x, y))
    draw.rectangle((x, y, x + 607, y + 239), outline="#f4c842", width=2)
    draw.text((x, y + 246), path.stem, fill="white")
sheet.save(ROOT / "docs" / "reference" / "modern-a6-bigpic-overview.png", "PNG", optimize=True)
