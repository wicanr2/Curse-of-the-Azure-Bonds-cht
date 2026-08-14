"""列出目前還沒有台帳列的待解讀函式，依**實際**指令數排序。

⚠ `docs/audit/function-index/*.md` 的「指令」欄用的是 IDA 原始邊界，
常常與 dump 裡的實際函式大小不符（實測 `dos overlay-21:198Ch` 索引寫 40 條、
dump 是 823 條）。挑工作時要看實際大小，否則會誤判成本。

用法：
    python3 scripts/pending.py [筆數]
"""

import glob
import json
import os
import sys

sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
from module_align import load, LEDGER

ROOT = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))


def main():
    limit = int(sys.argv[1]) if len(sys.argv) > 1 else 30
    ledger = json.load(open(LEDGER, encoding="utf-8"))["functions"]
    done = {(r["platform"], r["module"], r["ea"]) for r in ledger}
    cache, rows = {}, []
    for path in sorted(glob.glob(os.path.join(
            ROOT, "docs", "audit", "function-index", "*.md"))):
        mod = os.path.basename(path)[:-3]
        if mod.endswith(".bin"):
            continue
        plat, name = mod.split("-", 1)
        for line in open(path, encoding="utf-8"):
            if not line.startswith("| `"):
                continue
            col = line.split("|")
            if len(col) <= 10 or "待解讀" not in col[9]:
                continue
            ea = int(col[1].strip().strip("`"), 16)
            if (plat, name, ea) in done:
                continue
            if (plat, name) not in cache:
                cache[(plat, name)] = {f["items"][0]["ea"]: len(f["items"])
                                       for f in load(plat, name)}
            real = cache[(plat, name)].get(ea)
            rows.append((real if real else 10 ** 6, plat, name, ea))
    rows.sort()
    print("待解讀（無台帳列）%d 筆" % len(rows))
    for real, plat, name, ea in rows[:limit]:
        print("  %6s 條  %-5s %-14s %04X"
              % ("無 dump" if real >= 10 ** 6 else real, plat, name, ea))


main()
