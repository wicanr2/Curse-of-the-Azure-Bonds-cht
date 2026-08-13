"""用 prologue 邊界重新複驗「已解讀」條目——比 `ret` 判準嚴格。

先前的複驗（`verify_read_bounds.py`）只檢查「IDA 標的結尾是不是 `ret`」。
那條判準有漏洞：Turbo Pascal 的函式可以有**多個** `ret`（提早返回），IDA 在
第一個就截斷時結尾仍然是 `ret`，看起來完全正常，但後面還有一整段沒讀到。
`overlay-02:149Ch` 就是這樣——IDA 標 50 bytes，真實長度 544，結尾恰好是 `ret`。

prologue 匯出把「一支函式」定義成 `[55 89 e5, 下一個 55 89 e5)`，直接給出真實
長度。這裡拿它跟 IDA 的 size 比：差距大的，當初很可能只讀了 IDA 那一段。

判準看的是**指令數**，不是 bytes。函式後面常接著 Pascal 字串常數，
`dos/overlay-34:0000h` 的 prologue 區間有 576 bytes 但**只有 5 條指令**——
其餘全是資料，不是漏讀。真正的漏讀長這樣：`pc98/overlay-18:1756h` 的 IDA size
是 7 bytes，區間內卻有 **282 條指令**。

所以數「落在 IDA 標的範圍之外、但仍是指令」的條數，超過 `TOLERANCE` 條才算。

用法：python3 scripts/verify_read_prologue.py [--fix]
"""

import glob
import json
import os
import sys

ROOT = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
SWEEP = os.path.join(ROOT, "workplace", "re-sweep")
LEDGER = os.path.join(ROOT, "docs", "audit", "re-function-ledger.json")
TOLERANCE = 3      # 容許幾條落在 IDA 範圍外的指令


def prologue_index():
    out = {}
    for path in glob.glob(os.path.join(SWEEP, "*", "overlays", "prologue", "*.json")):
        name = os.path.basename(path)[:-5]
        platform, module = name.split("-", 1)
        for function in json.load(open(path, encoding="utf-8"))["functions"]:
            out[(platform, module, function["ea"])] = function
    return out


def main():
    fix = "--fix" in sys.argv
    ledger = json.load(open(LEDGER, encoding="utf-8"))
    index = prologue_index()

    checked, suspect = 0, []
    for entry in ledger["functions"]:
        if entry["state"] != "已解讀":
            continue
        function = index.get((entry["platform"], entry["module"], entry["ea"]))
        if function is None:
            continue                      # resident，或起點不是 prologue
        checked += 1
        ida = function["ida_size"]
        if ida is None:
            continue
        outside = sum(1 for item in function["items"]
                      if item["ea"] >= entry["ea"] + ida)
        if outside > TOLERANCE:
            suspect.append((entry, ida, function["size"], outside))

    suspect.sort(key=lambda r: -(r[2] - r[1]))
    print("複驗 %d 筆（起點是 prologue 且有匯出的）" % checked)
    print("IDA 範圍外還有超過 %d 條指令的：%d 筆" % (TOLERANCE, len(suspect)))
    for entry, ida, real, outside in suspect[:30]:
        print("  %-5s %-12s %04Xh  IDA=%-5d 真實=%-5d 範圍外指令=%-4d  %s"
              % (entry["platform"], entry["module"], entry["ea"], ida, real, outside,
                 entry["spec"].split("/")[-1][:28]))
    if not fix:
        print("\n（唯讀模式；加 --fix 才退回待解讀）")
        return 0

    keys = {(e["platform"], e["module"], e["ea"]) for e, _, _, _ in suspect}
    lookup = {(e["platform"], e["module"], e["ea"]): (i, o)
              for e, i, _, o in suspect}
    for entry in ledger["functions"]:
        key = (entry["platform"], entry["module"], entry["ea"])
        if key in keys and "真實範圍" not in entry.get("note", ""):
            ida, real = lookup[key]
            entry["state"] = "待解讀"
            entry["note"] = ("退回待解讀：當初依 IDA 的 %d bytes 判讀，但 prologue "
                             "區間內在那之後還有 %d 條指令。要用 "
                             "`scripts/show.py --whole` 重讀。原判讀：%s"
                             % (ida, real, entry.get("note", "")))
    json.dump(ledger, open(LEDGER, "w", encoding="utf-8"), ensure_ascii=False, indent=1)
    print("\n已處理")
    return 0


if __name__ == "__main__":
    sys.exit(main())
