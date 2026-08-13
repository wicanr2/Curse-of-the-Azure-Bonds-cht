"""匯出所有「小函式」的完整逐指令內容，供批次逐條閱讀。

小函式的整個 body 只有幾條指令，把它們一次全部匯出之後，閱讀與分類就不必
每個函式重開一次 IDA。**匯出的是完整 body（不是摘要），所以據此判讀等同
逐指令讀過。**

用法：
    tools/ida.sh py <module>.i64 export_small_functions.py /work/small.json [最大位元組數]

預設 48 bytes。超過門檻的函式不會出現在輸出裡——它們必須另外讀，不得由
本檔的缺席推論任何事。
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


def body(func_ea):
    items = []
    total = 0
    for chunk in idautils.Chunks(func_ea):
        ea = chunk[0]
        while ea < chunk[1]:
            size = max(1, ida_bytes.get_item_size(ea))
            total += size
            raw = ida_bytes.get_bytes(ea, size) or b""
            items.append({
                "ea": ea,
                "bytes": raw.hex(),
                "disasm": idc.generate_disasm_line(ea, 0),
                "code_refs": sorted(idautils.CodeRefsFrom(ea, 0)),
                "data_refs": sorted(idautils.DataRefsFrom(ea)),
            })
            ea += size
    return items, total


def main():
    if len(sys.argv) < 2:
        return 2
    out_path = sys.argv[1]
    limit = int(sys.argv[2]) if len(sys.argv) > 2 else 48
    ida_auto.auto_wait()

    functions = []
    for func_ea in idautils.Functions():
        items, total = body(func_ea)
        if total > limit:
            continue
        func = ida_funcs.get_func(func_ea)
        functions.append({
            "ea": func_ea,
            "name": ida_name.get_name(func_ea),
            "size": total,
            "arg_bytes": idc.get_frame_args_size(func_ea),
            "callers": sorted({x for x in idautils.CodeRefsTo(func_ea, 1)}),
            "is_lib": bool(func.flags & ida_funcs.FUNC_LIB),
            "items": items,
        })

    os.makedirs(os.path.dirname(out_path) or ".", exist_ok=True)
    with open(out_path, "w", encoding="utf-8") as handle:
        json.dump({"schema": "coab-small-functions/1", "limit": limit,
                   "functions": functions}, handle, ensure_ascii=False)
    return 0


if __name__ == "__main__":
    try:
        rc = main()
    except BaseException:
        import traceback
        log_path = (sys.argv[1] if len(sys.argv) > 1 else "/work/small") + ".error.log"
        try:
            os.makedirs(os.path.dirname(log_path) or ".", exist_ok=True)
            with open(log_path, "w", encoding="utf-8") as handle:
                traceback.print_exc(file=handle)
        except BaseException:
            pass
        rc = 3
    ida_pro.qexit(rc)
