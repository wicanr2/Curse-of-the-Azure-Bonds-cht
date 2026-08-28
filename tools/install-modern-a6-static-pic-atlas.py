#!/usr/bin/env python3
"""Install the painted static-PIC atlas using the traceable cell index."""

from __future__ import annotations

import hashlib
import json
from pathlib import Path

from PIL import Image

ROOT = Path(__file__).resolve().parents[1]
ATLAS = ROOT / "workplace" / "modern-a6-static-pic-painted.png"
INDEX = ROOT / "workplace" / "modern-a6-static-pic-reference.json"
SOURCE = ROOT / "assets" / "sprites"
OUTPUT = ROOT / "assets" / "modern-a6" / "sprites"
COLUMNS, ROWS = 5, 6


def digest(path: Path) -> str:
    return hashlib.sha256(path.read_bytes()).hexdigest()


def main():
    atlas = Image.open(ATLAS).convert("RGB")
    records = json.loads(INDEX.read_text())
    cell_width, cell_height = atlas.width / COLUMNS, atlas.height / ROWS
    representative: dict[str, Image.Image] = {}
    for record in records:
        number = record["cell"] - 1
        column, row = number % COLUMNS, number // COLUMNS
        margin_x = max(3, round(cell_width * 0.012))
        margin_y = max(3, round(cell_height * 0.014))
        box = (
            round(column * cell_width) + margin_x,
            round(row * cell_height) + margin_y,
            round((column + 1) * cell_width) - margin_x,
            round((row + 1) * cell_height) - margin_y,
        )
        source = SOURCE / record["file"]
        source_digest = digest(source)
        if source_digest not in representative:
            # 512 px is still above the maximum in-game picture viewport at
            # 1280×960, while avoiding 1024 px eager-load memory amplification.
            representative[source_digest] = atlas.crop(box).resize((512, 512), Image.Resampling.LANCZOS)
        representative[source_digest].save(OUTPUT / record["file"], "PNG", compress_level=6)
    print(f"installed {len(records)} static PIC entries from {len(representative)} unique painted cells")


if __name__ == "__main__":
    main()
