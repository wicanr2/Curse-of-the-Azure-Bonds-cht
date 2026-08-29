#!/usr/bin/env python3
"""Deterministically generate the four shipped CoAB locale catalogs.

Project strings are processed locally. Network access is neither used nor
required at runtime; the pinned container image already contains both models.
"""

from __future__ import annotations

import json
import re
import sys
from pathlib import Path

import ctranslate2
import sentencepiece as spm
from opencc import OpenCC

ROOT = Path(".")
INCREMENTAL = "--incremental" in sys.argv
UI_ONLY = "--ui-only" in sys.argv
FORMAT = re.compile(r"%(?:\[[0-9]+\])?[+#0 .'\-]*[0-9]*(?:\.[0-9]+)?[bcdoOqxXUeEfFgGsTpVv%]")

TERMS = {
    "Curse of the Azure Bonds": {"zh-CN": "青色枷的诅咒", "ja": "アズール・ボンズの呪い"},
    "Tyranthraxus": {"zh-CN": "提朗瑟克斯", "ja": "ティランスラクサス"},
    "Dracandros": {"zh-CN": "德拉坎德罗斯", "ja": "ドラカンドロス"},
    "Dragonbait": {"zh-CN": "龙饵", "ja": "ドラゴンベイト"},
    "Alias": {"zh-CN": "爱丽雅丝", "ja": "アリアス"},
    "Bane": {"zh-CN": "班恩", "ja": "ベイン"},
    "Moander": {"zh-CN": "摩安德", "ja": "モアンダー"},
    "Lathander": {"zh-CN": "洛山达", "ja": "ラサンダー"},
    "Fire Knives": {"zh-CN": "火刀", "ja": "ファイア・ナイブズ"},
    "Zhentarim": {"zh-CN": "散塔林会", "ja": "ゼンタリム"},
    "Zhentrim": {"zh-CN": "散塔林会", "ja": "ゼンタリム"},
    "Black Network": {"zh-CN": "黑网", "ja": "ブラック・ネットワーク"},
    "Zhentil Keep": {"zh-CN": "散提尔堡", "ja": "ゼンティル・キープ"},
    "Tilverton": {"zh-CN": "提尔佛顿", "ja": "ティルヴァトン"},
    "Myth Drannor": {"zh-CN": "迷斯卓诺", "ja": "ミス・ドラノール"},
    "Helm of Dragons": {"zh-CN": "龙盔", "ja": "ドラゴン・ヘルム"},
    "Gauntlet of Moander": {"zh-CN": "摩安德护手", "ja": "モアンダーのガントレット"},
    "Amulet of Lathander": {"zh-CN": "洛山达护符", "ja": "ラサンダーのアミュレット"},
    "Pool of Radiance": {"zh-CN": "光芒之池", "ja": "プール・オブ・レイディアンス"},
    "Guide Map": {"zh-CN": "攻略地图", "ja": "攻略マップ"},
    "Full Guide": {"zh-CN": "完整攻略", "ja": "完全攻略"},
}

ZH_TO_EN_TERMS = {
    "目前地圖攻略": "Current Guide Map",
    "攻略地圖": "Guide Map",
    "完整攻略": "Full Guide",
    "攻略點": "guide point",
    "探索資訊": "Exploration Information",
    "朝向": "Facing",
}

ZH_TERMS = {
    "提朗瑟克斯": "提朗瑟克斯", "德拉坎德羅斯": "德拉坎德罗斯",
    "愛麗雅絲": "爱丽雅丝", "龍餌": "龙饵", "班恩": "班恩", "摩安德": "摩安德",
    "洛山達": "洛山达", "火刀": "火刀", "散塔林會": "散塔林会", "黑網": "黑网",
    "散提爾堡": "散提尔堡", "提爾佛頓": "提尔佛顿", "迷斯卓諾": "迷斯卓诺",
    "龍盔": "龙盔", "摩安德護手": "摩安德护手", "洛山達護符": "洛山达护符",
}

cc = OpenCC("t2s")
engines = {}


def protect(text: str, terms: list[tuple[str, str]]) -> tuple[str, dict[str, str]]:
    replacements: dict[str, str] = {}
    serial = 0
    def token(value: str) -> str:
        nonlocal serial
        marker = f"<COABTOKEN{serial:04d}/>"
        serial += 1
        replacements[marker] = value
        return marker
    text = FORMAT.sub(lambda match: token(match.group(0)), text)
    for source, target in sorted(terms, key=lambda item: len(item[0]), reverse=True):
        text = re.sub(re.escape(source), lambda _: token(target), text, flags=re.IGNORECASE)
    return text, replacements


