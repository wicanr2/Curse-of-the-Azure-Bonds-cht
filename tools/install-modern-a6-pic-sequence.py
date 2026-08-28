#!/usr/bin/env python3
"""Install one painted PIC sheet and all byte-identical sequence aliases."""

from __future__ import annotations

import hashlib
import json
import sys
from pathlib import Path

from PIL import Image

ROOT = Path(__file__).resolve().parents[1]
SOURCE = ROOT / "assets" / "sprites"
OUTPUT = ROOT / "assets" / "modern-a6" / "sprites"
MANIFEST = ROOT / "assets" / "modern-a6" / "painted-pic-sequences.json"


def digest(path: Path) -> str:
    return hashlib.sha256(path.read_bytes()).hexdigest()


def main():
    if len(sys.argv) != 3:
        raise SystemExit("usage: install-modern-a6-pic-sequence.py picN-block-XX painted-sheet.png")
    key, sheet_path = sys.argv[1], Path(sys.argv[2])
    if not sheet_path.is_absolute():
        sheet_path = ROOT / sheet_path
    metadata = json.loads((ROOT / "workplace" / f"modern-a6-{key}-reference.json").read_text())
    target_paths = [SOURCE / name for name in metadata["files"]]
    target_signature = [digest(path) for path in target_paths]
    animation = json.loads((SOURCE / "animation.json").read_text())
    groups = {}
    for record in animation:
        name = record["name"]
        if name.startswith("pic"):
            groups.setdefault(name.split("-frame-")[0], []).append(SOURCE / name)
    aliases = []
    for candidate, paths in groups.items():
        paths = sorted(paths)
        if [digest(path) for path in paths] == target_signature:
            aliases.append((candidate, paths))
    sheet = Image.open(sheet_path).convert("RGB")
    columns, rows = metadata["columns"], metadata["rows"]
    cell_width, cell_height = sheet.width / columns, sheet.height / rows
    frames = []
    for index in range(len(target_paths)):
        column, row = index % columns, index // columns
        margin_x, margin_y = max(2, round(cell_width * 0.008)), max(2, round(cell_height * 0.008))
        box = (round(column * cell_width) + margin_x, round(row * cell_height) + margin_y,
               round((column + 1) * cell_width) - margin_x, round((row + 1) * cell_height) - margin_y)
        # 512 px already exceeds the largest 1280×960 runtime picture viewport.
        # Keeping every animation cell at 1024 px made the eager Ebiten loader
        # allocate hundreds of MiB twice (CPU + GPU) and could kill normal play.
        frames.append(sheet.crop(box).resize((512, 512), Image.Resampling.LANCZOS))
    installed_keys = set(json.loads(MANIFEST.read_text()))
    for alias, paths in aliases:
        for frame, path in zip(frames, paths):
            frame.save(OUTPUT / path.name, "PNG", compress_level=6)
        installed_keys.add(alias)
    MANIFEST.write_text(json.dumps(sorted(installed_keys), indent=2) + "\n")
    print(f"installed {len(frames)} painted frames for {len(aliases)} identical sequences: {', '.join(key for key, _ in aliases)}")


if __name__ == "__main__":
    main()
