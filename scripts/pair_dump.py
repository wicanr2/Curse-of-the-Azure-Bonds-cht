"""列出「兩平台同一個 entry、助憶碼序列相同、而且兩邊都還沒判讀」的函式配對。

這種配對讀一支就能涵蓋兩支——**前提是把另一支的運算元差異逐條列出來**
（作法與 spec 748 的雲霧雙胞胎相同）。差異若只落在 DS 位址與 overlay-local
位址，兩邊的語意就是同一份；只要出現常數或指令運算元的實質差別，就必須各自
判讀，不能當成同一支。

用法：
    python3 scripts/pair_dump.py                     # 列出所有配對（依指令數排序）
    python3 scripts/pair_dump.py <模組> <dos ea 十六進位>   # 印出 DOS 本體＋差異表
"""

import collections
import json
import os
import re
import sys

ROOT = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
SWEEP = os.path.join(ROOT, "workplace", "re-sweep")
INDEX = os.path.join(ROOT, "docs", "audit", "coab-function-index.json")
RET = ("retf", "retn", "ret")
MIN_LENGTH = 12


def load(platform, module):
    for folder in ("filled", "prologue"):
        path = os.path.join(SWEEP, platform, "overlays", folder,
                            "%s-%s.json" % (platform, module))
        if os.path.exists(path):
            return {f["ea"]: f
                    for f in json.load(open(path, encoding="utf-8"))["functions"]}
    return {}


def clean(item):
    return re.sub(r"\s+", " ", re.sub(r"\s*;.*$", "", item["disasm"].strip()))


def body(function):
    items = function["items"]
    index = [i for i, item in enumerate(items)
             if clean(item).split(" ")[0] in RET]
    return items[:index[-1] + 1] if index else items


def shape(function):
    return tuple(clean(item).split(" ")[0] for item in body(function))


def entries(platform, module):
    path = os.path.join(SWEEP, platform, "ovr-manifest.json")
    manifest = json.load(open(path, encoding="utf-8"))
    overlay = next((o for o in manifest["overlays"] if o["module"] == module), None)
    return {e["index"]: e["code_offset"] for e in overlay["entries"]} if overlay else {}


def find_pairs():
    rows = [r for r in json.load(open(INDEX, encoding="utf-8"))["functions"]
            if r.get("state") == "待解讀"]
    pending = collections.defaultdict(set)
    for row in rows:
        pending[(row["platform"], row["module"])].add(row["ea"])

    out = []
    for module in sorted({m for _, m in pending if m.startswith("overlay-")}):
        dos_entries, pc98_entries = entries("dos", module), entries("pc98", module)
        dos, pc98 = load("dos", module), load("pc98", module)
        index_of = {offset: index for index, offset in dos_entries.items()}
        for ea in sorted(pending.get(("dos", module), ())):
            index = index_of.get(ea)
            if index is None or index not in pc98_entries:
                continue
            other = pc98_entries[index]
            if other not in pending.get(("pc98", module), ()):
                continue
            a, b = dos.get(ea), pc98.get(other)
            if not a or not b:
                continue
            if shape(a) == shape(b) and len(shape(a)) >= MIN_LENGTH:
                out.append((module, ea, other, len(shape(a))))
    return out


def main():
    if len(sys.argv) < 3:
        pairs = find_pairs()
        print("配對 %d 組，涵蓋 %d 支" % (len(pairs), len(pairs) * 2))
        for module, ea, other, count in sorted(pairs, key=lambda p: p[3]):
            print("  %-11s dos %05X ↔ pc98 %05X  %d 條" % (module, ea, other, count))
        return 0

    module, ea = sys.argv[1], int(sys.argv[2], 16)
    dos, pc98 = load("dos", module), load("pc98", module)
    dos_entries, pc98_entries = entries("dos", module), entries("pc98", module)
    index = {offset: i for i, offset in dos_entries.items()}[ea]
    other = pc98_entries[index]
    a, b = body(dos[ea]), body(pc98[other])

    print("=== dos %s:%05Xh ↔ pc98 %05Xh（entry#%d，%d 條）"
          % (module, ea, other, index, len(a)))
    for item in a:
        print("  %04X  %s" % (item["ea"], item["disasm"]))
    print("--- 運算元差異（扣掉分支目標之外的全部列出）")
    shown = 0
    for x, y in zip(a, b):
        dx, dy = clean(x), clean(y)
        if dx == dy:
            continue
        head = dx.split(" ")[0]
        if head.startswith("j") or head == "loop":
            continue          # 分支目標本來就不同
        shown += 1
        print("  %2d  %04X %-34s | %04X %s" % (shown, x["ea"], dx, y["ea"], dy))
    print("  （實質差異 %d 條）" % shown)
    return 0


if __name__ == "__main__":
    sys.exit(main())
