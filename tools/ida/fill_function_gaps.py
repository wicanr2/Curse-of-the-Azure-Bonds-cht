"""把「函式範圍內卻沒被認成指令」的位元組補成指令。

為什麼需要：`overlay-02:149Ch`（opcode `1Eh`）的 body 裡有 49 bytes 與
192 bytes 兩段 IDA 完全沒認出來，`0C81h` 更是漏掉了最關鍵的 opcode 二次判讀
（[spec 587]）。這些缺口不會有任何警告，逐條讀時只會看到「---- 沒有匯出 ----」，
要一段一段手動解碼。

判準（保守，寧可少補）：

1. 位址落在某支函式的 `[start, 下一個 55 89 e5)` 範圍內——用 prologue 定界，
   不用 IDA 的 size（那本來就不可信）。
2. 位址目前是 undefined（不是 code 也不是 data）。
3. **不在已知字串的範圍內**。Pascal 的字串常數就躺在 code 段裡，
   把它們轉成指令會產生假的反組譯。字串清單由呼叫端傳進來。

補完後重跑 `auto_wait`，讓 IDA 把新指令接進既有的流程。

用法：
    tools/ida.sh py <module>.i64 fill_function_gaps.py /work/strings.json /work/report.json
"""

import json
import os
import sys

import ida_auto
import ida_bytes
import ida_funcs
import ida_loader
import ida_pro
import ida_segment
import ida_ua
import idautils


def prologue_bounds(segment):
    """所有 `55 89 e5` 的位址，當作函式邊界。"""
    blob = ida_bytes.get_bytes(segment.start_ea, segment.end_ea - segment.start_ea) or b""
    out, index = [], blob.find(b"\x55\x89\xe5")
    while index >= 0:
        out.append(segment.start_ea + index)
        index = blob.find(b"\x55\x89\xe5", index + 3)
    return out


def main():
    strings = set()
    if len(sys.argv) > 1 and os.path.exists(sys.argv[1]):
        for item in json.load(open(sys.argv[1], encoding="utf-8")):
            strings.update(range(item["offset"], item["offset"] + item["length"] + 1))
    report_path = sys.argv[2] if len(sys.argv) > 2 else "/work/gaps.json"
    ida_auto.auto_wait()

    segment = ida_segment.get_first_seg()
    starts = prologue_bounds(segment)
    ranges = [(starts[i], starts[i + 1] if i + 1 < len(starts) else segment.end_ea)
              for i in range(len(starts))]

    made, refused = 0, 0
    for start, end in ranges:
        ea = start
        while ea < end:
            if ea in strings or not ida_bytes.is_unknown(ida_bytes.get_flags(ea)):
                ea += max(1, ida_bytes.get_item_size(ea))
                continue
            size = ida_ua.create_insn(ea)
            if size:
                made += 1
                ea += size
            else:
                refused += 1
                ea += 1
    ida_auto.auto_wait()

    saved = ida_loader.save_database(
        ida_loader.get_path(ida_loader.PATH_TYPE_IDB), 0)
    json.dump({"made": made, "refused": refused, "saved": bool(saved),
               "functions": len(list(idautils.Functions()))},
              open(report_path, "w", encoding="utf-8"), indent=1)
    return 0


if __name__ == "__main__":
    try:
        rc = main()
    except BaseException:
        import traceback
        path = (sys.argv[2] if len(sys.argv) > 2 else "/work/gaps") + ".error.log"
        with open(path, "w", encoding="utf-8") as handle:
            traceback.print_exc(file=handle)
        rc = 3
    ida_pro.qexit(rc)
