"""把「落在已解讀函式內部」的 IDA 假函式標成邊界碎片。

IDA 會把一支 Turbo Pascal 函式切成好幾塊（共用出口、被 `jmp` 跳過的中段）。
那些碎塊各自出現在**全函式索引**裡（台帳沒有它們的條目，所以狀態預設是
`待解讀`），看起來像獨立的函式，但它們的指令**已經隨所屬函式讀過了**。

注意要掃的是索引不是台帳：台帳只記錄明確標過的條目，這些碎塊從來沒被標過。

判準（兩條都要成立）：

1. 該位址落在某支函式的 prologue 區間 `(start, end)` 內，且**不等於 start**。
   擁有者讀了沒有不影響分類——它本來就不是獨立的一支，只是註記不同。
2. 它自己**不是** prologue（不以 `55 89 e5` 或 `55 8b ec` 開頭）——是的話它
   就是獨立的一支，不能算碎片。

3. **擁有者的 IDA 範圍不能以 `ret` 結束。** 以 `ret` 結束代表 IDA 標的邊界
   就是真邊界，那支函式很短、後面的東西不屬於它——`pc98/overlay-16:5581h`
   是 `55 89 e5 89 ec 5d cb`（7 bytes 的空函式），但它後面既無 `55 89 e5`
   也無 `55 8b ec`，prologue 區間一路延伸 494 bytes，會把四個不相干的位址
   誤判成它的內部。

第 2 條防的是「把真函式當碎片」，第 3 條防的是「prologue 掃描找不到邊界」。
兩者是本專案在複驗已解讀條目時踩過的同兩個坑。

用法：python3 scripts/mark_interior_fragments.py [--write]
"""

import glob
import json
import os
import sys

ROOT = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
SWEEP = os.path.join(ROOT, "workplace", "re-sweep")
LEDGER = os.path.join(ROOT, "docs", "audit", "re-function-ledger.json")


def prologue_ranges():
    out = {}
    for path in glob.glob(os.path.join(SWEEP, "*", "overlays", "prologue", "*.json")):
        name = os.path.basename(path)[:-5]
        platform, module = name.split("-", 1)
        for function in json.load(open(path, encoding="utf-8"))["functions"]:
            out[(platform, module, function["ea"])] = function
    return out


def main():
    write = "--write" in sys.argv
    ledger = json.load(open(LEDGER, encoding="utf-8"))
    ranges = prologue_ranges()
    read_entries = {(e["platform"], e["module"], e["ea"]) for e in ledger["functions"]
                    if e["state"] == "已解讀"}
    index = json.load(open(os.path.join(ROOT, "docs", "audit",
                                        "coab-function-index.json"), encoding="utf-8"))
    pending = [f for f in index["functions"] if f["state"] == "待解讀"]

    blobs = {}
    def blob(platform, module):
        key = (platform, module)
        if key not in blobs:
            path = os.path.join(SWEEP, platform, "overlays", module + ".bin")
            blobs[key] = open(path, "rb").read() if os.path.exists(path) else None
        return blobs[key]

    hits = []
    for entry in pending:
        data = blob(entry["platform"], entry["module"])
        if data is None:
            continue
        ea = entry["ea"]
        if data[ea:ea + 3] in (b"\x55\x89\xe5", b"\x55\x8b\xec"):
            continue                          # 自己是 prologue ⇒ 獨立的一支
        owner = None
        for (platform, module, start), function in ranges.items():
            if (platform, module) != (entry["platform"], entry["module"]):
                continue
            if not (start < ea < function["end"]):
                continue
            # 擁有者還沒讀也算碎片——它本來就不是獨立的一支，只是狀態註記不同。
            pass
            ida = function["ida_size"]
            if ida:
                tail = data[max(0, start + ida - 3):start + ida]
                if tail[-1:] in (b"\xcb", b"\xc3") or tail[:1] in (b"\xca", b"\xc2"):
                    continue                  # 擁有者以 ret 結束 ⇒ 邊界是真的
            owner = (start, (platform, module, start) in read_entries)
            break
        if owner is not None:
            hits.append((entry, owner))

    done = sum(1 for _, (_, r) in hits if r)
    print("落在某支函式內部、且自己不是 prologue 的：%d 筆（其中擁有者已解讀 %d 筆）"
          % (len(hits), done))
    for entry, (owner, was_read) in hits[:20]:
        print("  %-5s %-12s %04Xh ⊂ %04Xh %s" % (entry["platform"], entry["module"],
                                                 entry["ea"], owner,
                                                 "（已讀）" if was_read else "（擁有者未讀）"))
    if not write:
        print("\n（唯讀模式；加 --write 才標成邊界碎片）")
        return 0

    lookup = {(e["platform"], e["module"], e["ea"]): o for e, o in hits}
    existing = {(e["platform"], e["module"], e["ea"]) for e in ledger["functions"]}
    for (platform, module, ea), (owner, was_read) in lookup.items():
        note = ("邊界碎片：落在 %04Xh 的 prologue 區間內部，自己不是 prologue。%s"
                % (owner, "指令已隨該函式讀過。" if was_read
                   else "所屬函式尚未解讀，讀它時會一併涵蓋。"))
        row = {"platform": platform, "module": module, "ea": ea,
               "state": "邊界碎片", "level": "",
               "spec": "docs/spec/587-ecl-handler-21-37-shared.md", "note": note}
        if (platform, module, ea) in existing:
            for e in ledger["functions"]:
                if (e["platform"], e["module"], e["ea"]) == (platform, module, ea):
                    e.update(row)
        else:
            ledger["functions"].append(row)
    json.dump(ledger, open(LEDGER, "w", encoding="utf-8"), ensure_ascii=False, indent=1)
    print("\n已標記 %d 筆" % len(lookup))
    return 0


if __name__ == "__main__":
    sys.exit(main())
