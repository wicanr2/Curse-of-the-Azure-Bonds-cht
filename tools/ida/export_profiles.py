"""為 module 內每個函式產生「證據剖面」，讓後續解讀不必每次重開 IDA。

剖面只收 IDA 已知的客觀事實，不做語意判斷：

- 呼叫關係：near call 目標、far call 的 `seg:off`（後續可解回 overlay entry）
- 資料接觸：讀／寫的具名或無名位址、字串參照
- 指令特徵：助憶碼直方圖、`int` 號碼、port I/O、迴圈（往回跳）數
- 介面：stack frame 大小、`retn N` 的參數位元組、chunk 範圍

用途是把「這個函式在做什麼」的第一輪判斷變成查表：例如只呼叫 RTL 又沒有
資料寫入的多半是格式化輔助；有 `out` 指令的是硬體存取；只有一條 far call
又立刻返回的是 thunk。**這些都是候選判斷，不是結論。**

用法：
    tools/ida.sh py <module>.i64 export_profiles.py /work/profiles.json
"""

import collections
import json
import os
import sys

import ida_auto
import ida_bytes
import ida_funcs
import ida_name
import ida_pro
import ida_ua
import idautils
import idc


def far_target(raw):
    if len(raw) >= 5 and raw[0] == 0x9A:
        return "%04X:%04X" % (raw[3] | (raw[4] << 8), raw[1] | (raw[2] << 8))
    return None


def profile(func_ea):
    func = ida_funcs.get_func(func_ea)
    chunks = [(c[0], c[1]) for c in idautils.Chunks(func_ea)]

    mnemonics = collections.Counter()
    far_calls = collections.Counter()
    near_calls = collections.Counter()
    interrupts = collections.Counter()
    ports = collections.Counter()
    data_reads, data_writes, strings = set(), set(), []
    backward_jumps = 0
    size = 0

    for start, end in chunks:
        ea = start
        while ea < end:
            item = max(1, ida_bytes.get_item_size(ea))
            size += item
            mnemonic = (idc.print_insn_mnem(ea) or "").lower()
            if mnemonic:
                mnemonics[mnemonic] += 1
            raw = ida_bytes.get_bytes(ea, item) or b""

            if mnemonic.startswith("call"):
                target = far_target(raw)
                if target:
                    far_calls[target] += 1
                for ref in idautils.CodeRefsFrom(ea, 0):
                    near_calls["%04X" % ref] += 1
            elif mnemonic == "int":
                interrupts["%02X" % idc.get_operand_value(ea, 0)] += 1
            elif mnemonic in ("in", "out"):
                index = 0 if mnemonic == "out" else 1
                ports["%04X" % idc.get_operand_value(ea, index)] += 1
            elif mnemonic.startswith("j"):
                for ref in idautils.CodeRefsFrom(ea, 0):
                    if ref < ea:
                        backward_jumps += 1

            for ref in idautils.DataRefsFrom(ea):
                flags = ida_bytes.get_full_flags(ref)
                if ida_bytes.is_strlit(flags):
                    text = idc.get_strlit_contents(ref)
                    if text:
                        strings.append(text.decode("latin-1", "replace")[:80])
                # 讀寫分類交給 xref 型別，不猜指令語意
                for xref in idautils.XrefsTo(ref):
                    if xref.frm != ea:
                        continue
                    if xref.type == 2:      # dr_W
                        data_writes.add(ref)
                    else:
                        data_reads.add(ref)
            ea = idc.next_head(ea, end) if idc.next_head(ea, end) > ea else ea + item

    return {
        "ea": func_ea,
        "name": ida_name.get_name(func_ea),
        "size": size,
        "chunks": [{"start": s, "end": e} for s, e in chunks],
        "frame_size": idc.get_frame_size(func_ea),
        "arg_bytes": idc.get_frame_args_size(func_ea),
        "is_lib": bool(func.flags & ida_funcs.FUNC_LIB),
        "callers": sorted({x for x in idautils.CodeRefsTo(func_ea, 1)}),
        "far_calls": dict(far_calls),
        "near_calls": dict(near_calls),
        "interrupts": dict(interrupts),
        "ports": dict(ports),
        "mnemonics": dict(mnemonics.most_common(12)),
        "backward_jumps": backward_jumps,
        "data_reads": sorted(data_reads),
        "data_writes": sorted(data_writes),
        "strings": strings[:12],
    }


def main():
    if len(sys.argv) < 2:
        return 2
    out_path = sys.argv[1]
    ida_auto.auto_wait()
    payload = {
        "schema": "coab-function-profiles/1",
        "profiles": [profile(ea) for ea in idautils.Functions()],
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
        log_path = (sys.argv[1] if len(sys.argv) > 1 else "/work/profiles") + ".error.log"
        try:
            os.makedirs(os.path.dirname(log_path) or ".", exist_ok=True)
            with open(log_path, "w", encoding="utf-8") as handle:
                traceback.print_exc(file=handle)
        except BaseException:
            pass
        rc = 3
    ida_pro.qexit(rc)
