"""從常駐資料段的原始 dump 抽出已定位的表格，產出可查的參考文件。

資料來源是 `tools/ida/dump_data_segments.py` 倒出來的 `dseg` 原始位元組。
`DS:xxxx` 直接就是這份檔案的位移（段基底對齊在段首），所以先前判讀裡寫的
每一個 `DS:` 位址都可以拿來當索引。

**每張表的位置與筆距都不是猜的**，來源標在下面的 `TABLES` 裡：那是先前逐條讀
函式時，從索引算式反推出來的（例如「`shl di, 4` 之後加 `37DAh`」就是 16 bytes
一筆、基底 `37DAh`）。本工具只負責把位元組排出來，不新增任何語意宣稱。

用法：
    python3 scripts/dseg_tables.py [--write]
"""

import json
import os
import sys

ROOT = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
DSEG = os.path.join(ROOT, "workplace", "re-sweep", "dos", "dseg",
                    "dos-dseg-dseg.bin")
OUT = os.path.join(ROOT, "docs", "audit", "resident-data-tables.md")


def pascal(blob, offset, limit=40):
    length = blob[offset]
    if not 0 < length <= limit:
        return None
    payload = blob[offset + 1:offset + 1 + length]
    if not all(32 <= b < 127 for b in payload):
        return None
    return payload.decode("cp437")


def strings(blob, base, stride, first, count, limit=40):
    """基底 ＋ 索引 × 筆距的 Pascal 短字串表；讀到不合法就停。"""
    out = []
    index = first
    while count is None or len(out) < count:
        offset = base + index * stride
        if offset + stride > len(blob):
            break
        text = pascal(blob, offset, limit)
        if text is None:
            break
        out.append((index, offset, text))
        index += 1
    return out


def signed(value):
    return value - 256 if value >= 128 else value


def main():
    write = "--write" in sys.argv
    blob = open(DSEG, "rb").read()
    lines = ["# 常駐資料段裡的表格（DOS）", "",
             "由 `scripts/dseg_tables.py` 從 `dos-dseg-dseg.bin` 抽出。",
             "`DS:xxxx` 就是這份 dump 的位移。每張表的基底與筆距來自先前逐條讀",
             "函式時反推的索引算式，本文件只排位元組、不新增語意宣稱。", ""]

    money = strings(blob, 0x0F93, 0x0B, 0, 7)
    lines += ["## `DS:0F93h` 金錢欄位名稱（7 筆，每筆 `0Bh` bytes，0 起算）",
              "", "| 索引 | 位移 | 名稱 |", "|---|---|---|"]
    lines += ["| %d | `%04Xh` | %s |" % row for row in money]
    lines.append("")

    dx = [signed(blob[0x2694 + i]) for i in range(9)]
    dy = [signed(blob[0x269D + i]) for i in range(9)]
    lines += ["## `DS:2694h` / `DS:269Dh` 方向位移（各 9 bytes）", "",
              "| 方向 | dx | dy |", "|---|---|---|"]
    lines += ["| %d | %d | %d |" % (i, dx[i], dy[i]) for i in range(9)]
    lines += ["", "索引 8 是 `(0, 0)`——原地。0 起順時針一圈。", ""]

    shape = [blob[0x27D8 + i] for i in range(13)]
    lines += ["## `DS:27D8h` 雲霧形狀（13 bytes，值是上表的方向索引）", "",
              "- 4 格版（`byte[27D7h + k]`，k = 1..4）：`%s`" % shape[0:4],
              "- 9 格版（`byte[27DBh + k]`，k = 1..9）：`%s`" % shape[4:13], ""]

    spells = strings(blob, 0x27BD, 0x29, 1, None)
    lines += ["## `DS:27BDh` 法術名稱（每筆 `29h` bytes，**1 起算**，共 %d 筆）"
              % len(spells), "", "| 編號 | 位移 | 名稱 |", "|---|---|---|"]
    lines += ["| %d | `%04Xh` | %s |" % row for row in spells]
    lines.append("")

    lines += ["## `DS:37DAh` 法術屬性（每筆 16 bytes，索引與法術名稱同）", "",
              "| 編號 | 名稱 | " + " | ".join("+%X" % i for i in range(16)) + " |",
              "|---|---|" + "---|" * 16]
    for index, _, name in spells:
        offset = 0x37DA + index * 16
        cells = " | ".join("%02X" % b for b in blob[offset:offset + 16])
        lines.append("| %d | %s | %s |" % (index, name, cells))
    lines.append("")

    print("金錢名稱 %d 筆；法術 %d 筆" % (len(money), len(spells)))
    if not write:
        print("（預覽模式；加 --write 才寫檔）")
        return 0
    os.makedirs(os.path.dirname(OUT), exist_ok=True)
    open(OUT, "w", encoding="utf-8").write("\n".join(lines))
    print("已寫入 %s" % OUT)
    return 0


if __name__ == "__main__":
    sys.exit(main())
