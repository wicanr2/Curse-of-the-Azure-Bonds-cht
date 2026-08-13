"""逐指令 dump 一個函式（或位址範圍），輸出 raw bytes、反組譯與 xref。

這是語意閉合階段的主力工具：結論一律要能回指到「哪個模組、哪個位址、哪幾個
byte」。輸出同時給人看與給程式讀（JSON），避免再用 grep `.asm` 當證據。

用法：
    tools/ida.sh py <module>.i64 dump_function.py /work/out/x.json <start> [end]

`start`／`end` 是十六進位（可帶 0x）。省略 end 時 dump 整個函式。
不改 database：不 rename、不加註解。
"""

import json
import os
import sys

import ida_auto
import ida_bytes
import ida_funcs
import ida_name
import ida_pro
import idautils
import idc


def parse_ea(text):
    return int(text, 16)


def dump(start, end):
    items = []
    ea = start
    while ea < end:
        size = max(1, ida_bytes.get_item_size(ea))
        raw = ida_bytes.get_bytes(ea, size) or b""
        refs_code = sorted(idautils.CodeRefsFrom(ea, 0))
        refs_data = sorted(idautils.DataRefsFrom(ea))
        items.append({
            "ea": ea,
            "bytes": raw.hex(),
            "disasm": idc.generate_disasm_line(ea, 0),
            "is_code": bool(ida_bytes.is_code(ida_bytes.get_full_flags(ea))),
            "code_refs": refs_code,
            "data_refs": refs_data,
            "xrefs_to": [x.frm for x in idautils.XrefsTo(ea)],
        })
        ea += size
    return items


def main():
    if len(sys.argv) < 3:
        return 2
    out_path = sys.argv[1]
    start = parse_ea(sys.argv[2])
    ida_auto.auto_wait()

    func = ida_funcs.get_func(start)
    if len(sys.argv) > 3:
        end = parse_ea(sys.argv[3])
    elif func is not None:
        end = func.end_ea
    else:
        end = start + 0x40

    payload = {
        "schema": "coab-ida-function-dump/1",
        "start": start,
        "end": end,
        "function": None if func is None else {
            "start": func.start_ea, "end": func.end_ea,
            "name": ida_name.get_name(func.start_ea),
            "callers": sorted({x for x in idautils.CodeRefsTo(func.start_ea, 1)}),
        },
        "items": dump(start, end),
    }
    os.makedirs(os.path.dirname(out_path) or ".", exist_ok=True)
    with open(out_path, "w", encoding="utf-8") as handle:
        json.dump(payload, handle, ensure_ascii=False)
    return 0


if __name__ == "__main__":
    try:
        rc = main()
    except BaseException:
        import traceback
        log_path = (sys.argv[1] if len(sys.argv) > 1 else "/work/dump") + ".error.log"
        try:
            os.makedirs(os.path.dirname(log_path) or ".", exist_ok=True)
            with open(log_path, "w", encoding="utf-8") as handle:
                traceback.print_exc(file=handle)
        except BaseException:
            pass
        rc = 3
    ida_pro.qexit(rc)
