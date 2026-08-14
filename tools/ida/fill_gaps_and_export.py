"""把 `export_by_prologue.py` 留下的 `gaps` 強制解成指令，再用同一個 schema 匯出。

問題：prologue 匯出裡有些位元組 IDA 從頭到尾沒認成 code（`is_code` 為假），
於是那一段完全不出現在 `items` 裡。**逐條讀的時候看不到缺口**——相鄰兩條指令
的 `ea` 中間少了幾個 byte，讀起來卻像是連續的。實測後果：
`overlay-12:356Ch` 被讀成「`push bp / mov bp,sp / in al,dx / pop bp / retf`」，
`in al, dx` 完全是幻覺——真相是 `89 EC`（`mov sp,bp`）掉了一個 byte，那是一支
空函式（spec 736、752）。

作法：對每個 gap 區間，先 `del_items` 清掉可能存在的資料定義，再從區間起點
逐條 `create_insn`。**一條都建不起來就跳過**（真的是資料，不硬幹）。全部處理
完再跑一次 `auto_wait`，然後以與 `export_by_prologue.py` 完全相同的邏輯匯出，
輸出仍帶 `gaps`——**補不起來的缺口要留在輸出裡，不能靜靜消失**。

輸出另外多一個 `filled` 欄位，記錄這一輪實際補了哪些區間，方便事後核對。

用法：
    tools/ida.sh py <module>.i64 fill_gaps_and_export.py /work/out.json
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
import ida_ua
import idautils
import idc

PROLOGUES = (b"\x55\x89\xe5", b"\x55\x8b\xec")


def segments():
    out = []
    segment = ida_segment.get_first_seg()
    while segment is not None:
        if segment.end_ea > segment.start_ea:
            out.append((segment.start_ea, segment.end_ea))
        segment = ida_segment.get_next_seg(segment.start_ea)
    return out


def prologue_starts(spans):
    starts = []
    for base, end in spans:
        blob = ida_bytes.get_bytes(base, end - base) or b""
        found = set()
        for pattern in PROLOGUES:
            index = blob.find(pattern)
            while index >= 0:
                found.add(base + index)
                index = blob.find(pattern, index + 3)
        found = sorted(found)
        if not found or found[0] != base:
            found.insert(0, base)
        starts.extend(found)
    starts.sort()
    return starts


def scan_gaps(start, stop):
    """回傳 [start, stop) 內未定義成 code 的區間。"""
    gaps, ea, gap_start = [], start, None
    while ea < stop:
        if ida_bytes.is_code(ida_bytes.get_flags(ea)):
            if gap_start is not None:
                gaps.append((gap_start, ea))
                gap_start = None
            ea += max(1, ida_bytes.get_item_size(ea))
        else:
            if gap_start is None:
                gap_start = ea
            ea += max(1, ida_bytes.get_item_size(ea))
    if gap_start is not None:
        gaps.append((gap_start, stop))
    return gaps


def fill(gap_start, gap_stop):
    """在區間內逐條建立指令；回傳實際補起來的 byte 數。"""
    ida_bytes.del_items(gap_start, ida_bytes.DELIT_SIMPLE, gap_stop - gap_start)
    ea, filled = gap_start, 0
    while ea < gap_stop:
        size = ida_ua.create_insn(ea)
        if not size:
            break
        filled += size
        ea += size
    return filled


def collect(start, stop):
    items, gaps, ea, gap_start = [], [], start, None
    while ea < stop:
        if ida_bytes.is_code(ida_bytes.get_flags(ea)):
            if gap_start is not None:
                gaps.append([gap_start, ea])
                gap_start = None
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
    return items, gaps


def main():
    out_path = sys.argv[1]
    ida_auto.auto_wait()

    spans = segments()
    starts = prologue_starts(spans)
    end_of = {}
    for base, end in spans:
        for start in starts:
            if base <= start < end:
                end_of[start] = end

    bounds = []
    for position, start in enumerate(starts):
        stop = end_of[start]
        if position + 1 < len(starts):
            stop = min(stop, starts[position + 1])
        bounds.append((start, stop))

    filled = []
    for start, stop in bounds:
        for gap_start, gap_stop in scan_gaps(start, stop):
            got = fill(gap_start, gap_stop)
            if got:
                filled.append([gap_start, gap_start + got])
    ida_auto.auto_wait()

    functions = []
    for start, stop in bounds:
        items, gaps = collect(start, stop)
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
    json.dump({"schema": "coab-prologue-functions/1", "filled": filled,
               "functions": functions},
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
