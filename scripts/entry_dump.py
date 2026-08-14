"""印出單一 overlay 函式的本體，並回報另一平台同 entry index 的狀態。

`pair_dump.py` 只處理「兩平台都未讀且助憶碼同形」的配對。實務上還有兩種情形
需要單獨看：

1. 只有一邊未讀（另一邊已經讀過，但 `transfer_reading_by_entry.py` 因為助憶碼
   序列不同而不肯轉移）。
2. 兩邊都未讀但序列不同——這種**一定要各自讀**，不能共用一份判讀。

用法：
    python3 scripts/entry_dump.py <平台> <模組> <ea 十六進位>

輸出：本體（截到最後一個 ret）＋另一平台同 entry index 的位址、指令數、
是否同形、台帳狀態。同形時會把運算元差異一併列出，這樣一次可以涵蓋兩支；
不同形時只印狀態，提醒必須分開讀。
"""

import json
import os
import re
import sys

ROOT = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
SWEEP = os.path.join(ROOT, "workplace", "re-sweep")
LEDGER = os.path.join(ROOT, "docs", "audit", "re-function-ledger.json")
RET = ("retf", "retn", "ret")


def load(platform, module):
    for sub in ("filled", "prologue"):
        path = os.path.join(SWEEP, platform, "overlays", sub,
                            "%s-%s.json" % (platform, module))
        if os.path.exists(path):
            data = json.load(open(path, encoding="utf-8"))
            return data["functions"] if isinstance(data, dict) else data
    raise SystemExit("找不到 %s %s 的匯出" % (platform, module))


def body(function):
    items = function["items"]
    index = [i for i, item in enumerate(items)
             if clean(item).split(" ")[0] in RET]
    return items[:index[-1] + 1] if index else items


def clean(item):
    return re.sub(r"\s+", " ", re.sub(r"\s*;.*$", "", item["disasm"].strip()))


def main():
    if len(sys.argv) != 4:
        raise SystemExit(__doc__)
    platform, module, ea = sys.argv[1], sys.argv[2], int(sys.argv[3], 16)
    other = "pc98" if platform == "dos" else "dos"

    here = load(platform, module)
    entries = [f["items"][0]["ea"] for f in here if f["items"]]
    if ea not in entries:
        raise SystemExit("%s %s 沒有 entry 在 %04Xh；可用的前幾個：%s"
                         % (platform, module, ea,
                            ", ".join("%04Xh" % e for e in entries[:8])))
    index = entries.index(ea)
    mine = body(here[index])

    print("=== %s %s:%05Xh（entry#%d，%d 條）"
          % (platform, module, ea, index + 1, len(mine)))
    for item in mine:
        print("  %04X  %s" % (item["ea"], clean(item)))

    there = load(other, module)
    if index >= len(there) or not there[index]["items"]:
        print("--- %s 沒有對應的 entry#%d" % (other, index + 1))
        return 0
    peer = body(there[index])
    peer_ea = peer[0]["ea"]

    state = None
    ledger = json.load(open(LEDGER, encoding="utf-8"))
    for entry in ledger["functions"]:
        if (entry["platform"], entry["module"], entry["ea"]) == (other, module, peer_ea):
            state = entry["state"]
            break

    same = ([clean(i).split(" ")[0] for i in mine]
            == [clean(i).split(" ")[0] for i in peer])
    print("--- %s 同 entry：%05Xh，%d 條，%s，台帳狀態 %s"
          % (other, peer_ea, len(peer), "同形" if same else "**不同形**",
             state or "（不在台帳）"))
    if not same:
        print("    不同形 → 兩支必須各自讀，不可共用判讀。")
        return 0
    print("--- 運算元差異")
    count = 0
    for a, b in zip(mine, peer):
        ta, tb = clean(a), clean(b)
        if ta == tb:
            continue
        head = ta.split(" ")[0]
        if head.startswith("j") or head in ("call", "loop"):
            if re.sub(r"(loc|sub|unk)_[0-9A-F]+", "L", ta) == \
               re.sub(r"(loc|sub|unk)_[0-9A-F]+", "L", tb):
                continue
        count += 1
        print("  %3d  %04X %-34s | %04X %s" % (count, a["ea"], ta, b["ea"], tb))
    print("  （實質差異 %d 條）" % count)
    return 0


if __name__ == "__main__":
    sys.exit(main())
