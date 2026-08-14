"""印出「以模組整體對齊後」的對應函式差異，取代 entry index 假設。

`entry_dump.py` 以 entry index 配對，但 13 個模組的兩平台函式數不同
（`overlay-22` 差 1 支、`overlay-16` 差 11 支），index 一錯就整組錯位。
本工具改用 `module_align.py` 的序列對齊結果找對應函式，再印差異。

用法：
    python3 scripts/aligned_diff.py <平台> <模組> <ea 十六進位> [--body]

只印差異塊與運算元差異；加 `--body` 才連本體一起印。
"""

import difflib
import json
import os
import re
import sys

sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
import module_align as align_mod

LEDGER = os.path.join(os.path.dirname(os.path.dirname(
    os.path.abspath(__file__))), "docs", "audit", "re-function-ledger.json")


def clean(item):
    return re.sub(r"\s+", " ", re.sub(r"\s*;.*$", "", item["disasm"].strip()))


def body(function):
    items = function["items"]
    index = [i for i, x in enumerate(items)
             if clean(x).split(" ")[0] in align_mod.RET]
    return items[:index[-1] + 1] if index else items


def main():
    if len(sys.argv) < 4:
        raise SystemExit(__doc__)
    platform, module, ea = sys.argv[1], sys.argv[2], int(sys.argv[3], 16)
    show_body = "--body" in sys.argv

    got = align_mod.align(module)
    if not got:
        raise SystemExit("找不到 %s 的匯出" % module)
    left, right, pairs = got
    mine_list = left if platform == "dos" else right
    peer_list = right if platform == "dos" else left
    peer_name = "pc98" if platform == "dos" else "dos"

    index = next((k for k, f in enumerate(mine_list)
                  if f["items"][0]["ea"] == ea), None)
    if index is None:
        raise SystemExit("%s %s 沒有 %05Xh 這支" % (platform, module, ea))

    partner = None
    for i, j in pairs:
        pick = i if platform == "dos" else j
        other = j if platform == "dos" else i
        if pick == index:
            partner = other
            break
    if partner is None:
        raise SystemExit("對齊後 %05Xh 沒有對應函式（單邊）" % ea)

    mine, peer = body(mine_list[index]), body(peer_list[partner])
    ledger = json.load(open(LEDGER, encoding="utf-8"))["functions"]
    state = {(e["platform"], e["module"], e["ea"]): (e["state"], e.get("spec"))
             for e in ledger}
    peer_ea = peer_list[partner]["items"][0]["ea"]
    peer_state = state.get((peer_name, module, peer_ea), ("待解讀", None))

    ms = [clean(x).split(" ")[0] for x in mine]
    ps = [clean(x).split(" ")[0] for x in peer]
    matcher = difflib.SequenceMatcher(None, ms, ps)

    print("=== %s %s:%05Xh（%d 條）↔ %s %05Xh（%d 條）相似度 %.3f"
          % (platform, module, ea, len(mine), peer_name, peer_ea, len(peer),
             matcher.ratio()))
    print("--- 對側台帳狀態：%s%s"
          % (peer_state[0], "／" + peer_state[1] if peer_state[1] else ""))

    if show_body:
        for item in mine:
            print("  %04X  %s" % (item["ea"], clean(item)))

    blocks = [op for op in matcher.get_opcodes() if op[0] != "equal"]
    if blocks:
        print("--- 助憶碼序列的差異區塊（%d 塊）" % len(blocks))
        for _, i1, i2, j1, j2 in blocks:
            print("    本邊 : %s" % ([clean(x) for x in mine[i1:i2]] or "（無）"))
            print("    對側 : %s" % ([clean(x) for x in peer[j1:j2]] or "（無）"))

    print("--- 對齊區段的運算元差異")
    count = 0
    for op, i1, i2, j1, j2 in matcher.get_opcodes():
        if op != "equal":
            continue
        for k in range(i2 - i1):
            a, b = clean(mine[i1 + k]), clean(peer[j1 + k])
            if a == b:
                continue
            if a.split(" ")[0] in ("jmp", "je", "jz", "jne", "jnz", "jb",
                                   "jnb", "ja", "jbe", "jl", "jg", "jle",
                                   "jge", "jns", "js", "loop"):
                continue
            count += 1
            print("  %3d  %-34s | %s" % (count, a, b))
    print("  （實質差異 %d 條）" % count)


if __name__ == "__main__":
    main()
