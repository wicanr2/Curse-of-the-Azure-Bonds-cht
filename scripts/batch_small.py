"""把某模組所有「待解讀且不超過 N bytes」的函式一次印完，供逐支閱讀。

一支 60 bytes 的 Turbo Pascal 函式大約 20 條指令，逐支開 `show.py` 的往返成本
遠高於閱讀本身。這支把它們接在一起印，並把 `mov di, offset` 指到的 Pascal 短
字串就地解出來——那通常就是整支函式在做什麼的答案。

用法：
    python3 scripts/batch_small.py <platform> <module> [最大 bytes] [起始序號] [支數]
"""

import json
import os
import re
import sys

ROOT = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
SWEEP = os.path.join(ROOT, "workplace", "re-sweep")


def main():
    platform, module = sys.argv[1], sys.argv[2]
    limit = int(sys.argv[3]) if len(sys.argv) > 3 else 60
    start = int(sys.argv[4]) if len(sys.argv) > 4 else 0
    count = int(sys.argv[5]) if len(sys.argv) > 5 else 999

    index = json.load(open(os.path.join(ROOT, "docs", "audit",
                                        "coab-function-index.json"), encoding="utf-8"))
    functions = index.get("functions") or index
    ledger = {(e["platform"], e["module"], e["ea"]): e["state"] for e in
              json.load(open(os.path.join(ROOT, "docs", "audit",
                                          "re-function-ledger.json"),
                             encoding="utf-8"))["functions"]}
    todo = sorted((f for f in functions
                   if f["platform"] == platform and f["module"] == module
                   and (f.get("size") or 0) <= limit
                   and ledger.get((platform, module, f["ea"]), "待解讀") == "待解讀"),
                  key=lambda f: f["ea"])[start:start + count]

    path = os.path.join(SWEEP, platform, "overlays", "prologue",
                        "%s-%s.json" % (platform, module))
    if module.endswith(".EXE") or not os.path.exists(path):
        # resident 執行檔不走 prologue 邊界（第 644 輪：271 支只有 85 支對得上，
        # 手寫組語的 RTL 與 TPOV stub 都沒有標準 prologue）。改用 IDA 邊界，
        # 並在每支結尾標出最後一條是不是 return。
        for candidate in (module + ".json", module + ".big.json"):
            trial = os.path.join(SWEEP, platform, "small", candidate)
            if os.path.exists(trial):
                path = trial
                break
    dumped = json.load(open(path, encoding="utf-8"))["functions"]
    items = {}
    for function in dumped:
        for item in function["items"]:
            items.setdefault(item["ea"], item)

    blob_path = os.path.join(SWEEP, platform, "overlays", module + ".bin")
    blob = open(blob_path, "rb").read() if os.path.exists(blob_path) else b""

    def pascal(offset):
        if offset >= len(blob):
            return None
        length = blob[offset]
        if not 1 <= length <= 60 or offset + 1 + length > len(blob):
            return None
        try:
            text = blob[offset + 1:offset + 1 + length].decode(
                "cp932" if platform == "pc98" else "cp437")
        except UnicodeDecodeError:
            return None
        return text if all(c.isprintable() for c in text) else None

    def next_prologue(ea):
        """下一支的起點，用原始位元組找 prologue。

        不能用索引裡的 `size`：那來自 IDA，而 IDA 會低估（`overlay-12:147Ah`
        說 86 bytes、實際 155）。用 size 當上界會**在半途停住而且看起來像印完
        了**——這正是第 639 輪踩到的坑。
        """
        hits = [blob.find(pattern, ea + 3)
                for pattern in (b"\x55\x89\xe5", b"\x55\x8b\xec")]
        hits = [h for h in hits if h > 0]
        return min(hits) if hits else len(blob)

    print("%s %s：待解讀且 <= %d bytes 共 %d 支" % (platform, module, limit, len(todo)))
    for function in todo:
        ea, size = function["ea"], function.get("size") or 0
        end = next_prologue(ea) if blob else ea + size
        inside = [a for a in sorted(items) if ea <= a < end]
        tail = ""
        if not blob and inside:
            last = re.sub(r"\s*;.*$", "", items[inside[-1]]["disasm"].strip())
            if not re.match(r"(ret[nf]?|jmp|iret)\b", last):
                tail = "  ⚠ 結尾不是 return，IDA 範圍可能不完整"
        print("\n=== %04Xh  IDA size=%d，讀到 %04Xh ===%s" % (ea, size, end, tail))
        for address in inside:
            item = items[address]
            text = re.sub(r"\s*;.*$", "", item["disasm"].strip())
            match = re.match(r"bf(..)(..)", item["bytes"])
            if match:
                offset = int(match.group(1), 16) | (int(match.group(2), 16) << 8)
                got = pascal(offset)
                if got:
                    text += "        ; '%s'" % got
            print("  %04X  %s" % (address, text))
    return 0


if __name__ == "__main__":
    sys.exit(main())
