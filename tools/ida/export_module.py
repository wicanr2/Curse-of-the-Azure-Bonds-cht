"""把一個 IDA database 的全部靜態事實匯出成 JSON（全函式覆蓋台帳的原料）。

用法（一律經 tools/ida.sh）：
    tools/ida.sh py <module>.i64 export_module.py /work/out/<module>.json

刻意只匯出 IDA 自己標好的事實：函式、xref 圖、字串、segment、未定義區。
不做語意猜測，也不改 database（不 rename、不加註解）——語意一律放外部 ledger。

輸出結構見同目錄 README 或 docs/spec/559。
"""

import hashlib
import json
import os
import sys

import ida_auto
import ida_bytes
import ida_funcs
import ida_ida
import ida_nalt
import ida_name
import ida_pro
import ida_segment
import idautils
import idc


def input_sha256():
    """IDA 只保存 MD5/SHA256 之一視版本而定，取不到就自己算原始輸入檔。"""
    for getter in ("retrieve_input_file_sha256", "retrieve_input_file_md5"):
        fn = getattr(ida_nalt, getter, None)
        if fn is None:
            continue
        try:
            value = fn()
        except Exception:
            continue
        if value:
            return {"algo": getter.replace("retrieve_input_file_", ""),
                    "value": value.hex() if isinstance(value, bytes) else str(value)}
    path = ida_nalt.get_input_file_path()
    if path and os.path.exists(path):
        digest = hashlib.sha256()
        with open(path, "rb") as handle:
            for chunk in iter(lambda: handle.read(1 << 20), b""):
                digest.update(chunk)
        return {"algo": "sha256", "value": digest.hexdigest()}
    return {"algo": "unknown", "value": ""}


def segments():
    out = []
    for ea in idautils.Segments():
        seg = ida_segment.getseg(ea)
        if seg is None:
            continue
        out.append({
            "name": ida_segment.get_segm_name(seg),
            "class": ida_segment.get_segm_class(seg),
            "start": seg.start_ea,
            "end": seg.end_ea,
            "bitness": seg.bitness,          # 0=16bit 1=32bit 2=64bit
            "sel": seg.sel,
            # 16-bit：段基址是 paragraph，linear = para<<4；與 segment:offset 對照要靠它
            "para": ida_segment.sel2para(seg.sel),
        })
    return out


def function_record(ea):
    func = ida_funcs.get_func(ea)
    name = ida_name.get_name(ea)
    # 9.4 的 idautils.Chunks 產生 (start, end) tuple，不是物件。
    chunks = [{"start": c[0], "end": c[1]} for c in idautils.Chunks(ea)]
    size = sum(c["end"] - c["start"] for c in chunks)

    calls_to, data_refs, code_refs_out = set(), set(), set()
    insn_count = 0
    for chunk in chunks:
        cur = chunk["start"]
        while cur < chunk["end"]:
            insn_count += 1
            for ref in idautils.CodeRefsFrom(cur, 0):
                code_refs_out.add(ref)
                target = ida_funcs.get_func(ref)
                if target is not None:
                    calls_to.add(target.start_ea)
            for ref in idautils.DataRefsFrom(cur):
                data_refs.add(ref)
            nxt = idc.next_head(cur, chunk["end"])
            cur = nxt if nxt > cur else cur + 1

    callers = set()
    for ref in idautils.CodeRefsTo(ea, 1):
        caller = ida_funcs.get_func(ref)
        callers.add(caller.start_ea if caller else ref)

    return {
        "ea": ea,
        "name": name,
        "auto_named": name.startswith(("sub_", "nullsub_", "loc_", "unknown_")),
        "start": func.start_ea,
        "end": func.end_ea,
        "size": size,
        "chunks": chunks,
        "instructions": insn_count,
        "flags": func.flags,
        "is_lib": bool(func.flags & ida_funcs.FUNC_LIB),
        "is_thunk": bool(func.flags & ida_funcs.FUNC_THUNK),
        "frame_size": idc.get_frame_size(ea),
        "ret_bytes": idc.get_frame_args_size(ea),
        "calls": sorted(calls_to),
        "callers": sorted(callers),
        "data_refs": sorted(data_refs),
    }


