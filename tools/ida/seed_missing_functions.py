"""把 IDA 漏認的函式補進資料庫，並回報每個位址失敗的原因。

為什麼需要：DOS 側有幾個 overlay（`11`／`25`／`04`／`15`／`17`）的指令覆蓋率
只有 25%～68%，而且 **overlay control block 宣告的 entry code_offset 有一批
沒有成為函式**——`overlay-15` 連 `0000h`（unit init）都不是。台帳的分母是
「IDA 認出的函式數」，這些漏認等於分母本身不完整。

失敗原因要分開記，否則只看到「跳過 8 個」無從判斷是已經有了、還是解碼不出來。

用法：
    tools/ida.sh py <module>.i64 seed_missing_functions.py /work/seeds.json /work/report.json
"""

import json
import os
import sys

import ida_auto
import ida_bytes
import ida_funcs
import ida_pro
import ida_ua
import idautils
import ida_segment
import ida_loader


def describe(ea):
    flags = ida_bytes.get_flags(ea)
    func = ida_funcs.get_func(ea)
    return {
        "ea": ea,
        "is_code": bool(ida_bytes.is_code(flags)),
        "is_data": bool(ida_bytes.is_data(flags)),
        "is_unknown": bool(ida_bytes.is_unknown(flags)),
        "in_func": None if func is None else func.start_ea,
        "bytes": (ida_bytes.get_bytes(ea, 6) or b"").hex(),
    }


def next_boundary(ea, seeds):
    """推不出邊界時的退路：下一個 seed、下一個既有函式起點、或段尾，取最近的。"""
    candidates = [s for s in seeds if s > ea]
    candidates += [f for f in idautils.Functions() if f > ea]
    segment = ida_segment.getseg(ea)
    if segment is not None:
        candidates.append(segment.end_ea)
    return min(candidates) if candidates else ida_bytes.BADADDR


def main():
    if len(sys.argv) < 3:
        return 2
    seeds = json.load(open(sys.argv[1], encoding="utf-8"))
    report_path = sys.argv[2]
    ida_auto.auto_wait()

    added, failed = [], []
    for ea in seeds:
        before = describe(ea)
        if before["in_func"] == ea:
            failed.append(dict(before, reason="已經是函式起點"))
            continue
        # DELIT_EXPAND 會把整個被誤判的資料項一起清掉；只刪 3 bytes 不夠。
        ida_bytes.del_items(ea, ida_bytes.DELIT_EXPAND | ida_bytes.DELIT_NOTRUNC, 1)
        if not ida_ua.create_insn(ea):
            failed.append(dict(before, reason="create_insn 失敗（不是合法指令）"))
            continue
        # 單參數的 add_func 在這批位址上一律失敗——IDA 決定不了函式結尾
        # （這些 overlay 的自動分析本來就沒走到這裡）。先用 IDA 自己的邊界
        # 推導求出結尾，推不出來才退回「下一個已知邊界」。
        bounds = ida_funcs.func_t(ea)
        if ida_funcs.find_func_bounds(bounds, ida_funcs.FIND_FUNC_DEFINE) == \
                ida_funcs.FIND_FUNC_OK and bounds.end_ea > ea:
            ok = ida_funcs.add_func(ea, bounds.end_ea)
            how = "find_func_bounds"
        else:
            ok = ida_funcs.add_func(ea, next_boundary(ea, seeds))
            how = "下一個已知邊界"
        if not ok:
            failed.append(dict(before, reason="add_func 失敗（%s）" % how))
            continue
        added.append({"ea": ea, "how": how,
                      "end": ida_funcs.get_func(ea).end_ea})
    ida_auto.auto_wait()

    # ⚠ ida_pro.qexit() 不會儲存資料庫。少了這一行，補上的函式只存在於這次
    # 執行的記憶體裡，.i64 完全沒變——重新匯出時看到的還是舊的函式清單。
    saved = ida_loader.save_database(
        ida_loader.get_path(ida_loader.PATH_TYPE_IDB), 0)

    os.makedirs(os.path.dirname(report_path) or ".", exist_ok=True)
    json.dump({"added": added, "failed": failed, "saved": bool(saved),
               "total_functions": len(list(idautils.Functions()))},
              open(report_path, "w", encoding="utf-8"), indent=1)
    return 0


if __name__ == "__main__":
    try:
        rc = main()
    except BaseException:
        import traceback
        path = (sys.argv[2] if len(sys.argv) > 2 else "/work/seed") + ".error.log"
        with open(path, "w", encoding="utf-8") as handle:
            traceback.print_exc(file=handle)
        rc = 3
    ida_pro.qexit(rc)
