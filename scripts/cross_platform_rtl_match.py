"""用逐位元組比對，把 DOS 的 RTL 名稱轉移到 PC-98 resident。

PC-98 `GAME.EXE` 沒有 FLIRT 可辨識的 RTL 名稱，但兩個版本用同一套 Turbo Pascal
RTL 編譯。**body 完全相同就是同一支函式**——這是位元組級證據，不是名稱相似
或位址推算。

判準（刻意保守）：

- 兩邊函式的完整 body bytes 完全相同。
- 長度 `>= 16` bytes。太短的 body 有可能碰巧相同（例如只有 prologue 加一個
  `call` 的樁），不足以當證據。
- DOS 側必須有 Borland 還原出的名稱；沒有名稱就沒有可轉移的資訊。

轉移的是**識別**（這是 RTL 的哪一支），不是語意。與第 566 輪相同，
`Random`／`Randomize`／`Sound`／`NoSound`／`Delay` 這類影響玩家可見結果的
一律不標不阻塞。

用法：
    python3 scripts/cross_platform_rtl_match.py [--write]
"""

import collections
import json
import os
import sys

ROOT = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
SWEEP = os.path.join(ROOT, "workplace", "re-sweep")
LEDGER = os.path.join(ROOT, "docs", "audit", "re-function-ledger.json")
SPEC = "docs/spec/570-cross-platform-rtl-byte-match.md"
MIN_BYTES = 16
PLAYER_VISIBLE = ("Random", "Randomize", "SOUND", "NOSOUND", "DELAY")


def body(function):
    return "".join(item["bytes"] for item in function["items"])


def main():
    write = "--write" in sys.argv
    dos = json.load(open(os.path.join(SWEEP, "dos", "small", "START.EXE.big.json"),
                         encoding="utf-8"))["functions"]
    pc98 = json.load(open(os.path.join(SWEEP, "pc98", "small", "PC98-GAME.EXE.json"),
                          encoding="utf-8"))["functions"]

    by_body = collections.defaultdict(list)
    for function in dos:
        by_body[body(function)].append(function)

    entries, skipped, ambiguous = [], [], []
    for function in pc98:
        raw = body(function)
        if len(raw) // 2 < MIN_BYTES:
            continue
        named = [f for f in by_body.get(raw, []) if not f["name"].startswith("sub_")]
        if not named:
            continue
        names = {f["name"] for f in named}
        if len(names) > 1:
            ambiguous.append((function["ea"], sorted(names)))
            continue
        name = names.pop()
        if any(marker in name for marker in PLAYER_VISIBLE):
            skipped.append((function["ea"], name))
            continue
        entries.append({
            "platform": "pc98", "module": "PC98-GAME.EXE", "ea": function["ea"],
            "state": "不阻塞", "level": "", "spec": SPEC,
            "note": "Turbo Pascal RTL：body 與 DOS `START.EXE` 的 `%s` "
                    "逐位元組相同（%d bytes）" % (name, len(raw) // 2),
        })

    print("可轉移並標不阻塞：%d" % len(entries))
    print("名稱不唯一而略過：%d" % len(ambiguous))
    print("玩家可見而保留待解讀：%d" % len(skipped))
    for ea, name in skipped:
        print("  %05Xh %s" % (ea, name))

    if not write:
        print("\n（預覽模式；加 --write 才寫入台帳）")
        return 0

    ledger = json.load(open(LEDGER, encoding="utf-8"))
    keys = {(e["platform"], e["module"], e["ea"]) for e in entries}
    kept = [e for e in ledger["functions"]
            if (e["platform"], e["module"], e["ea"]) not in keys]
    ledger["functions"] = kept + entries
    json.dump(ledger, open(LEDGER, "w", encoding="utf-8"), ensure_ascii=False, indent=1)
    print("\n已寫入 %s" % LEDGER)
    return 0


if __name__ == "__main__":
    sys.exit(main())
