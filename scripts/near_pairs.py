"""找「兩平台幾乎同形」的配對——`pair_dump.py` 因為一兩條差異而排除掉的那些。

起因（spec 811）：`overlay-25:03B1h` 兩平台的原始位元組完全相同，但 DOS 側的
IDA 在一個位置錯位，於是三條指令被解成 `iret` / `add bp, cx` 這種不可能出現在
Pascal 程式碼裡的東西。助憶碼序列因此不相等，`pair_dump.py` 判定「不同形」而
把整組排除。

這支列出「差異只佔極小比例」的配對，並把差異區塊印出來，讓人一眼判斷是
**反組譯錯位**（可以合併讀）還是**兩平台真的不同**（必須各自讀）。

判斷錯位的訊號：
- 差異區塊很短（一兩條），而且兩邊指令數只差 1..3。
- 其中一邊出現 `iret` / `into` / `hlt` / `in` / `out` / `xlat` / `sahf` 這類
  Turbo Pascal 不會產生的指令。
- 去對原始 `.bin` 的位元組，兩邊一樣。

用法：
    python3 scripts/near_pairs.py [最小比例]      # 預設 0.90
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
SUSPECT = ("iret", "into", "hlt", "in", "out", "xlat", "sahf", "lahf",
           "aaa", "aas", "aam", "aad", "daa", "das", "lock", "esc")


def load(platform, module):
    for sub in ("filled", "prologue"):
        path = os.path.join(SWEEP, platform, "overlays", sub,
                            "%s-%s.json" % (platform, module))
        if os.path.exists(path):
            data = json.load(open(path, encoding="utf-8"))
            return data["functions"] if isinstance(data, dict) else data
    return None


def names(function):
    out = []
    for item in function["items"]:
        text = re.sub(r"\s*;.*$", "", item["disasm"].strip())
        out.append(text.split()[0] if text else "")
    tail = [i for i, n in enumerate(out) if n in RET]
    return out[:tail[-1] + 1] if tail else out


def main():
    floor = float(sys.argv[1]) if len(sys.argv) > 1 else 0.90
    ledger = json.load(open(LEDGER, encoding="utf-8"))
    done = {(e["platform"], e["module"], e["ea"])
            for e in ledger["functions"] if e["state"] == "已解讀"}

    found = []
    for path in sorted(glob.glob(os.path.join(SWEEP, "dos", "overlays",
                                              "filled", "dos-overlay-*.json"))):
        module = "overlay-" + os.path.basename(path).split("overlay-")[1].split(".")[0]
        dos, pc98 = load("dos", module), load("pc98", module)
        if not dos or not pc98:
            continue
        for index, function in enumerate(dos):
            if index >= len(pc98):
                break
            if not function["items"] or not pc98[index]["items"]:
                continue
            dos_ea = function["items"][0]["ea"]
            pc98_ea = pc98[index]["items"][0]["ea"]
            if ("dos", module, dos_ea) in done or ("pc98", module, pc98_ea) in done:
                continue
            a, b = names(function), names(pc98[index])
            if a == b or not a or not b:
                continue
            ratio = difflib.SequenceMatcher(None, a, b).ratio()
            if ratio < floor:
                continue
            blocks = [op for op in difflib.SequenceMatcher(None, a, b).get_opcodes()
                      if op[0] != "equal"]
            odd = [n for op in blocks
                   for n in a[op[1]:op[2]] + b[op[3]:op[4]] if n in SUSPECT]
            found.append((ratio, module, dos_ea, pc98_ea, len(a), len(b),
                          blocks, odd))

    found.sort(key=lambda row: (-row[0], row[1]))
    print("幾乎同形的配對 %d 組（比例 >= %.2f）" % (len(found), floor))
    for ratio, module, dos_ea, pc98_ea, la, lb, blocks, odd in found:
        flag = "  ← 有不可能出現的指令 %s" % sorted(set(odd)) if odd else ""
        print("  %-12s dos %05X ↔ pc98 %05X  %d/%d 條  比例 %.3f  差異區塊 %d%s"
              % (module, dos_ea, pc98_ea, la, lb, ratio, len(blocks), flag))
    return 0


if __name__ == "__main__":
    sys.exit(main())