def restore(text: str, replacements: dict[str, str]) -> str:
    for marker, value in replacements.items():
        serial = int(re.search(r"([0-9]+)", marker).group(1))
        variant = re.compile(rf"<COABTOKEN0*{serial}/\s*>")
        if not variant.search(text):
            raise RuntimeError(f"AI translation damaged protected token {marker}: {text!r}")
        text = variant.sub(lambda _: value, text, count=1)
    return text


def engine(source: str, target: str):
    key = (source, target)
    if key not in engines:
        folder = {
            ("zt", "en"): "/opt/argos-packages/translate-zt_en-1_9",
            ("en", "ja"): "/opt/argos-packages/en_ja",
        }[key]
        processor = spm.SentencePieceProcessor(model_file=folder + "/sentencepiece.model")
        runtime = ctranslate2.Translator(folder + "/model", device="cpu", inter_threads=2, intra_threads=8)
        engines[key] = (processor, runtime)
    return engines[key]


def raw_batch(values: list[str], source: str, target: str) -> list[str]:
    if not values:
        return []
    processor, runtime = engine(source, target)
    tokens = [processor.encode(value, out_type=str) for value in values]
    results = runtime.translate_batch(tokens, beam_size=1, max_batch_size=256)
    return [processor.decode(item.hypotheses[0]).replace("▁", " ").strip() for item in results]


def segmented_fallback(value: str, source: str, target: str, term_targets: list[tuple[str, str]]) -> str:
    patterns = [FORMAT.pattern] + [re.escape(name) for name, _ in sorted(term_targets, key=lambda item: len(item[0]), reverse=True)]
    protected_re = re.compile("(" + "|".join(patterns) + ")", re.IGNORECASE)
    target_for = {name.lower(): translated for name, translated in term_targets}
    layout = []
    ordinary = []
    position = 0
    for match in protected_re.finditer(value):
        if match.start() > position:
            layout.append(("ai", len(ordinary)))
            ordinary.append(value[position:match.start()])
        token = match.group(0)
        layout.append(("literal", target_for.get(token.lower(), token)))
        position = match.end()
    if position < len(value):
        layout.append(("ai", len(ordinary)))
        ordinary.append(value[position:])
    translated = raw_batch(ordinary, source, target)
    return "".join(translated[index] if kind == "ai" else index for kind, index in layout)


def ai_batch(values: list[str], source: str, target: str) -> list[str]:
    # XML-like markers were verified against both bundled models. Translating
    # the complete sentence preserves grammar around placeholders while exact
    # format directives and glossary terms never enter the model.
    if source == "zt" and target == "en":
        term_targets = list(ZH_TO_EN_TERMS.items())
    else:
        term_targets = [(name, translations[target]) for name, translations in TERMS.items() if target in translations]
    protected_values = []
    replacement_sets = []
    for value in values:
        protected, replacements = protect(value, term_targets)
        protected_values.append(protected)
        replacement_sets.append(replacements)
    translated = raw_batch(protected_values, source, target)
    results = []
    for source_value, translated_value, replacements in zip(values, translated, replacement_sets):
        try:
            results.append(restore(translated_value, replacements))
        except RuntimeError:
            results.append(segmented_fallback(source_value, source, target, term_targets))
    sanitized = []
    for source_value, result in zip(values, results):
        allowed = {}
        for directive in FORMAT.findall(source_value):
            allowed[directive] = allowed.get(directive, 0) + 1
        seen = {}
        def keep_or_drop(match):
            directive = match.group(0)
            seen[directive] = seen.get(directive, 0) + 1
            return directive if seen[directive] <= allowed.get(directive, 0) else ""
        sanitized.append(FORMAT.sub(keep_or_drop, result))
    return sanitized


def ai(text: str, source: str, target: str) -> str:
    return ai_batch([text], source, target)[0]


def simplified(text: str) -> str:
    result = cc.convert(text)
    for source, target in sorted(ZH_TERMS.items(), key=lambda item: len(item[0]), reverse=True):
        result = result.replace(cc.convert(source), target)
    return result


