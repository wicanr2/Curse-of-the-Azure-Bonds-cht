"""列出某模組還沒判定的函式，依指令數由少到多。

清單來源是 `docs/audit/coab-function-index.json`——**台帳統計用的同一份全集**。
先前用 prologue 匯出自己數過一次，`PC98-GAME.EXE` 得到 148 支未分類，實際只有
59 支：prologue 掃描對 resident 會把一支切成好幾段，數字因此虛高。要對齊進度就
只能用索引這一份，不要另外數。

用法：
    python3 scripts/worklist.py [platform] [module] [n]
    python3 scripts/worklist.py            # 各模組待解讀數排行
"""

import collections
import json
import os
import sys

ROOT = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
INDEX = os.path.join(ROOT, "docs", "audit", "coab-function-index.json")


def main():
    rows = [r for r in json.load(open(INDEX, encoding="utf-8"))["functions"]
            if r.get("state") == "待解讀"]
    if len(sys.argv) < 3:
        counter = collections.Counter((r["platform"], r["module"]) for r in rows)
        print("待解讀合計 %d" % len(rows))
        for (platform, module), count in counter.most_common(30):
            print("  %-5s %-16s %d" % (platform, module, count))
        return 0

    platform, module = sys.argv[1], sys.argv[2]
    limit = int(sys.argv[3]) if len(sys.argv) > 3 else 20
    picked = [r for r in rows
              if r["platform"] == platform and r["module"] == module]
    picked.sort(key=lambda r: r.get("instructions") or 0)
    print("%s %s 待解讀 %d，最小的 %d 支：" % (platform, module, len(picked),
                                              min(limit, len(picked))))
    for row in picked[:limit]:
        print("  %05Xh  指令 %-4d bytes %-5s entry=%s  %s"
              % (row["ea"], row.get("instructions") or 0, row.get("size"),
                 "Y" if row.get("is_overlay_entry") else "-",
                 row.get("ida_name") or ""))
    return 0


if __name__ == "__main__":
    sys.exit(main())
