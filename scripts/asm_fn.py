"""從 IDA 的 `.asm` 全文列表裡取出一支函式的完整內容。

常駐執行檔（`START.EXE`／`PC98-GAME.EXE`）的指令級 JSON dump 只涵蓋一部分
函式，但 `workplace/re-sweep/<平台>/<檔名>.asm` 是完整的。本工具直接從那份
列表切出 `proc`..`endp` 之間的內容，順便把註解欄與位址欄留著（IDA 的 xref
註解常常就是答案）。

用法：
    python3 scripts/asm_fn.py <平台> <ea hex> [<ea hex> …]
    python3 scripts/asm_fn.py <平台> --name sub_1A410
"""

import os
import re
import sys

ROOT = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
ASM = {
    "dos": os.path.join(ROOT, "workplace", "re-sweep", "dos", "START.EXE.asm"),
    "pc98": os.path.join(ROOT, "workplace", "re-sweep", "pc98",
                         "PC98-GAME.EXE.asm"),
}


def cut(lines, name):
    start = None
    for i, line in enumerate(lines):
        if re.match(r"^%s\s+proc\b" % re.escape(name), line):
            start = i
            break
    if start is None:
        return None
    out = []
    for line in lines[start:]:
        out.append(line.rstrip())
        if re.match(r"^%s\s+endp\b" % re.escape(name), line):
            break
    return out


def main():
    plat = sys.argv[1]
    lines = open(ASM[plat], encoding="latin-1").read().splitlines()
    if sys.argv[2] == "--name":
        names = sys.argv[3:]
    else:
        names = ["sub_%X" % int(a, 16) for a in sys.argv[2:]]
    for name in names:
        body = cut(lines, name)
        if body is None:
            print("=== %s：在 .asm 裡找不到" % name)
            continue
        print("=== %s（%d 行）" % (name, len(body)))
        for line in body:
            print(line)


main()
