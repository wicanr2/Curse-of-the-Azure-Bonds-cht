#!/usr/bin/env python3
"""Produce a deterministic inventory for a Golden Box DOS image ZIP."""

from __future__ import annotations

import argparse
import hashlib
import pathlib
import zipfile


def classify(name: str) -> str:
    upper = name.upper()
    if upper.endswith(".EXE"):
        return "executable"
    if upper.endswith(".BAT"):
        return "dos-batch"
    if upper.endswith(".DAX"):
        if upper.startswith("ECL"):
            return "event-script-candidate"
        if upper.startswith(("PIC", "CPIC", "BIGPIC", "TITLE")):
            return "image-candidate"
        if upper.startswith(("GEO", "WALLDEF", "TILES")):
            return "dungeon-candidate"
        if upper.startswith(("MON", "HEAD", "BODY", "SPRIT", "COMSPR")):
            return "actor-candidate"
        return "dax-unknown"
    if upper == "ITEMS":
        return "table-candidate"
    if upper == "GAME.OVR":
        return "overlay-candidate"
    return "other"


def main() -> int:
    parser = argparse.ArgumentParser()
    parser.add_argument("image", type=pathlib.Path)
    args = parser.parse_args()

    with zipfile.ZipFile(args.image) as archive:
        infos = sorted(archive.infolist(), key=lambda info: info.filename.upper())
        print("# Golden Box image inventory")
        print(f"source: {args.image}")
        print(f"entries: {len(infos)}")
        print("\n| file | bytes | sha256 | class | marker |")
        print("|---|---:|---|---|---|")
        for info in infos:
            data = archive.read(info)
            digest = hashlib.sha256(data).hexdigest()[:16]
            marker = data[:2].hex(" ") if data else ""
            if data[:2] == b"MZ":
                marker += " (MZ)"
            print(f"| `{info.filename}` | {len(data)} | `{digest}` | {classify(info.filename)} | `{marker}` |")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
