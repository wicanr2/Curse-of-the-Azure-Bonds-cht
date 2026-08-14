"""找出「整支就是寫死版面」的函式：每一行都是 建字串 + 一個顯示呼叫。

依據 spec 916：建字串助手（DOS `0A54h:0634h`／PC-98 `0A65h:062Fh`）只彈掉
`@來源`，`@目的` 留在堆疊上當下一個呼叫的引數。所以「N 個 byte 參數 →
`@目的` → `@來源` → 建字串 → 顯示」是一個固定 16 條（N=4）的樣板。

本工具**只做形狀比對**，不做語意判定。它回報的是「這支函式有多少條指令
可以被樣板吃掉、剩幾條」。**剩 0 條才代表整支都是版面**，此時把版面表抄
出來就等於逐條讀完；剩下不是 0 的一律要人工看。

用法：
    python3 scripts/layout_scan.py                 # 掃全部待解讀函式
    python3 scripts/layout_scan.py <平台> <模組> <ea hex>   # 印單一支的版面
"""

import json
import os
import re
import sys

sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
from module_align import ROOT, LEDGER, clean, load

BUILD = ("0A54h:634h", "0A65h:62Fh")


def constants(platform, module):
    """把 overlay 的 Pascal 字串常數區掃成 {位移: bytes}。"""
    path = os.path.join(ROOT, "workplace", "re-sweep", platform,
                        "overlays", "%s.bin" % module)
    if not os.path.exists(path):
        return b""
    return open(path, "rb").read()


def pascal(blob, off):
    if off is None or off < 0 or off >= len(blob):
        return None
    n = blob[off]
    if off + 1 + n > len(blob):
        return None
    return blob[off + 1:off + 1 + n]


def parse(function, blob, encoding):
    """吃掉樣板，回傳 (版面列, 剩下沒被吃掉的指令)。"""
    items = [clean(x) for x in function["items"]]
    rows, rest = [], []
    args, src = [], None
    pending = []
    for m in items:
        pending.append(m)
        g = re.match(r"mov a[lx], ([0-9A-F]+)h?$", m)
        if g:
            args.append(int(g.group(1), 16))
            continue
        g = re.match(r"mov di, offset \w*?([0-9A-F]+)$", m)
        if g:
            src = int(g.group(1), 16)
            continue
        if m in ("push ss", "push cs", "push di", "push ax") or \
           re.match(r"lea di, \[bp[+-]\w+\]$", m):
            continue
        if m.startswith("call") and any(b in m for b in BUILD):
            continue
        if m.startswith("call") and src is not None and args:
            text = pascal(blob, src)
            rows.append((tuple(args), m, text))
            args, src, pending = [], None, []
            continue
        if m in ("push bp", "mov bp, sp", "mov sp, bp", "pop bp") or \
           m.startswith("sub sp,") or m.startswith("retf") or m.startswith("retn"):
            pending = []
            continue
        rest.extend(pending)
        args, src, pending = [], None, []
    rest.extend(pending)
    return rows, rest


def main():
    ledger = json.load(open(LEDGER, encoding="utf-8"))["functions"]
    done = {(r["platform"], r["module"], r["ea"]) for r in ledger
            if r["state"] != "待解讀"}
    if len(sys.argv) == 4:
        plat, mod, ea = sys.argv[1], sys.argv[2], int(sys.argv[3], 16)
        blob = constants(plat, mod)
        enc = "cp932" if plat == "pc98" else "cp437"
        for f in load(plat, mod):
            if f["ea"] != ea:
                continue
            rows, rest = parse(f, blob, enc)
            print("=== %s %s:%05Xh 共 %d 條，版面 %d 行，剩 %d 條"
                  % (plat, mod, ea, len(f["items"]), len(rows), len(rest)))
            for a, call, t in rows:
                s = t.decode(enc, "replace") if t else "?"
                print("  %-22s %-26s %r" % (list(a), call[10:], s))
            for m in rest:
                print("  剩 |", m)
        return

    import glob
    hits = []
    for plat in ("dos", "pc98"):
        pat = os.path.join(ROOT, "workplace", "re-sweep", plat,
                           "overlays", "*", "%s-*.json" % plat)
        mods = sorted({os.path.basename(p).split("-", 1)[1][:-5]
                       for p in glob.glob(pat)})
        for mod in mods:
            fns = load(plat, mod)
            if not fns:
                continue
            blob = constants(plat, mod)
            enc = "cp932" if plat == "pc98" else "cp437"
            for f in fns:
                if (plat, mod, f["ea"]) in done:
                    continue
                rows, rest = parse(f, blob, enc)
                if len(rows) >= 3 and len(rest) <= 200:
                    hits.append((len(rest), plat, mod, f["ea"],
                                 len(f["items"]), len(rows)))
    hits.sort()
    for rest, plat, mod, ea, n, rows in hits:
        print("剩%3d  %-5s %-11s %05Xh  %4d 條  版面 %d 行"
              % (rest, plat, mod, ea, n, rows))
    print("共", len(hits), "支")


main()
