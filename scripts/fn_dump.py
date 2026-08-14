"""直接印出某支函式的完整指令序列（不做兩平台對齊）。

給常駐執行檔（`START.EXE`／`PC98-GAME.EXE`）用——那兩支沒有 overlay 配對，
`aligned_diff.py` 幫不上忙。

用法：
    python3 scripts/fn_dump.py <平台> <模組> <ea hex> [<ea hex> …]
    python3 scripts/fn_dump.py <平台> <模組> --small <上限條數> [--limit N]
"""

import sys
import os

sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
from module_align import load, clean


def show(f):
    ea = f["items"][0]["ea"]
    print("=== %05Xh（%d 條）" % (ea, len(f["items"])))
    for it in f["items"]:
        print("  %05X  %s" % (it["ea"], clean(it)))


def main():
    plat, mod = sys.argv[1], sys.argv[2]
    fns = load(plat, mod)
    if "--small" in sys.argv:
        cap = int(sys.argv[sys.argv.index("--small") + 1])
        limit = 8
        if "--limit" in sys.argv:
            limit = int(sys.argv[sys.argv.index("--limit") + 1])
        import json
        ledger = json.load(open(os.path.join(
            os.path.dirname(os.path.dirname(os.path.abspath(__file__))),
            "docs", "audit", "re-function-ledger.json"), encoding="utf-8"))
        rows = {(r["platform"], r["module"], r["ea"])
                for r in ledger["functions"]}
        pick = [f for f in fns
                if len(f["items"]) <= cap
                and (plat, mod, f["items"][0]["ea"]) not in rows]
        pick.sort(key=lambda f: len(f["items"]))
        for f in pick[:limit]:
            show(f)
        print("（符合條件共 %d 支）" % len(pick))
        return
    want = {int(a, 16) for a in sys.argv[3:]}
    for f in fns:
        if f["items"][0]["ea"] in want:
            show(f)


main()
