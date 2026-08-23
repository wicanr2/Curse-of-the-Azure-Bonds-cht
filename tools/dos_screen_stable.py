#!/usr/bin/env python3
"""兩張原版畫面的**第一人稱可視區**是不是一模一樣。

★ 為什麼只比可視區。 整張畫面永遠不會兩次完全相同——時間、游標之類的東西一直
在動。而我們要判的是「第一人稱那一塊畫完了沒有」：文字先到、圖形後到，
`gate` 只等到文字出現就回，這時可視區可能還是**上一格**的畫面。

可視區是原生 (24,24) 起 88×88，擷取是兩倍（spec 406）。

用法：tools/dos_screen_stable.py a.png b.png    → 相同回 0，不同回 1
"""
import sys

ORIGIN = 48
SIZE = 88


def main():
    sys.path.insert(0, "workplace")
    from pngtool import pixels as read_png

    _, _, a = read_png(sys.argv[1])
    _, _, b = read_png(sys.argv[2])
    for y in range(ORIGIN, ORIGIN + SIZE * 2):
        for x in range(ORIGIN, ORIGIN + SIZE * 2):
            if tuple(a[y][x][:3]) != tuple(b[y][x][:3]):
                return 1
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
