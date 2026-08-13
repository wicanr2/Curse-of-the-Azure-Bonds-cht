"""以 `55 89 e5` prologue 為界匯出函式，取代「相信 IDA 的函式清單」。

為什麼不用 `idautils.Functions()` ＋ `Chunks`：

- IDA 會把一支 Turbo Pascal 函式切成好幾個「函式」（共用出口、被 `jmp` 跳過
  的中段），`overlay-02:0C81h` 被切成五塊，標的 size 只有真實長度的 1/7。
- 更糟的是**有些位元組是 code 卻不屬於任何函式的 chunk**，於是完全不出現在
  匯出裡——逐條讀時看到的是一支空殼，而且沒有任何警告。

改以 prologue 為界之後，「一支函式」＝`[某個 55 89 e5, 下一個 55 89 e5)`，
區間內**所有已定義的指令**都會被匯出，與 `scripts/show.py --whole` 的口徑一致。

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


def main():
    out_path = sys.argv[1]
    ida_auto.auto_wait()
    segment = ida_segment.get_first_seg()
    base, end = segment.start_ea, segment.end_ea
    blob = ida_bytes.get_bytes(base, end - base) or b""

    starts, index = [], blob.find(b"\x55\x89\xe5")
    while index >= 0:
        starts.append(base + index)
        index = blob.find(b"\x55\x89\xe5", index + 3)
    if not starts or starts[0] != base:
        starts.insert(0, base)          # 有些 overlay 的第一支不以標準 prologue 開頭

    functions = []
    for position, start in enumerate(starts):
        stop = starts[position + 1] if position + 1 < len(starts) else end
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
