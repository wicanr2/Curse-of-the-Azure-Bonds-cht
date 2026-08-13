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

但也不能數區間內**所有**指令：prologue 判準只認 `55 89 e5`（Turbo Pascal），
組合語言寫的常式用的是 `55 8b ec`，於是那種函式的起點不被認成邊界，前一支的
區間就一路吃到它後面去。`pc98/overlay-11:0766h` 其實只有 5 條指令
（`push bp / mov bp,sp / mov sp,bp / pop bp / retf`），後面 96 bytes 是資料、
再後面 `07CDh` 才是下一支（`55 8b ec` 開頭）。

所以判準是兩條，缺一不可：

1. **IDA 標的範圍如果以 `ret` 結束，就相信 IDA**——那是一支完整的函式。
   `pc98/overlay-16:5581h` 是 `55 89 e5 89 ec 5d cb`，7 bytes 的空函式，
   結尾正是 `retf`；它後面既不是 `55 89 e5` 也不是 `55 8b ec`，prologue
   掃描找不到邊界，區間一路延伸到 494 bytes 外，看起來像漏讀了 213 條指令。
2. 不以 `ret` 結束的，才用 prologue 區間檢查，而且只數**從起點連續**的指令
   （遇到第一個 gap 就停）。

第 1 條擋掉「prologue 掃描找不到下一個邊界」造成的高估，第 2 條擋掉
「IDA 在第一個 `ret` 就截斷」造成的漏讀。

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


def last_ea(items):
    """最後一條指令的結束位址。"""
    return items[-1]["ea"] + len(items[-1]["bytes"]) // 2


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
        # 只看第一個 gap 之前的指令：gap 之後多半已經是下一支函式或資料。
        # 第一關：IDA 標的範圍以 ret 結束就是完整的。
        tail = [i for i in function["items"] if i["ea"] < entry["ea"] + ida]
        if tail:
            last = tail[-1]["disasm"].strip().split()[0].lower()
            if last in ("retf", "retn", "ret") and \
                    last_ea(tail) == entry["ea"] + ida:
                continue
        limit = function["gaps"][0][0] if function["gaps"] else function["end"]
        outside = sum(1 for item in function["items"]
                      if entry["ea"] + ida <= item["ea"] < limit)
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