def strings():
    out = []
    sc = idautils.Strings()
    sc.setup(strtypes=[ida_nalt.STRTYPE_C, ida_nalt.STRTYPE_C_16,
                       ida_nalt.STRTYPE_PASCAL])
    sc.refresh()
    for item in sc:
        try:
            text = str(item)
        except Exception:
            continue
        refs = sorted(idautils.DataRefsTo(item.ea))
        out.append({"ea": item.ea, "length": item.length,
                    "type": item.strtype, "text": text, "refs": refs})
    return out


def named_data():
    """有名字、且不在函式內的資料位址（全域變數候選）。"""
    out = []
    for ea, name in idautils.Names():
        if ida_funcs.get_func(ea) is not None:
            continue
        refs = [{"frm": x.frm, "type": x.type, "iscode": x.iscode}
                for x in idautils.XrefsTo(ea)]
        out.append({"ea": ea, "name": name, "size": ida_bytes.get_item_size(ea),
                    "xrefs": refs})
    return out


def undefined_ranges():
    """code segment 內尚未被 IDA 定義成指令／資料的區段：真正的『沒看過』。"""
    out = []
    for ea in idautils.Segments():
        seg = ida_segment.getseg(ea)
        if seg is None or ida_segment.get_segm_class(seg) not in ("CODE", None, ""):
            continue
        cur, run_start = seg.start_ea, None
        while cur < seg.end_ea:
            flags = ida_bytes.get_full_flags(cur)
            undefined = not (ida_bytes.is_code(flags) or ida_bytes.is_data(flags))
            if undefined and run_start is None:
                run_start = cur
            elif not undefined and run_start is not None:
                out.append({"start": run_start, "end": cur})
                run_start = None
            cur += max(1, ida_bytes.get_item_size(cur))
        if run_start is not None:
            out.append({"start": run_start, "end": seg.end_ea})
    return out


def build_payload():
    """收集整個 database 的靜態事實；analyze_overlay.py 也直接呼叫這支。"""
    funcs = [function_record(ea) for ea in idautils.Functions()]
    return {
        "schema": "coab-ida-module-export/1",
        "input": {
            "path": ida_nalt.get_input_file_path(),
            "name": os.path.basename(ida_nalt.get_input_file_path() or ""),
            "hash": input_sha256(),
            "procname": ida_ida.inf_get_procname(),
            "min_ea": ida_ida.inf_get_min_ea(),
            "max_ea": ida_ida.inf_get_max_ea(),
        },
        "segments": segments(),
        "functions": funcs,
        "strings": strings(),
        "named_data": named_data(),
        "undefined_ranges": undefined_ranges(),
        "totals": {
            "functions": len(funcs),
            "instructions": sum(f["instructions"] for f in funcs),
            "code_bytes": sum(f["size"] for f in funcs),
        },
    }


def main():
    if len(sys.argv) < 2:
        return 2
    out_path = sys.argv[1]
    ida_auto.auto_wait()
    payload = build_payload()

    os.makedirs(os.path.dirname(out_path) or ".", exist_ok=True)
    with open(out_path, "w", encoding="utf-8") as handle:
        json.dump(payload, handle, ensure_ascii=False)
    return 0


# 被 analyze_overlay.py import 時只提供函式；只有直接被 IDA 執行才收工離開。
if __name__ == "__main__":
    try:
        rc = main()
    except BaseException:  # headless 沒有 stderr，錯誤一律落檔，否則就是靜默失敗
        import traceback
        log_path = (sys.argv[1] if len(sys.argv) > 1 else "/work/export") + ".error.log"
        try:
            os.makedirs(os.path.dirname(log_path) or ".", exist_ok=True)
            with open(log_path, "w", encoding="utf-8") as handle:
                traceback.print_exc(file=handle)
        except BaseException:
            pass
        rc = 3
    ida_pro.qexit(rc)
