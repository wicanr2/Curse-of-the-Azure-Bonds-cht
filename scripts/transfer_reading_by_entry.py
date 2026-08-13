"""把已讀完的 PC-98 overlay 函式判讀，依 entry index 轉移到 DOS 同一支。

依據（兩層，缺一不可）：

1. **entry index 相同**。兩平台的 TPOV control block stub offset 與 entry index
   一致（spec 562），所以 entry#N 在兩邊指的是原始碼裡的同一支程序。
2. **助憶碼序列完全相同**。逐指令比對 mnemonic（不比運算元——DS 位址與
   overlay-local 位址本來就不同）。序列不同就不轉移。

因此轉移過去的等級是 `strong inference` 而不是 `exact`：結構同構是強證據，
但**運算元裡的位址必須各自確認**——PC-98 的 `DS:A02Eh` 不等於 DOS 的同一個
數字。條目註記會明寫這一點。

用法：
    python3 scripts/transfer_reading_by_entry.py <overlay 模組名> [--write] [--dos-to-pc98]

預設方向是 PC-98 → DOS；加 `--dos-to-pc98` 走反向。兩個方向都要跑，因為兩邊
各自讀完的函式不一樣，只跑單向會漏掉另一邊已經讀完、這邊還空著的那些。
"""

import json
import os
import re
import sys

ROOT = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
SWEEP = os.path.join(ROOT, "workplace", "re-sweep")
LEDGER = os.path.join(ROOT, "docs", "audit", "re-function-ledger.json")


def mnemonics(function):
    out = []
    for item in function["items"]:
        text = re.sub(r"\s*;.*$", "", item["disasm"].strip())
        out.append(text.split()[0] if text else "")
    return tuple(out)


def entries(platform, module):
    manifest = json.load(open(os.path.join(SWEEP, platform, "ovr-manifest.json"),
                              encoding="utf-8"))
    overlay = next(o for o in manifest["overlays"] if o["module"] == module)
    return {e["index"]: e["code_offset"] for e in overlay["entries"]}


def small(platform, module):
    """函式本體，優先取 prologue 匯出。

    `small/` 匯出並不完整（`overlay-07` 缺 14 支、`overlay-22` 缺 16 支），
    缺的那些會被算成「缺匯出」而整批放棄轉移。prologue 匯出以 `55 89 e5`／
    `55 8b ec` 為界，涵蓋整個模組。
    """
    path = os.path.join(SWEEP, platform, "overlays", "prologue",
                        "%s-%s.json" % (platform, module))
    if os.path.exists(path):
        return {f["ea"]: f for f in json.load(open(path, encoding="utf-8"))["functions"]}
    path = os.path.join(SWEEP, platform, "small", module + ".bin.json")
    if not os.path.exists(path):
        return {}
    return {f["ea"]: f for f in json.load(open(path, encoding="utf-8"))["functions"]}


def main():
    if len(sys.argv) < 2:
        print(__doc__)
        return 2
    module = sys.argv[1]
    write = "--write" in sys.argv
    source_platform = "dos" if "--dos-to-pc98" in sys.argv else "pc98"
    target_platform = "pc98" if source_platform == "dos" else "dos"

    source_entries = entries(source_platform, module)
    target_entries = entries(target_platform, module)
    source_small = small(source_platform, module)
    target_small = small(target_platform, module)
    index_of = {offset: index for index, offset in source_entries.items()}

    ledger = json.load(open(LEDGER, encoding="utf-8"))
    already = {(e["platform"], e["module"], e["ea"]) for e in ledger["functions"]
               if e["state"] != "待解讀"}
    read_source = [e for e in ledger["functions"]
                   if e["platform"] == source_platform and e["module"] == module
                   and e["state"] == "已解讀"]

    transferred, mismatched, missing = [], 0, 0
    for entry in read_source:
        index = index_of.get(entry["ea"])
        if index is None or index not in target_entries:
            continue
        target_ea = target_entries[index]
        if (target_platform, module, target_ea) in already:
            continue
        source = source_small.get(entry["ea"])
        target = target_small.get(target_ea)
        if source is None or target is None:
            missing += 1
            continue
        if mnemonics(source) != mnemonics(target):
            mismatched += 1
            continue
        transferred.append({
            "platform": target_platform, "module": module, "ea": target_ea,
            "state": "已解讀", "level": "strong inference", "spec": entry["spec"],
            "note": "與 %s %s:%04Xh（entry#%d）助憶碼序列完全相同，語意同該筆："
                    "%s ⚠ 運算元中的 DS／overlay-local 位址兩平台不同，"
                    "引用位址前須各自確認" % (source_platform.upper(), module,
                                             entry["ea"], index, entry["note"]),
        })

    print("可轉移：%d；序列不同而略過：%d；缺匯出：%d"
          % (len(transferred), mismatched, missing))
    if not write:
        print("（預覽模式；加 --write 才寫入台帳）")
        return 0

    ledger["functions"].extend(transferred)
    json.dump(ledger, open(LEDGER, "w", encoding="utf-8"), ensure_ascii=False, indent=1)
    print("已寫入 %s" % LEDGER)
    return 0


if __name__ == "__main__":
    sys.exit(main())

