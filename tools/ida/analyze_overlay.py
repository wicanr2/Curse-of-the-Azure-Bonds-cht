"""把一段 raw TPOV overlay code 變成可分析的 IDA database，再匯出全函式清冊。

為什麼需要這支：raw overlay 沒有 MZ header、沒有 entry point，直接丟進 IDA
自動分析只會得到 0–1 個函式（實測 DOS GAME.OVR 整檔載入只有 1 個）。真正的
entry point 全在 resident executable 的 `CD 3F` five-byte stub 裡，由
cmd/ovr-manifest 匯出成 JSON；本腳本把那些 handler-local offset 當種子。

用法（經 tools/ida.sh，database 由 idat -A -p8086 -b0 當場建立）：
    tools/ida.sh raw idat -A -p8086 -b0 \
        "-S/work-tools/analyze_overlay.py <manifest.json> <module> <out.json>" overlay-02.bin

不改任何原始識別：不 rename、不加註解。種子只做 `add_func`，語意一律留給
外部 ledger。
"""

import json
import os
import sys

import ida_auto
import ida_bytes
import ida_funcs
import ida_pro
import ida_segment
import ida_ua

sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
import export_module  # noqa: E402


def force_16bit():
    """binary loader 可能給 32-bit segment；16-bit real mode 解錯會全盤皆錯。"""
    changed = []
    for ea in list(__import__("idautils").Segments()):
        seg = ida_segment.getseg(ea)
        if seg is None or seg.bitness == 0:
            continue
        seg.bitness = 0
        seg.update()
        changed.append(ea)
    return changed


def seed(entries):
    """把 entry stub 的 handler-local offset 標成函式起點。

    entries 的每筆至少要有 `code_offset`；額外種子（Borland 符號、unit 初始化
    的 offset 0）用同一條路徑進來，只是 origin 欄位不同。
    """
    made, failed = [], []
    for entry in entries:
        ea = entry["code_offset"]
        ida_bytes.del_items(ea, ida_bytes.DELIT_SIMPLE, 1)
        if ida_ua.create_insn(ea) <= 0:
            failed.append(entry)
            continue
        if ida_funcs.get_func(ea) is None and not ida_funcs.add_func(ea):
            failed.append(entry)
            continue
        made.append(entry)
    return made, failed


def sweep_unreached():
    """entry 之外的函式：對尚未定義的區段做一次保守的指令化嘗試。

    Turbo Pascal 的 overlay 內部呼叫是 near call，自動分析會從種子往下傳染；
    這一步只補「完全沒被任何種子觸及」的殘餘區段，並且只在該處能成功建立
    指令時才建函式。不能證明它是真的函式，因此輸出另外標記 origin=sweep。
    """
    import idautils

    created = []
    for seg_ea in list(idautils.Segments()):
        seg = ida_segment.getseg(seg_ea)
        if seg is None:
            continue
        ea = seg.start_ea
        while ea < seg.end_ea:
            flags = ida_bytes.get_full_flags(ea)
            if ida_bytes.is_code(flags) or ida_bytes.is_data(flags):
                ea += max(1, ida_bytes.get_item_size(ea))
                continue
            if ida_ua.create_insn(ea) > 0 and ida_funcs.add_func(ea):
                created.append(ea)
                func = ida_funcs.get_func(ea)
                ea = func.end_ea if func else ea + 1
            else:
                ea += 1
    ida_auto.auto_wait()
    return created


def main():
    if len(sys.argv) < 4:
        return 2
    manifest_path, module, out_path = sys.argv[1], sys.argv[2], sys.argv[3]
    do_sweep = "--sweep" in sys.argv[4:]

    with open(manifest_path, encoding="utf-8") as handle:
        manifest = json.load(handle)
    record = next((o for o in manifest["overlays"] if o["module"] == module), None)
    if record is None:
        raise SystemExit("manifest 沒有 module %s" % module)

    # 額外種子：unit 初始化一律在 code offset 0（TPOV entry stub 不含它），
    # 另可由 seeds.json 帶入 Borland 符號位址 {module: [offset, ...]}。
    extra = [{"index": -1, "stub_offset": 0, "code_offset": 0, "flags": 0,
              "origin": "unit-init"}]
    seeds_path = next((a for a in sys.argv[4:] if a.endswith(".json")), None)
    if seeds_path:
        with open(seeds_path, encoding="utf-8") as handle:
            table = json.load(handle)
        for offset in sorted(set(table.get(module, []))):
            extra.append({"index": -1, "stub_offset": 0, "code_offset": offset,
                          "flags": 0, "origin": "symbol"})

    ida_auto.auto_wait()
    bitness_fixed = force_16bit()
    made, failed = seed(record["entries"])
    ida_auto.auto_wait()
    extra_made, extra_failed = seed(extra)
    ida_auto.auto_wait()
    made.extend(extra_made)
    failed.extend(extra_failed)
    swept = sweep_unreached() if do_sweep else []

    payload = export_module.build_payload()
    payload["overlay"] = {
        "platform": manifest.get("platform", ""),
        "module": module,
        "index": record["index"],
        "file_offset": record["file_offset"],
        "code_size": record["code_size"],
        "code_sha256": record["code_sha256"],
        "entry_count": record["entry_count"],
        "relocation_offsets": record["relocation_offsets"],
        "seeded_entries": [
            {"index": e["index"], "stub_offset": e["stub_offset"],
             "code_offset": e["code_offset"], "flags": e["flags"]}
            for e in made
        ],
        "failed_entries": failed,
        "swept_functions": swept,
        "bitness_fixed_segments": bitness_fixed,
    }

    os.makedirs(os.path.dirname(out_path) or ".", exist_ok=True)
    with open(out_path, "w", encoding="utf-8") as handle:
        json.dump(payload, handle, ensure_ascii=False)
    return 0


try:
    rc = main()
except BaseException:
    import traceback
    log_path = (sys.argv[3] if len(sys.argv) > 3 else "/work/overlay") + ".error.log"
    try:
        os.makedirs(os.path.dirname(log_path) or ".", exist_ok=True)
        with open(log_path, "w", encoding="utf-8") as handle:
            traceback.print_exc(file=handle)
    except BaseException:
        pass
    rc = 3
ida_pro.qexit(rc)
