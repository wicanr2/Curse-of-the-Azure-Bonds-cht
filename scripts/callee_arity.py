"""從每支函式結尾的 `retf N` 建立參數個數表。

起因是 spec 711／712：本來從呼叫端「連續推了幾個 push」去推參數個數，把
`overlay-23 entry#21` 推成 7 個 word（實際 10）、`sub_E11h` 推成 4 個（實際 1）。
錯誤會傳染——推導鏈上只要有一支的個數猜錯，後面全歪，而且兩個呼叫點若犯同一個
假設，錯的結果看起來還會互相印證。

被呼叫者結尾的 `retf N` 是編譯器寫進去的事實：Turbo Pascal 由被呼叫者清堆疊，
`N` 就是參數佔的位元組數。所以

    參數 word 數 = N / 2

**這是起點；呼叫端的推入數量只能拿來交叉檢查。**

不適用的情況本工具會如實標出，不猜：
- `retf`（沒有立即數）→ 無參數。
- `retn` → 近呼叫，多半是同模組內被 `push cs ; call near` 當 far 用的程序，
  這種一樣以 `retf` 收尾才會被算進來；純 `retn` 標為 `near`。
- 結尾不是 ret 的（IDA 邊界錯或尾跳躍）→ 標為 `未知`。

用法：
    python3 scripts/callee_arity.py                    # 產生 JSON ＋ 統計
    python3 scripts/callee_arity.py dos overlay-23     # 只看一個模組
"""

import collections
import glob
import json
import os
import re
import sys

ROOT = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
SWEEP = os.path.join(ROOT, "workplace", "re-sweep")
OUT = os.path.join(ROOT, "docs", "audit", "callee-arity.json")
RET = re.compile(r"^ret([nf]?)\s*([0-9A-Fa-f]+h?)?$")


def classify(function):
    items = function["items"]
    if not items:
        return None, "空"
    text = re.sub(r"\s+", " ", items[-1]["disasm"].strip())
    match = RET.match(text)
    if not match:
        return None, "未知"
    kind, immediate = match.group(1), match.group(2)
    if immediate is None:
        return 0, "far" if kind == "f" else ("near" if kind == "n" else "ret")
    value = int(immediate.rstrip("hH"), 16)
    if value % 2:
        return None, "奇數位元組"
    return value // 2, "far" if kind == "f" else "near"


def main():
    only = sys.argv[1:3] if len(sys.argv) >= 3 else None
    result = {}
    stats = collections.Counter()
    for platform in ("dos", "pc98"):
        for path in sorted(glob.glob(os.path.join(
                SWEEP, platform, "overlays", "prologue", "*.json"))):
            module = os.path.basename(path)[len(platform) + 1:-5]
            if only and (platform, module) != tuple(only):
                continue
            for function in json.load(open(path, encoding="utf-8"))["functions"]:
                words, kind = classify(function)
                key = "%s/%s/%04X" % (platform, module, function["ea"])
                result[key] = {"words": words, "kind": kind}
                stats[kind if words is None else "%s %d word" % (kind, words)] += 1
                if only:
                    print("  %04Xh  %-6s %s" % (function["ea"], kind,
                                                "—" if words is None else "%d word" % words))
    if only:
        return 0
    json.dump({"schema": "coab-callee-arity/1", "functions": result},
              open(OUT, "w", encoding="utf-8"), ensure_ascii=False, indent=1)
    print("共 %d 支 → %s" % (len(result), OUT))
    for kind, count in stats.most_common(12):
        print("  %-14s %d" % (kind, count))
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
