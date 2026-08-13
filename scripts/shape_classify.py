"""依「指令組成」把待解讀函式分型，找出可以整支機械化讀完的那些。

`CHECKFX` 的 2,287 bytes 之所以能一次讀完，是因為它的指令組成極其單純：
除了 24 個 `cmp al, N` 與 161 組「`mov al, N` ＋ `call`」之外，**沒有任何
算術、沒有任何記憶體存取**。這種函式不必逐條讀——解析出表格內容，再用
「`mov` 數 ＝ `call` 數 ＝ 表格項數」的相等關係自證涵蓋完整。

分型判準只看助憶碼與運算元形狀，不看語意：

- `dispatch`：只有 `cmp <reg>, imm` 分派 ＋ `mov <reg>, imm` ＋ `call`，
  加上 `push`／跳躍／prologue。**沒有算術、沒有記憶體存取。**
- `forward`：沒有 `cmp`，只是把參數原樣 `push` 再 `call`（一到數個）。
- `arith`：其餘——含算術、記憶體存取、迴圈，必須逐條讀。

`dispatch` 與 `forward` 是可以批次處理的；`arith` 沒有捷徑。

用法：python3 scripts/shape_classify.py [--list <shape>]
"""

import collections
import glob
import json
import os
import re
import sys

ROOT = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
SWEEP = os.path.join(ROOT, "workplace", "re-sweep")
LEDGER = os.path.join(ROOT, "docs", "audit", "re-function-ledger.json")

FRAME = {"push bp", "mov bp, sp", "mov sp, bp", "pop bp"}
CONTROL = {"jmp", "jz", "jnz", "ja", "jb", "jbe", "jae", "jl", "jg", "jle", "jge",
           "jnb", "retf", "retn", "ret"}
ARITH = {"add", "sub", "adc", "sbb", "mul", "imul", "div", "idiv", "shl", "shr",
         "sar", "and", "or", "not", "neg", "inc", "dec", "cbw", "cwd", "xchg",
         "les", "lea", "movsw", "movsb", "cmpsb", "stosw", "loop", "rep", "repe",
         "int", "test", "sal", "rcl", "rcr", "rol", "ror", "cli", "sti", "out", "in"}


def normalise(text):
    return re.sub(r"\s*;.*$", "", text.strip())


def classify(function):
    kinds = collections.Counter()
    memory = False
    for item in function["items"]:
        text = normalise(item["disasm"])
        if text in FRAME:
            continue
        head = text.split()[0]
        kinds[head] += 1
        if head in ARITH:
            return "arith", kinds
        if "[" in text and head in ("mov", "cmp"):
            memory = True
    if memory:
        return "arith", kinds
    if kinds.get("call", 0) == 0:
        return "arith", kinds
    if kinds.get("cmp", 0) > 0:
        return "dispatch", kinds
    return "forward", kinds


def modules():
    for platform in ("dos", "pc98"):
        for path in sorted(glob.glob(os.path.join(
                SWEEP, platform, "overlays", "full", "%s-overlay-*.json" % platform))):
            yield platform, os.path.basename(path)[len(platform) + 1:-5], path


def main():
    want = sys.argv[sys.argv.index("--list") + 1] if "--list" in sys.argv else None
    ledger = json.load(open(LEDGER, encoding="utf-8"))
    done = {(e["platform"], e["module"], e["ea"]) for e in ledger["functions"]
            if e["state"] != "待解讀"}

    counts = collections.Counter()
    rows = []
    for platform, module, path in modules():
        for function in json.load(open(path, encoding="utf-8"))["functions"]:
            if (platform, module, function["ea"]) in done:
                continue
            shape, kinds = classify(function)
            counts[shape] += 1
            counts[shape + "_bytes"] += function["size"]
            if shape == want:
                rows.append((platform, module, function["ea"], function["size"],
                             kinds.get("cmp", 0), kinds.get("call", 0)))

    print("待解讀函式的分型：")
    for shape in ("dispatch", "forward", "arith"):
        print("  %-9s %5d 支 %8d bytes" % (shape, counts[shape], counts[shape + "_bytes"]))
    if want:
        rows.sort(key=lambda r: -r[3])
        print("\n%s（前 30 大）：" % want)
        print("  平台  模組         位址    size   cmp  call")
        for row in rows[:30]:
            print("  %-5s %-12s %04Xh %6d %5d %5d" % row)
    return 0


if __name__ == "__main__":
    sys.exit(main())
