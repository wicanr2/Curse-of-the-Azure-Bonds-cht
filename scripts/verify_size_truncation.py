"""檢查台帳裡「已解讀」的條目，IDA 給的長度是不是把函式截斷了。

`batch_small.py` 用索引裡的 `size` 決定印到哪，而那個 size 來自 IDA。IDA 低估
時會**在半途停住而且看起來像讀完了**（第 639 輪的 `overlay-12:147Ah`：IDA 說
86 bytes，實際 155）。

判準：函式最後一條指令必須是 `ret`／`retn`／`retf`／`iret`／無條件 `jmp`。
若不是，就到下一個 prologue 之間看還有多少指令沒被算進去。

用 `show.py --whole` 讀完的條目請在台帳裡標 `"bounds": "prologue"`，本工具會跳過
它們——那些走的是 prologue 區間，IDA 的 `size` 短不短都不影響。沒有這個標記的條目
就是**待查**：要嘛確認讀完了並補上標記，要嘛重讀。

用法：
    python3 scripts/verify_size_truncation.py [platform] [module]
"""

import json
import os
import re
import sys

ROOT = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
SWEEP = os.path.join(ROOT, "workplace", "re-sweep")


def main():
    only_platform = sys.argv[1] if len(sys.argv) > 1 else None
    only_module = sys.argv[2] if len(sys.argv) > 2 else None

    index = json.load(open(os.path.join(ROOT, "docs", "audit",
                                        "coab-function-index.json"), encoding="utf-8"))
    functions = index.get("functions") or index
    sizes = {(f["platform"], f["module"], f["ea"]): (f.get("size") or 0)
             for f in functions}
    ledger = json.load(open(os.path.join(ROOT, "docs", "audit",
                                         "re-function-ledger.json"),
                            encoding="utf-8"))["functions"]

    cache = {}
    suspect = 0
    for entry in ledger:
        if entry["state"] != "已解讀":
            continue
        platform, module, ea = entry["platform"], entry["module"], entry["ea"]
        if only_platform and platform != only_platform:
            continue
        if only_module and module != only_module:
            continue
        if entry.get("bounds") == "prologue":
            continue
        size = sizes.get((platform, module, ea), 0)
        if not size:
            continue
        key = (platform, module)
        if key not in cache:
            path = os.path.join(SWEEP, platform, "overlays", "prologue",
                                "%s-%s.json" % (platform, module))
            if not os.path.exists(path):
                cache[key] = None
            else:
                items = {}
                for function in json.load(open(path, encoding="utf-8"))["functions"]:
                    for item in function["items"]:
                        items.setdefault(item["ea"], item)
                blob_path = os.path.join(SWEEP, platform, "overlays", module + ".bin")
                cache[key] = (items, open(blob_path, "rb").read())
        if cache[key] is None:
            continue
        items, blob = cache[key]
        inside = sorted(a for a in items if ea <= a < ea + size)
        if not inside:
            continue
        last = re.sub(r"\s*;.*$", "", items[inside[-1]]["disasm"].strip())
        # IDA 把 near return 寫成 `retn`，`retf?` 加 \b 匹配不到它。
        if re.match(r"(ret[nf]?|jmp|iret)\b", last):
            continue
        hits = [blob.find(p, ea + 3) for p in (b"\x55\x89\xe5", b"\x55\x8b\xec")]
        end = min([h for h in hits if h > 0] or [len(blob)])
        extra = [a for a in items if ea + size <= a < end]
        if extra:
            suspect += 1
            print("%-5s %-12s %04Xh  IDA size=%-5d 最後一條=%-18s 之後還有 %d 條指令（到 %04Xh）"
                  % (platform, module, ea, size, last, len(extra), end))
    print("\n可能被截斷的「已解讀」條目：%d" % suspect)
    return 0


if __name__ == "__main__":
    sys.exit(main())
