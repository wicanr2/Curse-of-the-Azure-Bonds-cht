"""用「助憶碼序列相同」的跨平台配對函式，量出欄位偏移的對應關係。

兩平台是同一份原始碼編譯的，所以同一支程序的指令種類與順序一致，只有運算元
不同。把配對函式的 `[di+XXXh]` 依序對起來，就能讀出「PC-98 的某個偏移在 DOS
是哪個」——**不必猜、也不必先知道欄位是什麼**。

配對判準（保守）：助憶碼序列在各自平台的該模組內唯一、長度 `>= MIN_LEN`、
兩邊取出的偏移個數相同。

用法：
    python3 scripts/field_offset_map.py [模組名...]     # 預設掃全部
"""

import collections
import json
import os
import re
import sys

ROOT = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
SWEEP = os.path.join(ROOT, "workplace", "re-sweep")
OUT = os.path.join(ROOT, "docs", "audit", "cross-platform-field-offsets.json")
MIN_LEN = 12


def load(platform, module):
    path = os.path.join(SWEEP, platform, "overlays", "prologue",
                        "%s-%s.json" % (platform, module))
    if not os.path.exists(path):
        return {}
    return {f["ea"]: f for f in json.load(open(path, encoding="utf-8"))["functions"]}


def mnemonics(function):
    return tuple(re.sub(r"\s*;.*$", "", i["disasm"].strip()).split()[0]
                 for i in function["items"])


def offsets(function):
    out = []
    for item in function["items"]:
        out.extend(int(m.group(1), 16)
                   for m in re.finditer(r"\[di\+([0-9A-F]+)h\]", item["disasm"]))
    return out


def main():
    modules = sys.argv[1:]
    if not modules:
        directory = os.path.join(SWEEP, "pc98", "overlays", "prologue")
        modules = sorted({n[len("pc98-"):-len(".json")]
                          for n in os.listdir(directory) if n.startswith("pc98-")})

    votes = collections.Counter()
    pairs = 0
    for module in modules:
        pc98, dos = load("pc98", module), load("dos", module)
        if not pc98 or not dos:
            continue
        by_pc98 = collections.defaultdict(list)
        by_dos = collections.defaultdict(list)
        for ea, f in pc98.items():
            by_pc98[mnemonics(f)].append(ea)
        for ea, f in dos.items():
            by_dos[mnemonics(f)].append(ea)
        for signature, here in by_pc98.items():
            there = by_dos.get(signature)
            if not there or len(here) != 1 or len(there) != 1 or len(signature) < MIN_LEN:
                continue
            a, b = offsets(pc98[here[0]]), offsets(dos[there[0]])
            if not a or len(a) != len(b):
                continue
            pairs += 1
            for x, y in zip(a, b):
                votes[(x, y)] += 1

    # 同一個 PC-98 偏移可能對到多個 DOS 偏移（配對本身有雜訊），取票數最高的。
    best = {}
    for (x, y), count in votes.items():
        if x not in best or count > best[x][1]:
            best[x] = (y, count)

    print("配對函式 %d 對，量到 %d 個偏移" % (pairs, len(best)))
    ambiguous = 0
    rows = []
    for x in sorted(best):
        y, count = best[x]
        total = sum(c for (a, _), c in votes.items() if a == x)
        if count < total:
            ambiguous += 1
        rows.append({"pc98": x, "dos": y, "delta": x - y,
                     "votes": count, "total": total})
        print("  pc98 +%04Xh ↔ dos +%04Xh  差 %+d  %d/%d 票"
              % (x, y, x - y, count, total))
    print("\n票數不一致（同一偏移對到多個）：%d" % ambiguous)
    json.dump({"schema": "coab-cross-platform-field-offsets/1",
               "min_signature_length": MIN_LEN, "pairs": pairs, "offsets": rows},
              open(OUT, "w", encoding="utf-8"), ensure_ascii=False, indent=1)
    print("→ %s" % OUT)
    return 0


if __name__ == "__main__":
    sys.exit(main())
