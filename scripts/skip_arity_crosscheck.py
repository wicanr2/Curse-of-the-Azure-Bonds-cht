"""用 `ECL2` 的 SKIP 常式驗證 arity 表。

`SKIP`（PC-98 `overlay-07:1FB0h`，由 opcode `16h`~`1Bh` 呼叫）要跳過下一條
ECL 指令，就必須知道那條指令有幾個 operand。它裡面因此有一張**完整的
opcode → arity 對照**，而且是用 `cmp al, N` 逐值寫死的。

這是**獨立於 `READVAR` 參數的第二個來源**：先前的 arity 表來自「每個 handler
自己呼叫 `READVAR(n)` 的 n」，這張表來自「別人要跳過它時認為它有多長」。
兩者對不上就代表其中一個讀錯了。

解析判準：`cmp al, N` / `jz BODY` 收集到同一個 `mov al, arity` 為止；
`cmp al,4 / jb / cmp al,7 / jbe` 這種範圍比較也要處理。

用法：python3 scripts/skip_arity_crosscheck.py
"""

import json
import os
import re
import sys

ROOT = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
from checkfx_timing_table import instructions, immediate

AUDIT = os.path.join(ROOT, "docs", "audit", "ecl-handler-operand-audit.md")


def parse_skip(platform, module, start, end):
    """回傳 {opcode: arity}。"""
    by_ea, items = instructions(platform, module)
    rows = [it for it in items if start <= it["ea"] < end]
    table, pending, low = {}, [], None
    for index, item in enumerate(rows):
        text = re.sub(r"\s*;.*$", "", item["disasm"].strip())
        if text.startswith("cmp") and " al," in text:
            value = immediate(text)
            nxt = re.sub(r"\s*;.*$", "", rows[index + 1]["disasm"].strip())
            # `cmp al,4 / jb X / cmp al,7 / jbe Y` 是一段範圍比較（04h..07h）。
            # `jbe` 要先判，否則 startswith("jb") 會把它一起吃掉，04h~07h 就
            # 整段消失——表面上看起來像「SKIP 表沒有這幾個 opcode」。
            if nxt.startswith("jbe"):
                if low is not None:
                    pending.extend(range(low, value + 1))
                    low = None
                else:
                    pending.append(value)
            elif nxt.startswith("jb"):
                low = value
            else:
                pending.append(value)
        elif text.startswith("mov") and re.match(r"mov\s+al,\s*[0-9A-Fa-f]", text):
            arity = immediate(text)
            for opcode in pending:
                table[opcode] = arity
            pending = []
    return table


def audit_arity():
    out = {}
    for line in open(AUDIT, encoding="utf-8"):
        match = re.match(r"\|\s*`([0-9A-F]+)h`\s*\|[^|]*\|\s*(\d+)\s*\|", line)
        if match:
            out[int(match.group(1), 16)] = int(match.group(2))
    return out


def main():
    skip = parse_skip("pc98", "overlay-07", 0x1FB0, 0x20C5)
    audit = audit_arity()
    print("SKIP 表解出 %d 個 opcode；既有 arity 表 %d 個" % (len(skip), len(audit)))

    agree = disagree = only_skip = only_audit = 0
    problems = []
    for opcode in sorted(set(skip) | set(audit)):
        a, b = skip.get(opcode), audit.get(opcode)
        if a is None:
            only_audit += 1
            if b:                      # arity 0 不會出現在 SKIP 表裡（走 inc PC）
                problems.append((opcode, "SKIP 沒有", b))
        elif b is None:
            only_skip += 1
            problems.append((opcode, a, "arity 表沒有"))
        elif a == b:
            agree += 1
        else:
            disagree += 1
            problems.append((opcode, a, b))

    print("兩邊都有且一致：%d；不一致：%d" % (agree, disagree))
    print("只在 SKIP 表：%d；只在 arity 表：%d" % (only_skip, only_audit))
    if problems:
        print("\nopcode  SKIP  arity 表")
        for opcode, a, b in problems:
            print("  %02Xh   %-5s %s" % (opcode, a, b))
    return 0


if __name__ == "__main__":
    sys.exit(main())
