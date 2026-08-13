"""以 `55 89 e5` prologue 為界匯出函式，取代「相信 IDA 的函式清單」。

為什麼不用 `idautils.Functions()` ＋ `Chunks`：

- IDA 會把一支 Turbo Pascal 函式切成好幾個「函式」（共用出口、被 `jmp` 跳過
  的中段），`overlay-02:0C81h` 被切成五塊，標的 size 只有真實長度的 1/7。
- 更糟的是**有些位元組是 code 卻不屬於任何函式的 chunk**，於是完全不出現在
  匯出裡——逐條讀時看到的是一支空殼，而且沒有任何警告。

改以 prologue 為界之後，「一支函式」＝`[某個 prologue, 下一個 prologue)`，
區間內**所有已定義的指令**都會被匯出，與 `scripts/show.py --whole` 的口徑一致。

**prologue 有兩種寫法**，兩種都要認：

- `55 89 e5`——Turbo Pascal 編譯器產生的（`push bp` ＋ `mov bp, sp` 的短編碼）
- `55 8b ec`——組合語言寫的常式（MASM／TASM 的慣用編碼，同樣的兩條指令）

只認前者的話，組語常式的起點不會成為邊界，前一支函式的區間就一路吃到它後面。
`pc98/overlay-18:1756h` 其實只有 7 bytes（一支空函式），但區間被算到 189Dh，
看起來像「漏讀了 151 條指令」——**差點因此把正確的判讀退回**。

仍然會有缺口——真正 undefined（IDA 連指令都沒認出來）的位元組。這些在輸出裡
以 `gaps` 明確列出，不會被靜靜略過。

用法：
    tools/ida.sh py <module>.i64 export_by_prologue.py /work/out.json
"""

import json
import os
import sys

import ida_auto
import ida_bytes
import ida_funcs
import ida_name
import ida_pro
import ida_segment
import idautils
import idc


def segments():
    """所有 code segment 的 [start, end)。

    原本只取 `get_first_seg()`——raw overlay 只有一段所以看不出問題，但
    `PC98-GAME.EXE` 這種多段的 MZ 檔會**只匯出開頭那一段**（實測只出來 2 支
    函式），而輸出本身沒有任何異狀。
    """
    out = []
    segment = ida_segment.get_first_seg()
    while segment is not None:
        if segment.end_ea > segment.start_ea:
            out.append((segment.start_ea, segment.end_ea))
        segment = ida_segment.get_next_seg(segment.start_ea)
    return out


def main():
    out_path = sys.argv[1]
    ida_auto.auto_wait()

    starts, spans = [], segments()
    for base, end in spans:
        blob = ida_bytes.get_bytes(base, end - base) or b""
        found = set()
        for pattern in (b"\x55\x89\xe5", b"\x55\x8b\xec"):
            index = blob.find(pattern)
            while index >= 0:
                found.add(base + index)
                index = blob.find(pattern, index + 3)
        found = sorted(found)
        if not found or found[0] != base:
            found.insert(0, base)       # 有些段的第一支不以標準 prologue 開頭
        starts.extend(found)
    starts.sort()

    # 每一支的終點：下一個 prologue，但不得跨出所屬 segment。
    end_of = {}
    for base, end in spans:
        for start in starts:
            if base <= start < end:
                end_of[start] = end

    functions = []
    for position, start in enumerate(starts):
        stop = end_of[start]
        if position + 1 < len(starts):
            stop = min(stop, starts[position + 1])
        items, gaps, ea, gap_start = [], [], start, None
        while ea < stop:
            if ida_bytes.is_code(ida_bytes.get_flags(ea)):
                if gap_start is not None:
                    gaps.append([gap_start, ea]); gap_start = None
                size = max(1, ida_bytes.get_item_size(ea))
                raw = ida_bytes.get_bytes(ea, size) or b""
                items.append({
                    "ea": ea, "bytes": raw.hex(),
                    "disasm": idc.generate_disasm_line(ea, 0),
                    "code_refs": sorted(idautils.CodeRefsFrom(ea, 0)),
                    "data_refs": sorted(idautils.DataRefsFrom(ea)),
                })
                ea += size
            else:
                if gap_start is None:
                    gap_start = ea
                ea += max(1, ida_bytes.get_item_size(ea))
        if gap_start is not None:
            gaps.append([gap_start, stop])
        func = ida_funcs.get_func(start)
        functions.append({
            "ea": start, "end": stop, "size": stop - start,
            "name": ida_name.get_name(start) or "sub_%X" % start,
            "ida_size": (func.end_ea - func.start_ea) if func and func.start_ea == start else None,
            "arg_bytes": idc.get_frame_args_size(start) if func else 0,
            "callers": sorted({x for x in idautils.CodeRefsTo(start, 1)}),
            "gaps": gaps, "items": items,
        })

    os.makedirs(os.path.dirname(out_path) or ".", exist_ok=True)
    json.dump({"schema": "coab-prologue-functions/1", "functions": functions},
              open(out_path, "w", encoding="utf-8"), ensure_ascii=False)
    return 0


if __name__ == "__main__":
    try:
        rc = main()
    except BaseException:
        import traceback
        with open(sys.argv[1] + ".error.log", "w", encoding="utf-8") as handle:
            traceback.print_exc(file=handle)
        rc = 3
    ida_pro.qexit(rc)
