"""一次 dump 多支函式（或位址範圍），輸出 raw bytes、反組譯與 xref。

與 `dump_function.py` 的差別只有批次：語意閉合常常一次要看十幾支 handler，
每支各起一次容器的成本（映像載入＋auto_wait）遠大於 dump 本身。

用法：
    tools/ida.sh py <module>.i64 dump_functions_batch.py /work/out/x.json <spec>[,<spec>...]

`spec` 是 `start` 或 `start:end`，十六進位（可帶 0x）。省略 end 時 dump 整個函式。
不改 database：不 rename、不加註解。

輸出（JSON）：
    {"functions": [{"start", "end", "name", "item_count", "items": [...]}, ...]}
失敗時 traceback 寫進 `<輸出路徑>.error.log`——headless 的 stdout 與 exit code
都不可信，唯一可信訊號是輸出檔。
"""

import json
import os
import sys
import traceback

import ida_auto
import ida_bytes
import ida_funcs
import ida_name
import ida_pro
import idautils
import idc


def parse_ea(text):
    return int(text, 16)


def dump_range(start, end):
    items = []
    ea = start
    while ea < end:
        size = max(1, ida_bytes.get_item_size(ea))
        raw = ida_bytes.get_bytes(ea, size) or b""
        items.append({
            "ea": ea,
            "bytes": raw.hex(),
            "disasm": idc.generate_disasm_line(ea, 0),
            "is_code": bool(ida_bytes.is_code(ida_bytes.get_full_flags(ea))),
            "code_refs": sorted(idautils.CodeRefsFrom(ea, 0)),
            "data_refs": sorted(idautils.DataRefsFrom(ea)),
        })
        ea += size
    return items


def resolve(spec):
    if ":" in spec:
        head, tail = spec.split(":", 1)
        return parse_ea(head), parse_ea(tail)
    start = parse_ea(spec)
    func = ida_funcs.get_func(start)
    if func is None or func.start_ea != start:
        # 沒有被辨識成函式起點時，退回固定長度切片，讓人工判讀，
        # 不在這裡猜邊界。
        return start, start + 0x120
    return func.start_ea, func.end_ea


def main():
    if len(sys.argv) < 3:
        return 2
    out_path = sys.argv[1]
    specs = [s for s in ",".join(sys.argv[2:]).split(",") if s]
    ida_auto.auto_wait()

    functions = []
    for spec in specs:
        start, end = resolve(spec)
        items = dump_range(start, end)
        functions.append({
            "spec": spec,
            "start": start,
            "end": end,
            "name": ida_name.get_ea_name(start) or "",
            "is_function_start": ida_funcs.get_func(start) is not None
            and ida_funcs.get_func(start).start_ea == start,
            "item_count": len(items),
            "items": items,
        })

    with open(out_path, "w", encoding="utf-8") as fh:
        json.dump({"functions": functions}, fh, ensure_ascii=False, indent=1)
    return 0


if __name__ == "__main__":
    code = 2
    try:
        code = main()
    except Exception:
        try:
            with open(sys.argv[1] + ".error.log", "w", encoding="utf-8") as fh:
                fh.write(traceback.format_exc())
        except Exception:
            pass
    ida_pro.qexit(code or 0)
