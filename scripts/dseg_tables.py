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

    # ---- 建角三表（spec 1099）----------------------------------------
    # spec 1084／884：+74h 的種族編號是 1..7，0 不是合法值。
    races = [(1, "矮人"), (2, "精靈"), (3, "地精"), (4, "半精靈"),
             (5, "半身人"), (6, "半獸人"), (7, "人類")]
    lines += ["## `DS:3F86h` 種族屬性上下限（每列 `10h` bytes，索引 ＝ 種族）", "",
              "版面：力量下限（男／女）、力量上限（男／女）、百分比上限（男／女），",
              "其後五個屬性各一組（下限, 上限）。索引算式見 spec 1086。", "",
              "| 種族 | 力量（男） | 力量（女） | 智力 | 睿智 | 敏捷 | 體質 | 魅力 |",
              "|---|---|---|---|---|---|---|---|"]
    for index, name in races:
        row = blob[0x3F86 + index * 0x10:0x3F86 + index * 0x10 + 0x10]
        lines.append("| %d %s | %d–%d/%d | %d–%d/%d | %d–%d | %d–%d | %d–%d | %d–%d | %d–%d |" % (
            index, name, row[0], row[2], row[4], row[1], row[3], row[5],
            row[6], row[7], row[8], row[9], row[10], row[11],
            row[12], row[13], row[14], row[15]))
    lines += ["",
              "★ 七列每一格都與 AD&D 1e 對得上：矮人體質 12–19／魅力 ≤16、",
              "精靈敏捷 ≤19／魅力 ≥8、半身人力量 ≤17 且百分比 0、半獸人體質 13–19／",
              "魅力 ≤12、人類全部 3–18 且百分比上限 100。",
              "⚠ 種族編號從 **1** 起算（spec 1084／884），`3F86h` 本身（索引 0）",
              "屬於前一張表。", ""]

    lines += ["## `DS:4172h` 職業組合的六屬性最低要求（每筆 6 bytes，17 筆 ＝ 0..10h）", "",
              "欄位順序與 spec 1086 的 `+11h + i×2` 一致：力、智、睿、敏、體、魅。",
              "`0` 代表該屬性無要求。", "",
              "| 組合 | 力 | 智 | 睿 | 敏 | 體 | 魅 |", "|---|---|---|---|---|---|---|"]
    for index in range(17):
        row = blob[0x4172 + index * 6:0x4172 + index * 6 + 6]
        lines.append("| %d | %s |" % (index, " | ".join(str(b) for b in row)))
    lines += ["",
              "★ 索引 3 ＝ `12 9 13 0 9 17` 正是 AD&D 1e 聖騎士（魅力 17 是唯一指紋）；",
              "索引 4 ＝ 遊俠；索引 16（`10h`）＝ 智 9 ＋ 敏 9 ＝ 法師／盜賊。",
              "多職組合取各單職要求的較大值。第 18 筆（`41D8h`）是遞增序列，不屬於本表。", ""]

    lines += ["## `DS:404Ch` 起始年齡（種族 × `1Ch` ＋ 職業 × 4，每筆 `基礎 word ＋ 骰數 ＋ 骰面`）", "",
              "| 種族 | 欄 0 | 欄 1 | 欄 2 | 欄 3 | 欄 4 | 欄 5 | 欄 6 |",
              "|---|---|---|---|---|---|---|---|"]
    for index, name in races:
        cells = []
        for column in range(7):
            offset = 0x404C + index * 0x1C + column * 4
            entry = blob[offset:offset + 4]
            base = entry[0] | (entry[1] << 8)
            cells.append("%d+%dd%d" % (base, entry[2], entry[3]) if entry[2] else "—")
        lines.append("| %d %s | %s |" % (index, name, " | ".join(cells)))
    lines += ["",
              "★ 欄 2 是戰士（矮人 40+5d4、精靈 130+5d6、地精 60+5d4、半精靈 22+3d4、",
              "半身人 20+3d4、半獸人 13+1d4、人類 15+1d4），欄 5 是法師",
              "（精靈 150+5d6），欄 6 是盜賊（半身人 40+2d4、人類 18+1d4）",
              "——全部與 AD&D 1e 相符。",
              "⚠ 種族編號從 **1** 起算；`404Ch` 本身（索引 0）的 28 bytes 逐位元組",
              "等於 `3FF8h` 可選職業表的索引 6＋7，兩張表相鄰。", ""]

    lines += ["## `DS:3FF8h` 種族可選職業組合（每列 `0Eh` bytes，索引 ＝ 種族）", "",
              "索引算式來自 spec 1093：`角色^[75h] := byte[DS:3FF8h ＋ 種族 × 0Eh ＋ 選單索引]`",
              "——把「這個種族的第 n 個選項」翻成職業組合編號。", "",
              "| 種族 | 位移 | 數量 | 職業組合編號 | 原始 14 bytes |",
              "|---|---|---|---|---|"]
    for index, name in races:
        offset = 0x3FF8 + index * 0x0E
        row = blob[offset:offset + 0x0E]
        cells = " ".join("%02X" % b for b in row)
        lines.append("| %d %s | `%04Xh` | %d | `%s` | `%s` |" % (
            index, name, offset, row[0],
            " ".join("%02X" % b for b in row[1:1 + row[0]]), cells))
    lines += ["",
              "★ 矮人／地精／半身人三列完全相同（`03 02 06 0E`）——與 spec 884",
              "「種族 1、3、5 三列完全相同」互相印證。",
              "★★★ **每列第一個位元組是可選職業的數量**，其後才是職業組合編號",
              "⇒ spec 1093 的選單索引從 1 起算。人類 6 個 ＝ 牧師／戰士／法師／盜賊／",
              "聖騎士／遊俠（只有人類能當聖騎士與遊俠）；半精靈 13 個最多；",
              "矮人／地精／半身人各 3 個（戰士／盜賊／戰士-盜賊）——全部符合 AD&D 1e。", ""]

    print("金錢名稱 %d 筆；法術 %d 筆；建角三表已排出" % (len(money), len(spells)))
    if not write:
        print("（預覽模式；加 --write 才寫檔）")
        return 0
    os.makedirs(os.path.dirname(OUT), exist_ok=True)
    open(OUT, "w", encoding="utf-8").write("\n".join(lines))
    print("已寫入 %s" % OUT)
    return 0


if __name__ == "__main__":
    sys.exit(main())
