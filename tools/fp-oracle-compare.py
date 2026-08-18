#!/usr/bin/env python3
"""把原版第一人稱擷取與 remake 的同格畫面逐格比對（EGA 量化之後）。

索引檔每列是 `檔名 x y 朝向`（`#` 開頭是註解），檔案與索引同一個目錄。
比對範圍是 spec 406 的 88×88 可見區，兩邊都從畫面 (48,48) 起算——
原版擷取是 320×200 的兩倍，remake 的舞台內容也落在同一個矩形。

比的是 EGA 量化之後的 16 色索引，不是 RGB：DOSBox 與 remake 對同一個 EGA
顏色算出來的 RGB 差幾個階（168 vs 173），直接比 RGB 會有九成的格子「不同」，
把真正的差異埋掉。

用法：
    tools/fp-oracle-compare.py docs/reference/original-dos/first-person/index.tsv
"""
import os
import subprocess
import sys

sys.path.insert(0, os.path.join(os.path.dirname(os.path.abspath(__file__)), "..", "workplace"))
from pngtool import pixels  # noqa: E402

EGA = [(0, 0, 0), (0, 0, 170), (0, 170, 0), (0, 170, 170), (170, 0, 0), (170, 0, 170),
       (170, 85, 0), (170, 170, 170), (85, 85, 85), (85, 85, 255), (85, 255, 85),
       (85, 255, 255), (255, 85, 85), (255, 85, 255), (255, 255, 85), (255, 255, 255)]
FACING = {"N": 0, "E": 2, "S": 4, "W": 6}
ROOT = os.path.join(os.path.dirname(os.path.abspath(__file__)), "..")


def quantize(pixel):
    return min(range(16), key=lambda i: sum((pixel[c] - EGA[i][c]) ** 2 for c in range(3)))


def viewport(path):
    _, _, px = pixels(path)
    return [[quantize(tuple(px[48 + y * 2][48 + x * 2][:3])) for x in range(88)] for y in range(88)]


def main():
    index = sys.argv[1]
    extra = sys.argv[2:]
    folder = os.path.dirname(os.path.abspath(index))
    rows = [line.split("\t") for line in open(index).read().strip().split("\n")
            if line.strip() and not line.startswith("#")]
    total = 0
    for name, x, y, direction in rows:
        if direction not in FACING:
            print("跳過 %s（讀不到座標）" % name)
            continue
        shot = os.path.join(ROOT, "workplace", "fp-remake-%s-%s-%s.png" % (x, y, direction))
        cmd = ["tools/go.sh", "run", "./cmd/azure-bonds-game", "-tilverton-dungeon",
               "-dungeon-x", x, "-dungeon-y", y,
               "-dungeon-facing", str(FACING[direction]),
               "-screenshot", os.path.relpath(shot, ROOT)] + extra
        subprocess.run(cmd, cwd=ROOT, stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL, check=True)
        a = viewport(os.path.join(folder, name))
        b = viewport(shot)
        classes = {}
        for row in range(88):
            for col in range(88):
                if a[row][col] != b[row][col]:
                    key = (a[row][col], b[row][col])
                    classes[key] = classes.get(key, 0) + 1
        diff = sum(classes.values())
        total += diff
        detail = " ".join("%d→%d×%d" % (k[0], k[1], v) for k, v in
                          sorted(classes.items(), key=lambda kv: -kv[1])[:4])
        print("(%s,%s) %s 差 %4d %s" % (x, y, direction, diff, detail))
    print("合計不同格數 %d" % total)
    return 0 if total == 0 else 1


if __name__ == "__main__":
    raise SystemExit(main())
