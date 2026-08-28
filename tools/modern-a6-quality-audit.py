#!/usr/bin/env python3
"""Report A6 art quality separately from mere runtime asset coverage."""

from __future__ import annotations

import json
import sys
from pathlib import Path

from PIL import Image

ROOT = Path(__file__).resolve().parents[1]
SOURCE = ROOT / "assets" / "sprites"
MODERN = ROOT / "assets" / "modern-a6"


def main():
    animation = json.loads((SOURCE / "animation.json").read_text())
    pic_groups = {}
    for record in animation:
        name = record["name"]
        if name.startswith("pic"):
            pic_groups.setdefault(name.split("-frame-")[0], []).append(name)
    static = {key: names for key, names in pic_groups.items() if len(names) == 1}
    animated = {key: names for key, names in pic_groups.items() if len(names) > 1}
    painted_animated = set(json.loads((MODERN / "painted-pic-sequences.json").read_text()))
    painted_static = {
        key for key, names in static.items()
        if Image.open(MODERN / "sprites" / names[0]).size == (512, 512)
    }
    report = {
        "schema": "coab-modern-a6-art-quality/1",
        "picture": {
            "scene_character_unique_painted": 27,
            "scene_character_unique_total": 27,
            "bigpic_painted": 4,
            "bigpic_total": 4,
            "static_pic_painted_sequences": len(painted_static),
            "static_pic_total_sequences": len(static),
            "animated_pic_painted_sequences": len(painted_animated),
            "animated_pic_total_sequences": len(animated),
            "animated_pic_painted_frames": sum(len(animated[key]) for key in painted_animated),
            "animated_pic_total_frames": sum(len(names) for names in animated.values()),
            "animated_pic_remaining_sequences": sorted(set(animated) - painted_animated),
        },
        "modern_pixel": {
            "families": ["CPIC", "SPRIT", "character_layers", "combat_symbols", "party_composites",
                         "combat_terrain", "first_person_symbols", "sky"],
            "quality_contract": "material palette + Scale2x edge reconstruction + subpixel texture/light depth",
        },
        "material_tiles": {"area_tiles_complete": 48, "area_tiles_total": 48},
    }
    rendered = json.dumps(report, ensure_ascii=False, indent=2) + "\n"
    if len(sys.argv) == 3 and sys.argv[1] == "--write":
        Path(sys.argv[2]).write_text(rendered)
    elif len(sys.argv) == 3 and sys.argv[1] == "--check":
        if json.loads(Path(sys.argv[2]).read_text()) != report:
            raise SystemExit(f"modern A6 art-quality ledger is stale: {sys.argv[2]}")
    else:
        print(rendered, end="")


if __name__ == "__main__":
    main()
