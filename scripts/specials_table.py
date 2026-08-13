"""解析 CHECKSPECIALS／STORESPECIALS 的比對鏈，把 ECL 特殊位址對回紀錄欄位。

兩支都是一長串 `cmp ax, N` ＋ 分支，body 裡讀寫某個結構欄位。與 `CHECKFX`
（第 617 輪）同一種形狀，所以判準也一樣：**用比對次數對帳**——鏈上有幾個
`cmp ax, imm`，就必須解析出幾筆；對不上就是解析漏了，不是表就這麼大。

用法：
    python3 scripts/specials_table.py <ea16> [--module overlay-07]
"""

import json
import os
import re
import sys

ROOT = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))


def load(module, ea):
    """走 prologue 匯出並依位址範圍取指令——與 `show.py` 同一條路。

    不能只取「IDA 認為屬於這支的 items」：這種長比對鏈會被 IDA 切成好幾段，
    只取第一段就會少掉大半張表，而且**少得很像表本來就這麼大**。
    """
    sweep = os.path.join(ROOT, "workplace", "re-sweep", "pc98")
    path = os.path.join(sweep, "overlays", "prologue", "pc98-%s.json" % module)
    functions = json.load(open(path, encoding="utf-8"))["functions"]
    blob = open(os.path.join(sweep, "overlays", module + ".bin"), "rb").read()
    hits = [blob.find(pattern, ea + 3) for pattern in (b"\x55\x89\xe5", b"\x55\x8b\xec")]
    end = min([h for h in hits if h > 0] or [len(blob)])
    seen = {}
    for function in functions:
        for item in function["items"]:
            if ea <= item["ea"] < end:
                seen.setdefault(item["ea"], item)
    return [seen[k] for k in sorted(seen)]


def main():
    ea = int(sys.argv[1], 16)
    module = "overlay-07"
    if "--module" in sys.argv:
        module = sys.argv[sys.argv.index("--module") + 1]
    items = load(module, ea)

    rows, pending, body, compares = [], None, [], 0
    for item in items:
        text = re.sub(r"\s*;.*$", "", item["disasm"].strip())
        match = re.match(r"cmp\s+ax,\s*([0-9A-F]+)h?$", text)
        if match:
            if pending is not None:
                rows.append((pending, body))
            pending, body = int(match.group(1).rstrip("h"), 16), []
            compares += 1
            continue
        if pending is not None:
            body.append(text)
    if pending is not None:
        rows.append((pending, body))

    print("cmp 次數 = %d，解析出 %d 筆" % (compares, len(rows)))
    print()
    for value, lines in rows:
        base = fields = ""
        for line in lines:
            got = re.search(r"les\s+di,\s*ds:([0-9A-F]+)h", line)
            if got and not base:
                base = "DS:%sh" % got.group(1)
            got = re.search(r"(?:mov|add|sub)\s+(?:al|ax|dx),\s*es:\[di\+?([0-9A-F]*)h?\]", line)
            if got:
                fields += ("+%sh " % got.group(1)) if got.group(1) else "+0 "
            got = re.search(r"mov\s+es:\[di\+?([0-9A-F]*)h?\],", line)
            if got:
                fields += ("→+%sh " % got.group(1)) if got.group(1) else "→+0 "
        width = "byte" if any("al," in l for l in lines) else "word"
        sign = "cbw" if any(l == "cbw" for l in lines) else (
            "zero" if any("xor ah, ah" in l for l in lines) else "")
        print("%04Xh (7C00h+%04Xh)  %-10s %-14s %s %s"
              % (0x7C00 + value, value, base, fields.strip(), width, sign))
    return 0


if __name__ == "__main__":
    sys.exit(main())
