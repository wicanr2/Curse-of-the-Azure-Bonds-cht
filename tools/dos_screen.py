#!/usr/bin/env python3
"""把原版 DOS 擷取畫面上的文字讀成字串（原版行為 oracle 的狀態判別）。

★ 存在的理由：定時送鍵的序列很脆——任何一步的載入時間漂移都會讓後面每一個鍵
錯位。要穩定驅動原版就得**看畫面決定下一鍵**，而看畫面的最小可用形式是把
底部狀態列與選單文字讀出來。

畫面是 320×200 的 8×8 文字格（擷取時放大兩倍成 640×400）。每一格的背景色取
「格內最常出現的顏色」，其餘像素當前景——反白（選取中的項目）因此自動正規化成
與正常顯示相同的位元圖，不必另外處理。

字型表放在 tools/dos_screen_font.json（sig → 字元），由已知文字的畫面 bootstrap，
認不出來的格子讀成 `?`。
"""

import json
import os
import sys
from collections import Counter

sys.path.insert(0, os.path.join(os.path.dirname(os.path.abspath(__file__)), "..", "workplace"))

FONT_PATH = os.path.join(os.path.dirname(os.path.abspath(__file__)), "dos_screen_font.json")
COLS, ROWS = 40, 25


def _pixels(path):
    from pngtool import pixels  # workplace/pngtool.py
    return pixels(path)


def cell_signature(px, col, row, scale=2):
    """回傳該格的 64-bit 位元圖（背景 0、前景 1）與前景顏色。"""
    vals = []
    for y in range(8):
        for x in range(8):
            p = px[(row * 8 + y) * scale][(col * 8 + x) * scale]
            vals.append(tuple(p[:3]))
    bg = Counter(vals).most_common(1)[0][0]
    bits = 0
    fg = None
    for i, v in enumerate(vals):
        if v != bg:
            bits |= 1 << i
            if fg is None:
                fg = v
    return bits, bg, fg


def load_font():
    if os.path.exists(FONT_PATH):
        with open(FONT_PATH, encoding="utf-8") as fh:
            return {int(k, 16): v for k, v in json.load(fh).items()}
    return {}


def save_font(table):
    with open(FONT_PATH, "w", encoding="utf-8") as fh:
        json.dump({"%016x" % k: v for k, v in sorted(table.items())}, fh,
                  indent=0, ensure_ascii=False, sort_keys=True)
        fh.write("\n")


def read_screen(path, table=None, scale=2):
    """回傳 25 行文字（每行 40 字）。"""
    table = load_font() if table is None else table
    _, _, px = _pixels(path)
    lines = []
    for row in range(ROWS):
        out = []
        for col in range(COLS):
            bits, _, _ = cell_signature(px, col, row, scale)
            out.append(" " if bits == 0 else table.get(bits, "?"))
        lines.append("".join(out).rstrip())
    return lines


def learn(path, row, col, text, table, scale=2):
    _, _, px = _pixels(path)
    for i, ch in enumerate(text):
        bits, _, _ = cell_signature(px, col + i, row, scale)
        if bits == 0:
            continue
        table[bits] = ch
    return table


def main():
    args = sys.argv[1:]
    if not args:
        print(__doc__)
        return 2
    if args[0] == "learn":
        table = load_font()
        path = args[1]
        for spec in args[2:]:
            row, col, text = spec.split(":", 2)
            learn(path, int(row), int(col), text, table)
        save_font(table)
        print("字型表 %d 筆" % len(table))
        return 0
    for path in args:
        for i, line in enumerate(read_screen(path)):
            if line.strip():
                print("%2d| %s" % (i, line))
        print("--", path)
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