def read(path: str):
    return json.loads((ROOT / path).read_text(encoding="utf-8"))


def write(path: str, value) -> None:
    target = ROOT / path
    target.write_text(json.dumps(value, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")


def translated_map(values: dict[str, str], source: str, target: str) -> dict[str, str]:
    if target == "zh-CN":
        return {key: simplified(value) for key, value in values.items()}
    result: dict[str, str] = {}
    entries = list(values.items())
    for start in range(0, len(entries), 256):
        chunk = entries[start:start + 256]
        translations = ai_batch([value for _, value in chunk], source, target)
        result.update((key, value) for (key, _), value in zip(chunk, translations))
        print(f"{target}: {min(start + 256, len(entries))}/{len(entries)}", flush=True)
    return result


def incremental_map(path: str, nested: tuple[str, ...], values: dict[str, str], source: str, target: str) -> dict[str, str]:
    existing: dict[str, str] = {}
    if INCREMENTAL and (ROOT / path).exists():
        loaded = read(path)
        for key in nested:
            loaded = loaded[key]
        existing = loaded
    missing = {key: value for key, value in values.items() if key not in existing}
    existing.update(translated_map(missing, source, target))
    return {key: existing[key] for key in values}


def sanitize_formats(source: dict[str, str], translated: dict[str, str]) -> dict[str, str]:
    for key, value in translated.items():
        allowed = {}
        for directive in FORMAT.findall(source[key]):
            allowed[directive] = allowed.get(directive, 0) + 1
        seen = {}
        def keep_or_drop(match):
            directive = match.group(0)
            seen[directive] = seen.get(directive, 0) + 1
            return directive if seen[directive] <= allowed.get(directive, 0) else ""
        value = FORMAT.sub(keep_or_drop, value)
        valid_starts = {match.start() for match in FORMAT.finditer(value)}
        translated[key] = "".join(character for index, character in enumerate(value) if character != "%" or index in valid_starts)
    return translated


def translate_tree(value, source: str, target: str, field: str = ""):
    if isinstance(value, dict):
        return {key: translate_tree(item, source, target, key) for key, item in value.items()}
    if isinstance(value, list):
        return [translate_tree(item, source, target, field) for item in value]
    if isinstance(value, str) and field in {"title", "label", "summary"}:
        return simplified(value) if target == "zh-CN" else ai(value, source, target)
    return value


ui_zt = read("assets/locale/zh-TW.json")["strings"]
pack_en = read("gamepack/pack/20-locale.en.json")["locales"]["en"]
ui_en = incremental_map("assets/locale/en.json", ("strings",), ui_zt, "zt", "en")
write("assets/locale/en.json", {"language": "en", "strings": sanitize_formats(ui_zt, ui_en)})
write("assets/locale/zh-CN.json", {"language": "zh-CN", "strings": sanitize_formats(ui_zt, incremental_map("assets/locale/zh-CN.json", ("strings",), ui_zt, "zt", "zh-CN"))})
write("assets/locale/ja.json", {"language": "ja", "strings": sanitize_formats(ui_zt, incremental_map("assets/locale/ja.json", ("strings",), ui_en, "en", "ja"))})

pack_zt = read("gamepack/pack/20-locale.zh-TW.json")["locales"]["zh-TW"]
if not UI_ONLY:
    write("gamepack/pack/20-locale.zh-CN.json", {"locales": {"zh-CN": sanitize_formats(pack_zt, incremental_map("gamepack/pack/20-locale.zh-CN.json", ("locales", "zh-CN"), pack_zt, "zt", "zh-CN"))}})
    write("gamepack/pack/20-locale.ja.json", {"locales": {"ja": sanitize_formats(pack_en, incremental_map("gamepack/pack/20-locale.ja.json", ("locales", "ja"), pack_en, "en", "ja"))}})

guide = read("assets/guide/maps.zh-TW.json")
if not UI_ONLY and (not INCREMENTAL or not (ROOT / "assets/guide/maps.en.json").exists()):
    guide_en = translate_tree(guide, "zt", "en")
    write("assets/guide/maps.en.json", guide_en)
    write("assets/guide/maps.zh-CN.json", translate_tree(guide, "zt", "zh-CN"))
    write("assets/guide/maps.ja.json", translate_tree(guide_en, "en", "ja"))

print("done", flush=True)
