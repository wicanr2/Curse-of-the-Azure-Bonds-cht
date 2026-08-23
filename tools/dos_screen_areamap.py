#!/usr/bin/env python3
"""從原版的區域地圖（`AREA`）畫面核對隊伍站在哪一格。

★ 為什麼需要。 `tools/dos-oracle-jump-capture.sh` 用自製存檔把原版放到指定的
格子，然後從畫面把座標讀回來核對——存檔寫進去不等於原版照著站。但**有些地圖
畫面上不顯示座標**（`geo5-b33` 那一類只顯示朝向與時間），第一人稱那一格只核對
得到朝向。**這一支就是那些圖的位置核對**。

★★ 區域地圖上有隊伍標記（一個朝向箭頭），而地圖是 16×16 的地圖開一個 11×11 的
視窗，視窗原點是 `clamp(隊伍 − 5, 0, 5)`。所以「預測標記該在哪一個字元格、
再看那一格是不是箭頭」就是一個不靠畫面文字的位置驗證。

⚠ 判準刻意是**預測後核對**，不是「找出箭頭再反推座標」：反推需要先知道視窗
原點，而視窗原點又要先知道座標。預測只需要目標座標，而目標座標正是我們要驗的
那一個假設。

用法：
    tools/dos_screen_areamap.py <區域地圖畫面.png> <x> <y>
印 `ok` 或 `mismatch`，離開碼 0／1。
"""
import sys

CHAR = 16          # 字元格 16×16 像素
MAP_ROW = 3        # 地圖內部左上角的字元格
MAP_COL = 3
WINDOW = 11        # 視窗 11×11
MAP_SIZE = 16      # 地圖 16×16

# 朝北的隊伍標記（從 `geo5-b33` (1,4) 的區域地圖取下來，16×16 的箭頭）。
ARROW_NORTH = [
    "########..######",
    "########..######",
    "######..##..####",
    "######..##..####",
    "####..######..##",
    "####..######..##",
    "##..##########..",
    "##..##########..",
    "##....######....",
    "##....######....",
    "####..######..##",
    "####..######..##",
    "####..######..##",
    "####..######..##",
    "####..........##",
    "####..........##",
]


def glyph(pixels, row, column):
    """把一個字元格轉成「亮／暗」的點陣圖。

    ⚠ 判「不是黑色」而不是判某個顏色：區域地圖的配色跟著地圖宣告走，
    寫死顏色會在換圖時安靜地失效。
    """
    out = []
    for y in range(CHAR):
        line = ""
        for x in range(CHAR):
            pixel = pixels[row * CHAR + y][column * CHAR + x][:3]
            line += "." if tuple(pixel) == (0, 0, 0) else "#"
        out.append(line)
    return out


def expected_cell(x, y):
    """預測隊伍標記會落在哪一個字元格。"""
    origin_x = min(max(x - WINDOW // 2, 0), MAP_SIZE - WINDOW)
    origin_y = min(max(y - WINDOW // 2, 0), MAP_SIZE - WINDOW)
    return MAP_ROW + y - origin_y, MAP_COL + x - origin_x


def main():
    sys.path.insert(0, "workplace")
    from pngtool import pixels as read_png

    path, x, y = sys.argv[1], int(sys.argv[2]), int(sys.argv[3])
    _, _, pixels = read_png(path)
    row, column = expected_cell(x, y)
    if not (0 <= row < len(pixels) // CHAR and 0 <= column < len(pixels[0]) // CHAR):
        print("mismatch (預測的字元格在畫面外)")
        return 1
    if glyph(pixels, row, column) == ARROW_NORTH:
        print("ok")
        return 0
    print("mismatch (字元格 %d,%d 不是朝北的隊伍標記)" % (row, column))
    return 1


if __name__ == "__main__":
    raise SystemExit(main())
