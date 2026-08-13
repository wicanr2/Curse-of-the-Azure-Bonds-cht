"""複驗所有標為「已解讀」的函式：IDA 的邊界對嗎？

`overlay-02:0C81h` 的教訓——IDA 標 53 bytes、實際 396，而且漏掉的正好是最
關鍵的兩段（opcode 二次判讀與另一半分支）。**靠 IDA 的函式清單讀，會得到
一支空殼，而且不會有任何警告。**

所以每一筆「已解讀」都要回頭問：我當初讀的範圍，是這支函式的全部嗎？

判準分兩步：

1. **該位址本身要以 `55 89 e5` 開頭**，才算得上一支函式的起點。不是的話
   本檢查不適用——那是某支大函式的中段或共用出口。這類另外列出。
2. **IDA 標的結尾處必須是 `retf` 或 `retn`。** 是的話邊界正確；不是的話，
   函式在那裡還沒結束，當初就沒讀完。

不要用「下一個 prologue」當真實邊界：函式之間常夾著 Pascal 字串常數
（`HEALDUDE` 後面就是兩則訊息），那會把資料算成漏讀的程式碼，
`overlay-23:2419h` 因此被誤報成「少讀 48 bytes」。

用法：python3 scripts/verify_read_bounds.py [--fix]
    --fix 把有落差的條目退回「待解讀」，並在 note 裡寫明原因。
"""

import json
import os
import sys

ROOT = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
SWEEP = os.path.join(ROOT, "workplace", "re-sweep")
LEDGER = os.path.join(ROOT, "docs", "audit", "re-function-ledger.json")
INDEX = os.path.join(ROOT, "docs", "audit", "coab-function-index.json")
TOLERANCE = 16


def blob(platform, module):
    path = os.path.join(SWEEP, platform, "overlays", module + ".bin")
    return open(path, "rb").read() if os.path.exists(path) else None


def main():
    fix = "--fix" in sys.argv
    ledger = json.load(open(LEDGER, encoding="utf-8"))
    sizes = {(f["platform"], f["module"], f["ea"]): f["size"]
             for f in json.load(open(INDEX, encoding="utf-8"))["functions"]}

    cache, suspect, checked, not_prologue = {}, [], 0, []
    for entry in ledger["functions"]:
        if entry["state"] != "已解讀":
            continue
        key = (entry["platform"], entry["module"])
        if key not in cache:
            cache[key] = blob(*key)
        data = cache[key]
        if data is None:            # resident（.EXE）沒有 raw bin，跳過
            continue
        ea = entry["ea"]
        size = sizes.get((entry["platform"], entry["module"], ea))
        if size is None:
            continue
        checked += 1
        if data[ea:ea + 3] != b"\x55\x89\xe5":
            not_prologue.append((entry, size))
            continue
        end = ea + size
        tail = data[max(0, end - 3):end]
        ends_clean = (tail[-1:] in (b"\xcb", b"\xc3")      # retf／retn
                      or tail[:1] in (b"\xca", b"\xc2"))   # retf imm16／retn imm16
        if ends_clean:
            continue
        nxt = data.find(b"\x55\x89\xe5", ea + 3)
        if nxt < 0:
            nxt = len(data)
        suspect.append((entry, size, nxt - ea, nxt - end))

    suspect.sort(key=lambda r: -r[3])
    print("複驗 %d 筆已解讀（overlay 部分；resident 沒有 raw bin，未納入）" % checked)
    print("起點不是 prologue、不適用本檢查的：%d 筆（多為共用出口或大函式中段）"
          % len(not_prologue))
    print("IDA 標的結尾不是 ret，代表當初沒讀完的：%d 筆" % len(suspect))
    for entry, size, real, gap in suspect[:25]:
        print("  %-5s %-12s %04Xh  IDA=%-5d 實際=%-5d 差 %d"
              % (entry["platform"], entry["module"], entry["ea"], size, real, gap))

    if not fix:
        print("\n（唯讀模式；加 --fix 才把有落差的退回待解讀）")
        return 0

    keys = {(e["platform"], e["module"], e["ea"]) for e, _, _, _ in suspect}
    for entry in ledger["functions"]:
        if (entry["platform"], entry["module"], entry["ea"]) in keys:
            row = next(r for r in suspect
                       if (r[0]["platform"], r[0]["module"], r[0]["ea"])
                       == (entry["platform"], entry["module"], entry["ea"]))
            entry["state"] = "待解讀"
            entry["note"] = ("退回待解讀：當初依 IDA 的 %d bytes 判讀，但那個位置"
                             "不是 `ret`——函式在那裡還沒結束（下一個 prologue 在 "
                             "+%d）。要用 `scripts/show.py --whole` 重讀。原判讀：%s"
                             % (row[1], row[2], entry.get("note", "")))
    json.dump(ledger, open(LEDGER, "w", encoding="utf-8"), ensure_ascii=False, indent=1)
    print("\n已把 %d 筆退回待解讀" % len(keys))
    return 0


if __name__ == "__main__":
    sys.exit(main())
