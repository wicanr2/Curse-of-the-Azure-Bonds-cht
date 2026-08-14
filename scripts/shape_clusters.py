"""在**同一個平台、同一個模組內**找出助憶碼序列完全相同的待解讀函式群組。

同一段原始碼被複製貼上、只改幾個常數的情形在這份程式碼裡很常見（spec 748 的
雲霧解除是 268 條指令只差 13 個運算元；spec 754 的 `overlay-28` 三張查表也是
同一個模子）。這種群組可以**讀一支、把其餘每一支的運算元差異逐條列出來**——
只要差異窮舉完整，每一支都算實際讀完，不是外推。

判準：助憶碼序列（**只取到最後一條 `ret`**，理由同 spec 761）完全相同，且長度
`>= 12` 條。太短的序列（`push bp / mov bp,sp / … / retf`）會把一堆語意無關的
小函式湊在一起，反而沒有價值。

輸出按「群組大小 × 指令數」排序——先做省最多力氣的。

用法：
    python3 scripts/shape_clusters.py [最少幾支一組]
"""

import collections
import glob
import json
import os
import re
import sys

ROOT = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
SWEEP = os.path.join(ROOT, "workplace", "re-sweep")
INDEX = os.path.join(ROOT, "docs", "audit", "coab-function-index.json")
RET = ("retf", "retn", "ret")
MIN_LENGTH = 12


def body(function):
    names = []
    for item in function["items"]:
        text = re.sub(r"\s*;.*$", "", item["disasm"].strip())
        names.append(text.split()[0] if text else "")
    tail = [i for i, name in enumerate(names) if name in RET]
    return tuple(names[:tail[-1] + 1]) if tail else tuple(names)


def load(platform, module):
    for folder in ("filled", "prologue"):
        path = os.path.join(SWEEP, platform, "overlays", folder,
                            "%s-%s.json" % (platform, module))
        if os.path.exists(path):
            return {f["ea"]: f
                    for f in json.load(open(path, encoding="utf-8"))["functions"]}
    return {}


def main():
    minimum = int(sys.argv[1]) if len(sys.argv) > 1 else 3
    rows = [r for r in json.load(open(INDEX, encoding="utf-8"))["functions"]
            if r.get("state") == "待解讀"]

    cache, groups = {}, collections.defaultdict(list)
    for row in rows:
        key = (row["platform"], row["module"])
        if key not in cache:
            cache[key] = load(*key)
        function = cache[key].get(row["ea"])
        if not function or not function["items"]:
            continue
        shape = body(function)
        if len(shape) < MIN_LENGTH:
            continue
        groups[(row["platform"], row["module"], shape)].append(row["ea"])

    picked = [(key, eas) for key, eas in groups.items() if len(eas) >= minimum]
    picked.sort(key=lambda item: -len(item[0][2]) * len(item[1]))
    total = sum(len(eas) for _, eas in picked)
    print("群組 %d 個，涵蓋 %d 支待解讀（門檻：同模組同平台、>= %d 條、>= %d 支）"
          % (len(picked), total, MIN_LENGTH, minimum))
    for (platform, module, shape), eas in picked:
        print("  %-5s %-14s %3d 條 × %2d 支： %s"
              % (platform, module, len(shape), len(eas),
                 " ".join("%05X" % ea for ea in sorted(eas))))
    return 0


if __name__ == "__main__":
    sys.exit(main())
