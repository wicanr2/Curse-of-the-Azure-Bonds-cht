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
import re
import subprocess
import sys

sys.path.insert(0, os.path.join(os.path.dirname(os.path.abspath(__file__)), "..", "workplace"))
from pngtool import pixels  # noqa: E402

EGA = [(0, 0, 0), (0, 0, 170), (0, 170, 0), (0, 170, 170), (170, 0, 0), (170, 0, 170),
       (170, 85, 0), (170, 170, 170), (85, 85, 85), (85, 85, 255), (85, 255, 85),
       (85, 255, 255), (255, 85, 85), (255, 85, 255), (255, 255, 85), (255, 255, 255)]
FACING = {"N": 0, "E": 2, "S": 4, "W": 6}
ROOT = os.path.join(os.path.dirname(os.path.abspath(__file__)), "..")
SAVE_DIR = "workplace/dos-oracle/game/SAVE"


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
    per_map = {}
    failed = []
    for name, x, y, direction in rows:
        if direction not in FACING:
            print("跳過 %s（讀不到座標）" % name)
            continue
        # 檔名前綴就記著這張畫面是哪一張圖：`geo{檔集}-b{段}-…`。
        # ⚠ 兩邊要餵**同一份存檔**：remake 這一側改用 `-savgam-import`，
        # 而不是走 `-tilverton-dungeon` 的故事流程。故事流程只到得了提爾佛頓，
        # 而且它與原版的輸入不是同一個東西——同一份存檔才是對等的比較。
        # 兩種前綴：`geo{檔集}-b{段}-…`（ECL 段與幾何區塊同號），
        # 以及 `geo{檔集}-e{ECL段}-b{幾何區塊}-…`（兩者不同時）。
        # ⚠ 後者是必要的：GEO5 的 ECL 段 `0x31` 與 `0x32` **共用**幾何區塊 `0x32`，
        # 牆磚選圖卻不同。只記一個號碼的話這兩張圖分不開。
        match = re.match(r"geo(\d+)-e([0-9a-fA-F]{2})-b([0-9a-fA-F]{2})-", name)
        if match:
            area = match.group(1)
            block = str(int(match.group(2), 16))
            geoBlock = str(int(match.group(3), 16))
        else:
            match = re.match(r"geo(\d+)-b([0-9a-fA-F]{2})-", name)
            if not match:
                print("跳過 %s（檔名裡沒有地圖前綴）" % name)
                continue
            area = match.group(1)
            block = geoBlock = str(int(match.group(2), 16))
        shot = os.path.join(ROOT, "workplace",
                            "fp-remake-%s-%s-%s-%s-%s-%s.png" % (area, block, geoBlock, x, y, direction))
        author = ["tools/go.sh", "run", "./cmd/dos-save-export",
                  "-base", "workplace/orig-savgamb.dat",
                  "-out", SAVE_DIR, "-slot", "A", "-area", area,
                  "-ecl-block", block, "-map-block", geoBlock,
                  "-x", x, "-y", y, "-facing", str(FACING[direction])]
        # ⚠ 單一張失敗**不能中止整批**。五百多張跑到一半掛掉的話，前面那些張的
        # 數字就沒有小計可看，而失敗往往是暫時的（兩個 go run 搶 Xvfb）。
        # 失敗的張數要**逐張報出來並計入合計**，不能靜靜跳過——靜靜跳過會讓
        # 「合計差 N 格」看起來比實際好。
        try:
            subprocess.run(author, cwd=ROOT, stdout=subprocess.DEVNULL,
                           stderr=subprocess.DEVNULL, check=True)
            cmd = ["tools/go.sh", "run", "./cmd/azure-bonds-game",
                   "-savgam-dir", SAVE_DIR, "-savgam-slot", "A", "-savgam-import",
                   "-first-person",
                   "-screenshot", os.path.relpath(shot, ROOT)] + extra
            subprocess.run(cmd, cwd=ROOT, stdout=subprocess.DEVNULL,
                           stderr=subprocess.DEVNULL, check=True)
        except subprocess.CalledProcessError as failure:
            failed.append(name)
            print("%-34s (%s,%s) %s 產不出畫面（%s）" % (name, x, y, direction, failure), flush=True)
            continue
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
        # ⚠ 檔名一定要印出來：多張圖一起比的時候，只印座標的話**分不出哪一列是
        # 哪一張圖**，而每張圖的座標又會重複 ⇒ 數字沒辦法歸因到地圖。
        # ⚠ `flush` 不能省：輸出導到檔案時 Python 會整段緩衝，五百多張的跑批
        # 在結束前**看不到任何一列**，讀起來像卡住了。
        print("%-34s (%s,%s) %s 差 %5d %s" % (name, x, y, direction, diff, detail), flush=True)
        per_map[name.rsplit("-x", 1)[0]] = per_map.get(name.rsplit("-x", 1)[0], [0, 0, 0])
        bucket = per_map[name.rsplit("-x", 1)[0]]
        bucket[0] += diff
        bucket[1] += 1
        if diff == 0:
            bucket[2] += 1
    for prefix in sorted(per_map):
        cells, frames, exact = per_map[prefix]
        print("小計 %-30s 逐格相同 %2d／%2d，差 %6d 格" % (prefix, exact, frames, cells))
    if failed:
        print("⚠ 有 %d 張產不出畫面，沒有計入上面的數字：%s" % (len(failed), "、".join(failed)))
    print("合計不同格數 %d（比了 %d／%d 張）" % (total, len(rows) - len(failed), len(rows)))
    return 0 if total == 0 and not failed else 1


if __name__ == "__main__":
    raise SystemExit(main())
