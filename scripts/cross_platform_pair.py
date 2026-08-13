"""用助憶碼序列，把 DOS 與 PC-98 同一 overlay 的函式一對一配對。

兩平台是同一份 Turbo Pascal 原始碼編譯的，同一支程序的**指令種類與順序**
會一致；不一致的只有運算元（DS 位址、overlay-local 位址、立即值都可能不同）。
所以比對只看 mnemonic，不看運算元。

配對判準（三條同時成立，刻意保守）：

1. 助憶碼序列完全相同。
2. 序列長度 `>= MIN_LEN`。太短的序列（`push bp / mov bp,sp / pop bp / retf`）
   在一個 overlay 裡會出現幾十次，配對沒有鑑別力。
3. **序列在各自平台的該 overlay 內唯一**。有兩支以上同形就整組放棄——
   寧可少配，不可配錯。

配上之後轉移的等級是 `strong inference`：結構同構是強證據，但**運算元裡的
位址必須各自確認**。

用法：
    python3 scripts/cross_platform_pair.py            # 量測配對率
    python3 scripts/cross_platform_pair.py --write    # 把已讀側的判讀轉移過去
"""

import collections
import json
import os
import re
import sys

ROOT = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
SWEEP = os.path.join(ROOT, "workplace", "re-sweep")
LEDGER = os.path.join(ROOT, "docs", "audit", "re-function-ledger.json")
MIN_LEN = 8


def full(platform, module):
    if module.endswith(".EXE"):
        path = os.path.join(SWEEP, "dos" if platform == "dos" else "pc98",
                            "full", "%s-%s.json" % (platform, module))
        if platform == "pc98":
            path = os.path.join(SWEEP, "PC98-GAME.EXE.i64")  # 佔位，下面改
    name = "%s-%s" % (platform, module)
    for base in (os.path.join(SWEEP, platform, "overlays", "full"),
                 os.path.join(SWEEP, platform, "full"),
                 os.path.join(SWEEP, "dos", "full"),
                 os.path.join(SWEEP, "pc98", "full")):
        candidate = os.path.join(base, name + ".json")
        if os.path.exists(candidate):
            return json.load(open(candidate, encoding="utf-8"))["functions"]
    return []


def mnemonics(function):
    out = []
    for item in function["items"]:
        text = re.sub(r"\s*;.*$", "", item["disasm"].strip())
        out.append(text.split()[0].lower() if text else "")
    return tuple(out)


def unique_index(functions):
    """回傳「序列 → 函式」，只保留該平台內唯一且夠長的序列。"""
    buckets = collections.defaultdict(list)
    for function in functions:
        sequence = mnemonics(function)
        if len(sequence) >= MIN_LEN:
            buckets[sequence].append(function)
    return {sequence: group[0] for sequence, group in buckets.items()
            if len(group) == 1}


def main():
    write = "--write" in sys.argv
    ledger = json.load(open(LEDGER, encoding="utf-8"))
    state = {(e["platform"], e["module"], e["ea"]): e for e in ledger["functions"]}

    modules = ["overlay-%02d" % i for i in range(36)] + ["START.EXE", "PC98-GAME.EXE"]
    pairs, added, total_pairable = 0, [], 0
    per_module = []
    for module in modules:
        dos_module = "START.EXE" if module == "PC98-GAME.EXE" else module
        pc98_module = "PC98-GAME.EXE" if module == "START.EXE" else module
        if module == "PC98-GAME.EXE":
            continue
        dos_functions = full("dos", dos_module)
        pc98_functions = full("pc98", pc98_module)
        if not dos_functions or not pc98_functions:
            continue
        dos_index, pc98_index = unique_index(dos_functions), unique_index(pc98_functions)
        common = set(dos_index) & set(pc98_index)
        pairs += len(common)
        module_new = 0
        for sequence in common:
            a = ("dos", dos_module, dos_index[sequence]["ea"])
            b = ("pc98", pc98_module, pc98_index[sequence]["ea"])
            sa, sb = state.get(a), state.get(b)
            read_a = sa is not None and sa["state"] == "已解讀"
            read_b = sb is not None and sb["state"] == "已解讀"
            if read_a == read_b:
                continue
            source, target = (sa, b) if read_a else (sb, a)
            other = "%s %s:%04Xh" % (source["platform"], source["module"], source["ea"])
            added.append({
                "platform": target[0], "module": target[1], "ea": target[2],
                "state": "已解讀", "level": "strong inference", "spec": source["spec"],
                "note": "與 %s 助憶碼序列完全相同（%d 條指令，且該序列在兩邊各自的"
                        "模組內唯一），語意同該筆：%s ⚠ 運算元中的 DS／overlay-local "
                        "位址兩平台不同，引用位址前須各自確認"
                        % (other, len(sequence), source["note"]),
            })
            module_new += 1
        if common:
            per_module.append((module, len(common), module_new))
        total_pairable += min(len(dos_index), len(pc98_index))

    print("可配對函式對：%d（兩邊各自唯一且 >= %d 條指令）" % (pairs, MIN_LEN))
    print("其中一邊已讀、可立刻轉移：%d" % len(added))
    print("\n模組            配對對數  可轉移")
    for module, count, new in per_module:
        print("  %-14s %6d  %6d" % (module, count, new))

    if not write:
        print("\n（預覽模式；加 --write 才寫入台帳）")
        return 0

    keys = {(e["platform"], e["module"], e["ea"]) for e in added}
    ledger["functions"] = [e for e in ledger["functions"]
                           if (e["platform"], e["module"], e["ea"]) not in keys] + added
    json.dump(ledger, open(LEDGER, "w", encoding="utf-8"), ensure_ascii=False, indent=1)
    print("\n已寫入 %s" % LEDGER)
    return 0


if __name__ == "__main__":
    sys.exit(main())
