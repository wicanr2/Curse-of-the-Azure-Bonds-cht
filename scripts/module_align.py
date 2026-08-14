"""用序列比對重新對齊兩平台的函式清單，不再假設 entry index 一一對應。

背景：`pair_dump.py` / `entry_dump.py` / `one_sided_triage.py` 都以「同一個
entry index 就是同一支函式」為前提。實測**這個前提會壞**——`overlay-22` 的
PC-98 側有 117 支、DOS 側只有 116 支，從 index 12 之後整組錯開一格，於是
`entry_dump` 會把 PC-98 的「暫時改欄位再還原」配到 DOS 的「常數字串」上。

作法：把每支函式壓成一個助憶碼序列，再用 `difflib.SequenceMatcher` 對**整個
模組的函式清單**做一次對齊（相等的判準是助憶碼序列的相似度夠高）。對齊之後
才有資格說「這兩支是同一支」。

⚠ 本工具只給對應關係與相似度，**不做任何語意判定**。要標已解讀一律要把
差異逐條看過。

用法：
    python3 scripts/module_align.py                 # 掃全部模組，只印錯開的
    python3 scripts/module_align.py <模組>          # 印該模組的完整對齊表
    python3 scripts/module_align.py <模組> --pending # 只印至少一邊待解讀的
"""

import difflib
import json
import os
import re
import sys

ROOT = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
SWEEP = os.path.join(ROOT, "workplace", "re-sweep")
LEDGER = os.path.join(ROOT, "docs", "audit", "re-function-ledger.json")
RET = ("retf", "retn", "ret")


def clean(item):
    return re.sub(r"\s+", " ", re.sub(r"\s*;.*$", "", item["disasm"].strip()))


def shape(function):
    items = function["items"]
    index = [i for i, x in enumerate(items) if clean(x).split(" ")[0] in RET]
    body = items[:index[-1] + 1] if index else items
    return [clean(x).split(" ")[0] for x in body]


def load(platform, module):
    for sub in ("filled", "prologue"):
        path = os.path.join(SWEEP, platform, "overlays", sub,
                            "%s-%s.json" % (platform, module))
        if os.path.exists(path):
            data = json.load(open(path, encoding="utf-8"))
            functions = data["functions"] if isinstance(data, dict) else data
            return [f for f in functions if f["items"]]
    return []


class Key(object):
    """把函式包成可雜湊、以『助憶碼序列夠像』為相等判準的物件。"""

    __slots__ = ("shape", "bucket")

    def __init__(self, function):
        self.shape = shape(function)
        # 用長度分桶讓 __hash__ 合法（相等的兩者必須同 hash）。
        self.bucket = len(self.shape) // 8

    def __hash__(self):
        return self.bucket

    def __eq__(self, other):
        if abs(len(self.shape) - len(other.shape)) > 8:
            return False
        return difflib.SequenceMatcher(
            None, self.shape, other.shape).quick_ratio() >= 0.75


def align(module):
    left, right = load("dos", module), load("pc98", module)
    if not left or not right:
        return None
    matcher = difflib.SequenceMatcher(None, [Key(f) for f in left],
                                      [Key(f) for f in right])
    pairs = []
    for op, i1, i2, j1, j2 in matcher.get_opcodes():
        if op == "equal":
            for k in range(i2 - i1):
                pairs.append((i1 + k, j1 + k))
        elif op == "replace" and (i2 - i1) == (j2 - j1):
            for k in range(i2 - i1):
                pairs.append((i1 + k, j1 + k))
        else:
            for k in range(i1, i2):
                pairs.append((k, None))
            for k in range(j1, j2):
                pairs.append((None, k))
    return left, right, pairs


def main():
    ledger = json.load(open(LEDGER, encoding="utf-8"))["functions"]
    state = {(e["platform"], e["module"], e["ea"]): e["state"] for e in ledger}
    modules = ["overlay-%02d" % n for n in range(36)]
    only = None
    pending_only = False
    for arg in sys.argv[1:]:
        if arg == "--pending":
            pending_only = True
        else:
            only = arg
    if only:
        modules = [only]

    for module in modules:
        got = align(module)
        if not got:
            continue
        left, right, pairs = got
        shifted = [p for p in pairs if p[0] is not None and p[1] is not None
                   and p[0] != p[1]]
        orphan = [p for p in pairs if None in p]
        if not only:
            if shifted or orphan:
                print("%s：dos %d 支／pc98 %d 支，錯開 %d 對、單邊 %d 支"
                      % (module, len(left), len(right), len(shifted),
                         len(orphan)))
            continue

        print("%s：dos %d 支／pc98 %d 支" % (module, len(left), len(right)))
        for i, j in pairs:
            dos_ea = left[i]["items"][0]["ea"] if i is not None else None
            p98_ea = right[j]["items"][0]["ea"] if j is not None else None
            dos_st = state.get(("dos", module, dos_ea), "待解讀") if i is not None else "—"
            p98_st = state.get(("pc98", module, p98_ea), "待解讀") if j is not None else "—"
            if pending_only and "待解讀" not in (dos_st, p98_st):
                continue
            ratio = ""
            if i is not None and j is not None:
                ratio = " %.3f" % difflib.SequenceMatcher(
                    None, shape(left[i]), shape(right[j])).ratio()
            print("  dos %s %-6s ↔ pc98 %s %-6s%s"
                  % ("%05X" % dos_ea if i is not None else "  —  ", dos_st,
                     "%05X" % p98_ea if j is not None else "  —  ", p98_st,
                     ratio))


if __name__ == "__main__":
    main()
