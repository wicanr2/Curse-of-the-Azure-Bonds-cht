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
    races = ["0（見下方註）", "矮人", "精靈", "地精", "半精靈", "半身人"]
    lines += ["## `DS:3F86h` 種族屬性上下限（每列 `10h` bytes，索引 ＝ 種族）", "",
              "版面：力量下限（男／女）、力量上限（男／女）、百分比上限（男／女），",
              "其後五個屬性各一組（下限, 上限）。索引算式見 spec 1086。", "",
              "| 種族 | 力量（男） | 力量（女） | 智力 | 睿智 | 敏捷 | 體質 | 魅力 |",
              "|---|---|---|---|---|---|---|---|"]
    for index, name in enumerate(races):
        row = blob[0x3F86 + index * 0x10:0x3F86 + index * 0x10 + 0x10]
        lines.append("| %s | %d–%d/%d | %d–%d/%d | %d–%d | %d–%d | %d–%d | %d–%d | %d–%d |" % (
            name, row[0], row[2], row[4], row[1], row[3], row[5],
            row[6], row[7], row[8], row[9], row[10], row[11],
            row[12], row[13], row[14], row[15]))
    lines += ["",
              "⚠ **索引 0 那列不是屬性限制**（原始位元組 `%s`）。"
              % " ".join("%02X" % b for b in blob[0x3F86:0x3F96]),
              "種族 1..5 的每一格都與 AD&D 1e 對得上（矮人體質 12–19／魅力 ≤16、",
              "精靈敏捷 ≤19／魅力 ≥8、半身人力量 ≤17 且百分比 0），索引 0 沒有。", ""]

    lines += ["## `DS:4172h` 職業組合的六屬性最低要求（每筆 6 bytes，13 筆）", "",
              "欄位順序與 spec 1086 的 `+11h + i×2` 一致：力、智、睿、敏、體、魅。",
              "`0` 代表該屬性無要求。", "",
              "| 組合 | 力 | 智 | 睿 | 敏 | 體 | 魅 |", "|---|---|---|---|---|---|---|"]
    for index in range(13):
        row = blob[0x4172 + index * 6:0x4172 + index * 6 + 6]
        lines.append("| %d | %s |" % (index, " | ".join(str(b) for b in row)))
    lines += ["",
              "★ 索引 3 ＝ `12 9 13 0 9 17` 正是 AD&D 1e 聖騎士（魅力 17 是唯一指紋）；",
              "索引 4 ＝ `13 13 14 0 14 0` 是遊俠。多職組合取各單職要求的較大值。", ""]

    lines += ["## `DS:404Ch` 起始年齡（種族 × `1Ch` ＋ 職業 × 4，每筆 `基礎 word ＋ 骰數 ＋ 骰面`）", "",
              "| 種族 | 欄 0 | 欄 1 | 欄 2 | 欄 3 | 欄 4 | 欄 5 | 欄 6 |",
              "|---|---|---|---|---|---|---|---|"]
    for index, name in enumerate(races):
        cells = []
        for column in range(7):
            offset = 0x404C + index * 0x1C + column * 4
            entry = blob[offset:offset + 4]
            base = entry[0] | (entry[1] << 8)
            cells.append("%d+%dd%d" % (base, entry[2], entry[3]) if entry[2] else "—")
        lines.append("| %s | %s |" % (name, " | ".join(cells)))
    human = []
    for column in range(7):
        entry = blob[0x4110 + column * 4:0x4110 + column * 4 + 4]
        base = entry[0] | (entry[1] << 8)
        human.append("%d+%dd%d" % (base, entry[2], entry[3]) if entry[2] else "—")
    lines += ["| **人類（`DS:4110h`，見下）** | %s |" % " | ".join(human), "",
              "★ 欄 2 是戰士（矮人 40+5d4、精靈 130+5d6、地精 60+5d4、半精靈 22+3d4、",
              "半身人 20+3d4），欄 5 是法師（精靈 150+5d6），欄 6 是盜賊",
              "（半身人 40+2d4）——全部與 AD&D 1e 相符。",
              "⚠ **索引 0 那列不是年齡資料**；人類的年齡在 `DS:4110h`",
              "（欄 2 ＝ 15+1d4 ＝ AD&D 1e 人類戰士）。本表不宣稱它是同一張表的第 8 列",
              "還是獨立的一張表。", ""]

    lines += ["## `DS:3FF8h` 種族可選職業組合（每列 `0Eh` bytes，索引 ＝ 種族）", "",
              "索引算式來自 spec 1093：`角色^[75h] := byte[DS:3FF8h ＋ 種族 × 0Eh ＋ 選單索引]`",
              "——把「這個種族的第 n 個選項」翻成職業組合編號。", "",
              "| 種族 | 位移 | 14 個位元組 |", "|---|---|---|"]
    for index, name in enumerate(races):
        offset = 0x3FF8 + index * 0x0E
        cells = " ".join("%02X" % b for b in blob[offset:offset + 0x0E])
        lines.append("| %s | `%04Xh` | `%s` |" % (name, offset, cells))
    lines += ["",
              "★ 矮人／地精／半身人三列完全相同（`03 02 06 0E`）；精靈 8 個；",
              "半精靈 14 個全滿。⚠ 半精靈那列的 `0D` 出現兩次且沒有 `00` 結尾，",
              "所以 `00` 是「組合編號 0」還是結束標記**無法從資料本身判定**；",
              "實際可選數量由選單有幾項決定，而項數不在這張表裡。",
              "⚠ 索引 0 那列（`12 12 64 32 …`）的值超出職業組合編號的值域 `00h`..`10h`,",
              "與 `3F86h`／`404Ch` 一樣不是該表的資料。", ""]

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
