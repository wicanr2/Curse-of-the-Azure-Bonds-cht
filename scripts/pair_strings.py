"""用已配對函式內的引用順序，把 DOS 英文與 PC-98 日文字串一一對上。

為什麼不能按模組內的出現順序對：英文有單複數分歧
（`takes 1 point of damage` 與 ` points of damage `），日文只有一句，
所以兩平台的**條數不同**，按序號對會整段錯位。

可靠的對法是走已配對的函式：`cross_platform_pair.py` 的配對保證兩支函式的
**助憶碼序列完全相同**，於是 `mov di, imm16` 出現的位置與次數也相同——
第 k 個引用必然是同一件事。只在「兩邊第 k 個引用都指向字串」時才輸出。

因此本表的每一列都有兩層依據：函式配對（結構同構）＋ 引用序位（同一條指令）。
沒配上的字串不會出現在這裡，**這是下界不是全集**。

用法：
    python3 scripts/pair_strings.py [--write]
"""

import json
import os
import re
import sys

ROOT = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
from cross_platform_pair import full, unique_index

STRINGS = os.path.join(ROOT, "docs", "audit", "embedded-strings.json")
OUT_MD = os.path.join(ROOT, "docs", "audit", "string-pairs.md")
OUT_JSON = os.path.join(ROOT, "docs", "audit", "string-pairs.json")


def di_immediates(function):
    """body 中所有 `mov di, imm16` 的立即值，依出現順序。"""
    out = []
    for item in function["items"]:
        raw = bytes.fromhex(item["bytes"])
        if len(raw) == 3 and raw[0] == 0xBF:
            out.append(raw[1] | (raw[2] << 8))
    return out


def main():
    write = "--write" in sys.argv
    strings = json.load(open(STRINGS, encoding="utf-8"))["modules"]
    lookup = {(platform, module, item["offset"]): item["text"]
              for platform, modules in strings.items()
              for module, items in modules.items() for item in items}

    pairs, mismatched = [], 0
    for index in range(36):
        module = "overlay-%02d" % index
        dos_index, pc98_index = unique_index(full("dos", module)), unique_index(full("pc98", module))
        for sequence in set(dos_index) & set(pc98_index):
            dos_function, pc98_function = dos_index[sequence], pc98_index[sequence]
            dos_refs, pc98_refs = di_immediates(dos_function), di_immediates(pc98_function)
            if len(dos_refs) != len(pc98_refs):
                mismatched += 1          # 助憶碼相同就不該發生；發生就跳過並記數
                continue
            for position, (a, b) in enumerate(zip(dos_refs, pc98_refs)):
                english = lookup.get(("dos", module, a))
                japanese = lookup.get(("pc98", module, b))
                if english is None or japanese is None:
                    continue
                pairs.append({"module": module, "dos_function": dos_function["ea"],
                              "pc98_function": pc98_function["ea"], "position": position,
                              "dos_offset": a, "pc98_offset": b,
                              "english": english, "japanese": japanese})

    pairs.sort(key=lambda p: (p["module"], p["dos_function"], p["position"]))
    seen, unique_pairs = set(), []
    for pair in pairs:
        key = (pair["module"], pair["dos_offset"], pair["pc98_offset"])
        if key in seen:
            continue
        seen.add(key)
        unique_pairs.append(pair)

    print("對上 %d 條（去重後 %d 條）；引用數不一致而跳過的函式對：%d"
          % (len(pairs), len(unique_pairs), mismatched))
    if not write:
        for pair in unique_pairs[:15]:
            print("  %-11s %-32s %s" % (pair["module"], pair["english"], pair["japanese"]))
        print("（預覽模式；加 --write 才寫報表）")
        return 0

    json.dump({"schema": "coab-string-pairs/1", "pairs": unique_pairs},
              open(OUT_JSON, "w", encoding="utf-8"), ensure_ascii=False, indent=1)
    lines = ["# 英日字串對照（由函式配對 ＋ 引用序位推得）", "",
             "每一列的依據有兩層：兩支函式的助憶碼序列完全相同（結構同構），",
             "以及該字串是兩邊 body 裡**第幾個** `mov di, offset` 的目標（同一條指令）。",
             "不是按模組內的出現順序對——英文有單複數分歧、日文沒有，條數不同，",
             "按序號對會整段錯位。",
             "",
             "沒配上的字串不會出現在這裡，**這是下界不是全集**。", "",
             "| 模組 | DOS | PC-98 | 英文 | 日文 |", "|---|---|---|---|---|"]
    for pair in unique_pairs:
        lines.append("| %s | `%04Xh` | `%04Xh` | %s | %s |"
                     % (pair["module"], pair["dos_offset"], pair["pc98_offset"],
                        pair["english"].replace("|", "\\|"),
                        pair["japanese"].replace("|", "\\|")))
    open(OUT_MD, "w", encoding="utf-8").write("\n".join(lines) + "\n")
    print("→ %s" % OUT_MD)
    return 0


if __name__ == "__main__":
    sys.exit(main())
