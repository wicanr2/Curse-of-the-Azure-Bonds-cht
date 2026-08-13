"""盤點內容相同、位址不同的字串常數。

起因是 `overlay-22:407Ah` 與 `578Bh`：兩支邏輯逐指令相同的治療 opcode，各自
指向 `CS:4070h` 與 `CS:5781h`，而兩個位址上的內容都是 `is Healed`。編譯器沒有
合併相同的字串常數，所以**同一句話在同一個 overlay 裡可能有好幾份**。

對中文化的意義：只改其中一份，遊戲會在某些路徑顯示中文、某些路徑顯示英文，
而觸發條件通常和法術／分支綁在一起，很難重現。用「字串去重之後的清單」翻譯
會系統性地漏掉這些。

輸入是 `scripts/scan_pascal_strings.py` 的結果。那份掃描本身是**下界**
（只認得出有 `mov di, offset` 引用或落在連續字串表裡的），所以本報表列出的
重複組數同樣是下界——沒列到不代表沒有重複。

用法：
    python3 scripts/duplicate_strings.py
"""

import collections
import json
import os

ROOT = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
SRC = os.path.join(ROOT, "docs", "audit", "embedded-strings.json")
OUT = os.path.join(ROOT, "docs", "audit", "duplicate-strings.md")
MIN_LEN = 4


def main():
    modules = json.load(open(SRC, encoding="utf-8"))["modules"]
    rows = [(platform, module, item["offset"], item["text"])
            for platform, per in modules.items()
            for module, items in per.items()
            for item in items]

    within = collections.defaultdict(list)
    across = collections.defaultdict(set)
    for platform, module, offset, text in rows:
        if len(text.strip()) < MIN_LEN:
            continue
        within[(platform, module, text)].append(offset)
        across[(platform, text)].add(module)

    repeated = {k: sorted(v) for k, v in within.items() if len(v) > 1}
    spread = {k: sorted(v) for k, v in across.items() if len(v) > 1}

    lines = ["# 內容相同、位址不同的字串常數", "",
             "由 `scripts/duplicate_strings.py` 產生。編譯器沒有合併相同的字串",
             "常數，所以同一句話在同一個 overlay 裡可能有好幾份。**中文化必須每一份",
             "都改**——只改一份會讓遊戲在某些路徑顯示中文、某些路徑顯示英文，而觸發",
             "條件通常和法術或分支綁在一起，很難重現。", "",
             "來源掃描是下界（見 `scan_pascal_strings.py`），所以本表也是下界：",
             "**沒列到不代表沒有重複**。", "",
             "## 同一模組內重複（%d 組）" % len(repeated), "",
             "| 平台 | 模組 | 份數 | 位址 | 內容 |", "|---|---|---:|---|---|"]
    for (platform, module, text), offsets in sorted(
            repeated.items(), key=lambda kv: (-len(kv[1]), kv[0])):
        lines.append("| %s | %s | %d | %s | %s |"
                     % (platform, module, len(offsets),
                        " ".join("`%04Xh`" % o for o in offsets),
                        text.replace("|", "\\|")))

    lines += ["", "## 跨模組重複（%d 句）" % len(spread), "",
              "同一句話出現在多個 overlay。這類不會互相影響，但翻譯要一致，",
              "否則同一句在不同場景會有兩種譯法。", "",
              "| 平台 | 模組數 | 模組 | 內容 |", "|---|---:|---|---|"]
    for (platform, text), mods in sorted(
            spread.items(), key=lambda kv: (-len(kv[1]), kv[0])):
        lines.append("| %s | %d | %s | %s |"
                     % (platform, len(mods), " ".join(sorted(mods)),
                        text.replace("|", "\\|")))

    open(OUT, "w", encoding="utf-8").write("\n".join(lines) + "\n")
    print("模組內重複 %d 組、跨模組重複 %d 句 → %s"
          % (len(repeated), len(spread), OUT))
    total = sum(len(v) - 1 for v in repeated.values())
    print("模組內多出來的份數合計 %d（這就是「只改一份」會漏掉的數量）" % total)
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
