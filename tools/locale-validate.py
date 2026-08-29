#!/usr/bin/env python3
"""Fail-closed structural validation for all shipped player locales."""

import json
import re
from pathlib import Path

FORMAT = re.compile(r"%(?:\[[0-9]+\])?[+#0 .'\-]*[0-9]*(?:\.[0-9]+)?[bcdoOqxXUeEfFgGsTpVv%]")
LANGUAGES = ("zh-TW", "zh-CN", "ja", "en")


def read(path):
    return json.loads(Path(path).read_text(encoding="utf-8"))


def signature(value):
    return FORMAT.findall(value)


ui = {lang: read(f"assets/locale/{lang}.json")["strings"] for lang in LANGUAGES}
pack = {lang: read(f"gamepack/pack/20-locale.{lang}.json")["locales"][lang] for lang in LANGUAGES}
guides = {lang: read(f"assets/guide/maps.{lang}.json") for lang in LANGUAGES}

for name, catalogs in (("UI", ui), ("game-pack", pack)):
    expected = set(catalogs["zh-TW"])
    for lang, values in catalogs.items():
        if set(values) != expected:
            raise SystemExit(f"{name} {lang}: key set differs from zh-TW")
        for key, value in values.items():
            if not value.strip():
                raise SystemExit(f"{name} {lang}: empty {key}")
            if signature(value) != signature(catalogs["zh-TW"][key]):
                raise SystemExit(f"{name} {lang}: format signature differs at {key}")
            valid_starts = {match.start() for match in FORMAT.finditer(value)}
            if any(character == "%" and index not in valid_starts for index, character in enumerate(value)):
                raise SystemExit(f"{name} {lang}: stray percent at {key}")

expected_maps = set(guides["zh-TW"]["maps"])
for lang, catalog in guides.items():
    if catalog.get("schema") != "coab-guide-maps/1" or set(catalog["maps"]) != expected_maps:
        raise SystemExit(f"guide {lang}: schema/map set differs")
    for key in expected_maps:
        reference = guides["zh-TW"]["maps"][key]["points"]
        translated = catalog["maps"][key]["points"]
        if [(p["x"], p["y"], p["source"]) for p in translated] != [(p["x"], p["y"], p["source"]) for p in reference]:
            raise SystemExit(f"guide {lang}: event identity differs at {key}")

print(f"locale validation passed: UI {len(ui['zh-TW'])} keys; game-pack {len(pack['zh-TW'])} keys; guide {len(expected_maps)} maps")
