"""把函式剖面整理成分流報告，決定「先讀哪一批」。

輸入是 `tools/ida/export_profiles.py` 的產物加上 overlay manifest 與 Borland
符號。它會把 far call 解回「哪個 overlay 的哪個 entry、原始函式名是什麼」，
再依客觀特徵分組。

**這份報告只排序，不改台帳狀態。** 分組是候選判斷：thunk 可能不只是 thunk，
「只呼叫 RTL」也可能是關鍵的格式化邏輯。要標狀態一律回
`docs/audit/re-function-ledger.json`，並附證據。

用法：
    python3 scripts/profile_triage.py > docs/audit/function-triage.md
"""

import collections
import glob
import json
import os

ROOT = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
SWEEP = os.path.join(ROOT, "workplace", "re-sweep")
HEADER_PARAGRAPHS = {"dos": None, "pc98": None}   # 由 manifest 推得


def load_json(path):
    with open(path, encoding="utf-8") as handle:
        return json.load(handle)


def overlay_index(platform):
    """control segment → (module, entries)，以及 module → entries。"""
    manifest = load_json(os.path.join(SWEEP, platform, "ovr-manifest.json"))
    executable = load_json(os.path.join(SWEEP, platform, "ovr-manifest.json"))["executable"]
    # MZ header paragraph 數：由第一段 control 的 file offset 與段基址反推不可靠，
    # 改用兩平台已知值——DOS 07B0h／PC-98 09A0h（spec 519／412 記錄過）。
    header = {"dos": 0x7B0, "pc98": 0x9A0}[platform]
    by_segment = {}
    for overlay in manifest["overlays"]:
        offset = overlay["executable_offset"] - header
        if offset >= 0 and offset % 16 == 0:
            by_segment[offset // 16] = overlay
    return by_segment, executable


def symbol_names():
    path = os.path.join(SWEEP, "pc98", "borland-symbols.json")
    if not os.path.exists(path):
        return {}
    names = collections.defaultdict(dict)
    for symbol in load_json(path)["symbols"]:
        if symbol.get("overlay_code_offset") and symbol.get("overlay_module"):
            names[symbol["overlay_module"]].setdefault(symbol["offset"], symbol["name"])
    return names


def resolve_far(target, by_segment, pc98_names, index_names):
    """`seg:off` → 「module entry#N 名稱」。off 是 control block 的 stub offset。"""
    try:
        segment, offset = (int(part, 16) for part in target.split(":"))
    except ValueError:
        return None
    overlay = by_segment.get(segment)
    if overlay is None:
        return None
    if offset < 0x20 or (offset - 0x20) % 5:
        return "%s stub?%04Xh" % (overlay["module"], offset)
    index = (offset - 0x20) // 5
    entries = overlay["entries"]
    if index >= len(entries):
        return "%s entry#%d(超出)" % (overlay["module"], index)
    name = index_names.get((overlay["module"], index))
    return "%s entry#%d %s" % (overlay["module"], index, name or "")


def main():
    pc98_names = symbol_names()
    # entry index → 名稱（由 PC-98 建立，兩平台共用；見 spec 562）
    pc98_segments, _ = overlay_index("pc98")
    index_names = {}
    for overlay in pc98_segments.values():
        for entry in overlay["entries"]:
            name = pc98_names.get(overlay["module"], {}).get(entry["code_offset"])
            if name:
                index_names[(overlay["module"], entry["index"])] = name

    print("# 函式分流報告")
    print()
    print("由 `scripts/profile_triage.py` 產生，不要手改。**只排序，不改台帳狀態。**")
    print("分組是候選判斷：`thunk` 可能不只是 thunk，「只呼叫 RTL」也可能是關鍵邏輯。")
    print()

    for platform in ("dos", "pc98"):
        by_segment, _ = overlay_index(platform)
        groups = collections.defaultdict(list)
        totals = 0
        helper_use = collections.Counter()

        for path in sorted(glob.glob(os.path.join(SWEEP, platform, "profiles", "*.json"))):
            if path.endswith(".error.log"):
                continue
            module = os.path.basename(path)[:-5].replace(".bin", "")
            for item in load_json(path)["profiles"]:
                totals += 1
                far = item["far_calls"]
                resolved = []
                for target, count in far.items():
                    label = resolve_far(target, by_segment, pc98_names, index_names)
                    if label:
                        resolved.append((label, count))
                        helper_use[label] += count
                record = (module, item["ea"], item["name"], item["size"],
                          len(item["callers"]), resolved)

                if item["ports"]:
                    groups["硬體存取（有 in／out）"].append(record)
                elif item["interrupts"]:
                    groups["軟體中斷（有 int）"].append(record)
                elif item["size"] <= 16 and len(far) + len(item["near_calls"]) <= 1:
                    groups["極小函式／thunk 候選"].append(record)
                elif not far and not item["near_calls"] and not item["data_writes"]:
                    groups["無呼叫且無寫入（純計算候選）"].append(record)
                elif item["strings"]:
                    groups["有字串參照"].append(record)
                else:
                    groups["其他"].append(record)

        print("## %s（%d 個函式）" % (platform.upper(), totals))
        print()
        print("| 分組 | 數量 |")
        print("|---|---:|")
        for name in sorted(groups, key=lambda k: -len(groups[k])):
            print("| %s | %d |" % (name, len(groups[name])))
        print()
        print("最常被呼叫的共用 routine（依 far call 次數）：")
        print()
        print("| routine | 次數 |")
        print("|---|---:|")
        for label, count in helper_use.most_common(15):
            print("| %s | %d |" % (label, count))
        print()
        biggest = sorted((r for rows in groups.values() for r in rows),
                         key=lambda r: -r[3])[:20]
        print("最大的 20 個函式（優先解讀候選）：")
        print()
        print("| 模組 | 位址 | 大小 | 被呼叫 | 主要 far call |")
        print("|---|---|---:|---:|---|")
        for module, ea, name, size, callers, resolved in biggest:
            top = "、".join("%s×%d" % (label, count)
                            for label, count in sorted(resolved, key=lambda x: -x[1])[:2])
            print("| %s | `%04Xh` | %d | %d | %s |" % (module, ea, size, callers, top or "—"))
        if platform != "pc98":
            print()
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
