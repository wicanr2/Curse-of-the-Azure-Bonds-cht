"""列出一個 module 內每個函式的呼叫指令與目標（含 far call 的 segment:offset）。

用途：Turbo Pascal 的跨 unit 呼叫是 far call，IDA 在單一 overlay 的 database
裡不會把它建成函式，所以 `export_module.py` 的 `calls` 只涵蓋 near call。
要問「這個 handler 呼叫了哪個共用 routine 幾次」就需要這一份。

輸出每筆：函式起點、指令位址、raw bytes、助憶碼、目標（near 為 code-local
位址；far 為 `seg:off`）。不做語意判斷。

用法：
    tools/ida.sh py <module>.i64 export_calls.py /work/calls.json
"""

import json
import os
import sys

import ida_auto
import ida_bytes
import ida_funcs
import ida_pro
import idautils
import idc


def far_target(ea, raw):
    """`9A off_lo off_hi seg_lo seg_hi` 是 far call ptr16:16。"""
    if len(raw) >= 5 and raw[0] == 0x9A:
        offset = raw[1] | (raw[2] << 8)
        segment = raw[3] | (raw[4] << 8)
        return "%04X:%04X" % (segment, offset)
    return None


def main():
    if len(sys.argv) < 2:
        return 2
    out_path = sys.argv[1]
    ida_auto.auto_wait()

    records = []
    for func_ea in idautils.Functions():
        func = ida_funcs.get_func(func_ea)
        if func is None:
            continue
        for chunk in idautils.Chunks(func_ea):
            ea = chunk[0]
            while ea < chunk[1]:
                size = max(1, ida_bytes.get_item_size(ea))
                mnemonic = idc.print_insn_mnem(ea)
                if mnemonic and mnemonic.lower().startswith("call"):
                    raw = ida_bytes.get_bytes(ea, size) or b""
                    refs = sorted(idautils.CodeRefsFrom(ea, 0))
                    records.append({
                        "function": func_ea,
                        "ea": ea,
                        "bytes": raw.hex(),
                        "disasm": idc.generate_disasm_line(ea, 0),
                        "near_targets": refs,
                        "far_target": far_target(ea, raw),
                    })
                ea += size

    os.makedirs(os.path.dirname(out_path) or ".", exist_ok=True)
    with open(out_path, "w", encoding="utf-8") as handle:
        json.dump({"schema": "coab-ida-calls/1", "calls": records}, handle,
                  ensure_ascii=False)
    return 0


if __name__ == "__main__":
    try:
        rc = main()
    except BaseException:
        import traceback
        log_path = (sys.argv[1] if len(sys.argv) > 1 else "/work/calls") + ".error.log"
        try:
            os.makedirs(os.path.dirname(log_path) or ".", exist_ok=True)
            with open(log_path, "w", encoding="utf-8") as handle:
                traceback.print_exc(file=handle)
        except BaseException:
            pass
        rc = 3
    ida_pro.qexit(rc)
