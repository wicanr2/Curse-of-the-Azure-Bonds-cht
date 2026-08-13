"""印出指定函式的完整逐指令 body（從 full 匯出讀，不重開 IDA）。

用法：python3 scripts/show.py <platform> <module> <ea16> [ea16...]
"""
import json, os, sys
ROOT = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
SWEEP = os.path.join(ROOT, "workplace", "re-sweep")

def load(platform, module):
    for base in (os.path.join(SWEEP, platform, "overlays", "full"),
                 os.path.join(SWEEP, platform, "full")):
        p = os.path.join(base, "%s-%s.json" % (platform, module))
        if os.path.exists(p):
            return {f["ea"]: f for f in json.load(open(p, encoding="utf-8"))["functions"]}
    raise SystemExit("找不到 full 匯出：%s %s" % (platform, module))

def show_range(platform, module, start, end):
    """把落在 [start, end) 的所有指令依位址接回來。

    IDA 的函式邊界不可信：Turbo Pascal 的共用 epilogue（`mov sp,bp / pop bp /
    retf N`）與被 `jmp` 跳過的中段，常被切成獨立「函式」。要讀完整的一支，
    唯一可靠的做法是走位址範圍。
    """
    fns = load(platform, module)
    items = [it for f in fns.values() for it in f["items"] if start <= it["ea"] < end]
    items.sort(key=lambda it: it["ea"])
    print("=== %s %s %04Xh..%04Xh  共 %d 條指令" % (platform, module, start, end, len(items)))
    previous = None
    for it in items:
        if previous is not None and it["ea"] != previous:
            print("  ---- %04Xh..%04Xh 沒有匯出（IDA 未認成指令）----" % (previous, it["ea"]))
        print("  %04X  %-16s %s" % (it["ea"], it["bytes"], it["disasm"]))
        previous = it["ea"] + len(it["bytes"]) // 2


def main():
    if sys.argv[1] == "--range":
        show_range(sys.argv[2], sys.argv[3], int(sys.argv[4], 16), int(sys.argv[5], 16))
        return
    platform, module = sys.argv[1], sys.argv[2]
    fns = load(platform, module)
    for arg in sys.argv[3:]:
        ea = int(arg, 16)
        f = fns.get(ea)
        if f is None:
            print("!! %s %s %04Xh 不在匯出裡" % (platform, module, ea)); continue
        print("=== %s %s %04Xh %s  size=%d args=%d callers=%d"
              % (platform, module, ea, f["name"], f["size"], f["arg_bytes"], len(f["callers"])))
        for it in f["items"]:
            print("  %04X  %-16s %s" % (it["ea"], it["bytes"], it["disasm"]))
        print()

if __name__ == "__main__":
    main()

# --- 位址範圍模式 ---------------------------------------------------------
# IDA 的函式邊界不可信：Turbo Pascal 的共用 epilogue 與被跳過的中段常被切成
# 獨立「函式」。要讀完整的一支，用位址範圍把所有落在區間內的指令接回來。
#   python3 scripts/show.py --range <platform> <module> <start16> <end16>
