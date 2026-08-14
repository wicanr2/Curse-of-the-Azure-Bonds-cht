"""列出「只有一邊未讀、另一邊已解讀」的函式，附助憶碼相似度與差異塊數。

背景：`transfer_reading_by_entry.py` 只在助憶碼序列**完全相同**時才肯把判讀
轉到對側。實務上大量函式兩邊只差位址與字串，序列因為 `xor ah, ah` 之類的
編譯器碼型而不完全相同，於是一直卡在待解讀。

本工具**不做任何判定**，只把候選按「相似度高 → 差異塊少 → 條數少」排序，
讓人知道先看哪一批最省力。**要不要標已解讀，一律以逐條看過差異為準。**

用法：
    python3 scripts/one_sided_triage.py            # 全部
    python3 scripts/one_sided_triage.py <模組>     # 單一模組
"""

import difflib
import glob
import json
import os
import re
import sys

ROOT = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
SWEEP = os.path.join(ROOT, "workplace", "re-sweep")
LEDGER = os.path.join(ROOT, "docs", "audit", "re-function-ledger.json")
RET = ("retf", "retn", "ret")


def clean(item):
    return re.sub(r"\s+", " ", re.sub(r"\s*;.*$", "", item["disasm"].strip()))


def body(function):
    items = function["items"]
    index = [i for i, item in enumerate(items)
             if clean(item).split(" ")[0] in RET]
    return items[:index[-1] + 1] if index else items


def load_index():
    index = {}
    for platform in ("dos", "pc98"):
        for sub in ("filled", "prologue"):
            pattern = os.path.join(SWEEP, platform, "overlays", sub,
                                   "%s-overlay-*.json" % platform)
            for path in glob.glob(pattern):
                module = os.path.basename(path)
                module = module.replace(platform + "-", "").replace(".json", "")
                if (platform, module) in index:
                    continue
                data = json.load(open(path, encoding="utf-8"))
                functions = data["functions"] if isinstance(data, dict) else data
                index[(platform, module)] = [f for f in functions if f["items"]]
    return index


def main():
    only = sys.argv[1] if len(sys.argv) > 1 else None
    ledger = json.load(open(LEDGER, encoding="utf-8"))["functions"]
    state = {(e["platform"], e["module"], e["ea"]): e["state"] for e in ledger}
    index = load_index()

    rows = []
    for (platform, module), functions in index.items():
        if only and module != only:
            continue
        peer = "pc98" if platform == "dos" else "dos"
        peers = index.get((peer, module), [])
        for i, function in enumerate(functions):
            if i >= len(peers):
                continue
            ea = function["items"][0]["ea"]
            if state.get((platform, module, ea), "待解讀") != "待解讀":
                continue
            peer_ea = peers[i]["items"][0]["ea"]
            if state.get((peer, module, peer_ea)) != "已解讀":
                continue
            mine = [clean(x).split(" ")[0] for x in body(function)]
            theirs = [clean(x).split(" ")[0] for x in body(peers[i])]
            matcher = difflib.SequenceMatcher(None, mine, theirs)
            ratio = matcher.ratio()
            blocks = sum(1 for op in matcher.get_opcodes() if op[0] != "equal")
            rows.append((ratio, blocks, len(mine), platform, module, ea, peer_ea))

    rows.sort(key=lambda r: (-r[0], r[1], r[2]))
    print("候選 %d 筆（未讀側 → 已讀側）" % len(rows))
    for ratio, blocks, size, platform, module, ea, peer_ea in rows:
        print("  %.3f  差異塊 %2d  %4d 條  %-4s %s %05X ← 對側 %05X"
              % (ratio, blocks, size, platform, module, ea, peer_ea))


if __name__ == "__main__":
    main()
