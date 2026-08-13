"""抽出 `CHECKFX` 的「時機 → effect id 清單」分派表。

`CHECKFX(timing, subject)` 不是一支做很多事的大函式，而是一張表：每個 timing
值對應一組 effect id，逐一交給 `sub_269`（effect 鏈遍歷，spec 577）去找目標
身上有沒有該效果、有就 `CALLEFFECT` 分派。

所以命中判定裡的 `CHECKFX(0Ah, target)`＋`CHECKFX(10h, attacker)`、豁免裡的
`CHECKFX(0Ch, char)`、傷害裡的 `CHECKFX(06h)`／`CHECKFX(14h)`、施加效果的
`CHECKFX(09h)`——問的都是「這個時機有哪些效果要介入」。

解析對象是 `export_small_functions.py`（IDAPython）匯出的逐指令序列，判準是
純結構的：

    cmp al, N / jnz NEXT                  → body 在後，下一個 case 在 NEXT
    cmp al, N / jz BODY / jmp NEXT        → body 在 BODY，下一個 case 在 NEXT
    mov al, X / push ax / push cs / call  → 該 case 觸發 effect X

**兩種形狀都要處理。** body 超過 short jump 的 127 bytes 時，編譯器就從第一種
換成第二種；只認第一種會在第一個大 case 上停住（`CHECKFX` 是 `05h`，於是
`00h..17h` 24 個 case 只解出 6 個，而且不會有任何錯誤訊息）。

用法：python3 scripts/checkfx_timing_table.py <platform> <module> <start16> [--write]
"""

import json
import os
import re
import sys

ROOT = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
SWEEP = os.path.join(ROOT, "workplace", "re-sweep")
OUT_MD = os.path.join(ROOT, "docs", "audit", "checkfx-timing-table.md")


def instructions(platform, module):
    path = os.path.join(SWEEP, platform, "overlays", "full",
                        "%s-%s.json" % (platform, module))
    items = [it for f in json.load(open(path, encoding="utf-8"))["functions"]
             for it in f["items"]]
    items.sort(key=lambda it: it["ea"])
    return {it["ea"]: it for it in items}, items


def immediate(text):
    match = re.search(r",\s*([0-9A-Fa-f]+h|[0-9]+)\b", text)
    if not match:
        return None
    raw = match.group(1)
    return int(raw[:-1], 16) if raw.endswith("h") else int(raw)


def target(item):
    """short/near jump 的絕對目標，取 IDA 已解出的 code ref。"""
    refs = [r for r in item["code_refs"] if r != item["ea"] + len(item["bytes"]) // 2]
    return refs[0] if refs else None


def parse(platform, module, start):
    by_ea, items = instructions(platform, module)
    order = [it["ea"] for it in items]
    table, cursor = [], start
    guard = 0
    while cursor is not None and guard < 4000:
        guard += 1
        item = by_ea.get(cursor)
        if item is None:
            break
        text = re.sub(r"\s*;.*$", "", item["disasm"].strip())
        if text.startswith("cmp") and " al," in text:
            value = immediate(text)
            index = order.index(cursor)
            jump = by_ea[order[index + 1]]
            first = re.sub(r"\s*;.*$", "", jump["disasm"].strip())
            if first.startswith("jz"):
                body = target(jump)
                second = by_ea[order[index + 2]]
                end = target(second)
                body_index = order.index(body) if body in order else index + 3
            else:
                end = target(jump)
                body_index = index + 2
            if end is None:
                break
            ids, pending = [], None
            for ea in order[body_index:]:
                if ea >= end:
                    break
                line = re.sub(r"\s*;.*$", "", by_ea[ea]["disasm"].strip())
                if line.startswith("mov") and " al," in line:
                    pending = immediate(line)
                elif line.startswith("call") and pending is not None:
                    ids.append((pending, line.split()[-1]))
                    pending = None
            table.append({"timing": value, "end": end,
                          "calls": [{"id": i, "callee": c} for i, c in ids]})
            cursor = end
            continue
        # 下一個 case 起點不一定緊接著就是 `cmp al, N`（中間可能有 case body
        # 的收尾）。往下找，連續這麼多條都沒遇到就當作鏈結束。
        index = order.index(cursor)
        cursor = None
        for ea in order[index + 1:index + 40]:
            line = re.sub(r"\s*;.*$", "", by_ea[ea]["disasm"].strip())
            if line.startswith("cmp") and " al," in line:
                cursor = ea
                break
    return table


def main():
    platform, module, start = sys.argv[1], sys.argv[2], int(sys.argv[3], 16)
    # 最後一輪會走到 epilogue 前的 fallthrough，沒有 timing 值，丟掉。
    table = [row for row in parse(platform, module, start) if row["timing"] is not None]
    print("%s %s %04Xh：解出 %d 個 timing" % (platform, module, start, len(table)))
    for row in table:
        ids = ", ".join("%02Xh" % c["id"] for c in row["calls"])
        print("  timing %02Xh → %d 個：%s" % (row["timing"], len(row["calls"]), ids or "（無）"))

    if "--write" not in sys.argv:
        print("\n（預覽模式；加 --write 才寫報表）")
        return 0

    lines = ["# `CHECKFX` 的時機分派表", "",
             "`CHECKFX(timing, subject)` 是一張表：每個 timing 對應一組 effect id，",
             "逐一交給 effect 鏈遍歷（`sub_269`，[spec 577](../spec/577-attempttohit-and-effect-chain-walk.md)）",
             "去看目標身上有沒有該效果，有就 `CALLEFFECT` 分派。",
             "",
             "所以規則各處的 `CHECKFX(0Ah)`／`CHECKFX(0Ch)`／`CHECKFX(06h)` 問的都是",
             "「這個時機有哪些效果要介入」。",
             "",
             "由 `scripts/checkfx_timing_table.py` 從 IDAPython 匯出的逐指令序列解析，",
             "判準是純結構的（`cmp al, N` ／ `jnz` 的目標即下一個 case 起點）。", "",
             "| timing | effect id |", "|---|---|"]
    for row in table:
        ids = "、".join("`%02Xh`" % c["id"] for c in row["calls"]) or "（無）"
        lines.append("| `%02Xh` | %s |" % (row["timing"], ids))
    lines += ["", "## 已知的呼叫點", "",
              "| 呼叫處 | timing | 出處 |", "|---|---|---|",
              "| `PUTDAMAGE` 進入時 | `06h` | [581](../spec/581-putdamage-pipeline.md) |",
              "| `PUTDAMAGE` 無豁免時 | `14h` | 同上 |",
              "| `PUTEFFECT` | `09h` | 同上 |",
              "| `ATTEMPTTOHIT` 對目標 | `0Ah` | [577](../spec/577-attempttohit-and-effect-chain-walk.md) |",
              "| `ATTEMPTTOHIT` 對攻擊者 | `10h` | 同上 |",
              "| `MAKESAVE` | `0Ch` | [582](../spec/582-makesave-and-losedude.md) |",
              "| `KILLDUDE`／`PUTDAMAGE` 死亡後 | `0Dh` | [579](../spec/579-character-status-fields.md) |"]
    open(OUT_MD, "w", encoding="utf-8").write("\n".join(lines) + "\n")
    print("→ %s" % OUT_MD)
    return 0


if __name__ == "__main__":
    sys.exit(main())
