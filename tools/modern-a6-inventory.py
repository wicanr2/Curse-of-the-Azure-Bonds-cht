#!/usr/bin/env python3
"""Build the source-of-truth denominator for the modern A6 redraw."""

from __future__ import annotations

import hashlib
import json
import sys
from pathlib import Path


ROOT = Path(__file__).resolve().parents[1]


def digest(path: Path) -> str:
    value = hashlib.sha256()
    value.update(path.read_bytes())
    return value.hexdigest()


def collect(directory: Path, pattern: str) -> list[Path]:
    return sorted(directory.glob(pattern))


sprites = ROOT / "assets" / "sprites"
runtime = ROOT / "assets" / "runtime-images"
modern = ROOT / "assets" / "modern-a6"

families = {
    "scene_characters": collect(sprites, "character-area-*.png"),
    "combat_cpic": collect(sprites, "cpic*-item-*.png"),
    "picture_animation": collect(sprites, "pic*-frame-*.png"),
    "sprite_animation": collect(sprites, "sprit*-frame-*.png"),
    "big_picture": collect(sprites, "bigpic*-item-*.png"),
    "character_heads": collect(sprites, "chead-*.png"),
    "character_bodies": collect(sprites, "cbody-*.png"),
    "scene_heads": collect(sprites, "head*-item-*.png"),
    "scene_bodies": collect(sprites, "body*-item-*.png"),
    "combat_symbols": collect(sprites, "comspr-*.png"),
    "party_composites": collect(sprites, "party*.png"),
    "tiles": collect(runtime / "tiles", "*.png"),
    "combat_terrain": collect(runtime / "combat", "*.png"),
    "first_person_symbols": collect(runtime / "symbols", "*.png"),
    "sky": collect(runtime / "sky", "*.png"),
}

modern_patterns = {
    "scene_characters": collect(modern / "pictures", "character-area-*.png"),
    "big_picture": collect(modern / "pictures", "bigpic*.png"),
    "combat_cpic": collect(modern / "sprites", "cpic*.png"),
    "picture_animation": collect(modern / "sprites", "pic*-frame-*.png"),
    "sprite_animation": collect(modern / "sprites", "sprit*-frame-*.png"),
    "character_heads": collect(modern / "sprites", "chead-*.png"),
    "character_bodies": collect(modern / "sprites", "cbody-*.png"),
    "scene_heads": collect(modern / "sprites", "head*-item-*.png"),
    "scene_bodies": collect(modern / "sprites", "body*-item-*.png"),
    "combat_symbols": collect(modern / "sprites", "comspr-*.png"),
    "party_composites": collect(modern / "sprites", "party*.png"),
    "tiles": collect(modern / "tiles", "*.png"),
    "combat_terrain": collect(modern / "combat", "*.png"),
    "first_person_symbols": collect(modern / "symbols", "*.png"),
    "sky": collect(modern / "sky", "*.png"),
}

report = {"schema": "coab-modern-a6-inventory/1", "families": {}}
for name, paths in families.items():
    hashes = {digest(path) for path in paths}
    modern_paths = modern_patterns.get(name, [])
    report["families"][name] = {
        "source_files": len(paths),
        "unique_source_visuals": len(hashes),
        "modern_files": len(modern_paths),
        "modern_unique_visuals": len({digest(path) for path in modern_paths}),
        "source_glob": str(paths[0].parent.relative_to(ROOT)) if paths else "",
    }

rendered = json.dumps(report, ensure_ascii=False, indent=2, sort_keys=True) + "\n"
if len(sys.argv) == 3 and sys.argv[1] == "--check":
    expected = json.loads(Path(sys.argv[2]).read_text(encoding="utf-8"))
    if expected != report:
        raise SystemExit(f"modern A6 inventory is stale: {sys.argv[2]}")
elif len(sys.argv) == 3 and sys.argv[1] == "--write":
    Path(sys.argv[2]).write_text(rendered, encoding="utf-8")
else:
    print(rendered, end="")
