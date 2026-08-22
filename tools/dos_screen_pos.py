#!/usr/bin/env python3
"""從原版擷取畫面讀出「x y 朝向」。

⚠ **不是每一張圖都會顯示座標。** 提爾佛頓與下水道的狀態列是 `7,13 N 00:00`，
而 GEO3 段 0x15 那類只有 `N 00:00`——沒有座標。讀不到座標時**朝向仍然讀得到**，
所以分成兩種回報，讓呼叫端自己決定要不要接受：

    7   13  N       座標與朝向都讀到
    ?   ?   N       只有朝向（這張圖不顯示座標）
    ?   ?   ?       什麼都沒讀到（畫面不對）

把後兩種混成同一種，會讓「這張圖不顯示座標」與「畫面根本不對」看起來一樣。
"""
import os
import re
import sys

sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
from dos_screen import read_screen  # noqa: E402

text = " ".join(read_screen(sys.argv[1]))
full = re.search(r"(\d+),(\d+) ([NESW])", text)
if full:
    print("\t".join(full.groups()))
else:
    # 朝向後面一定跟著時間，拿它把單獨的 N/E/S/W 與其他字母分開。
    facing = re.search(r"\b([NESW]) \d+:\d+", text)
    print("?\t?\t" + (facing.group(1) if facing else "?"))
