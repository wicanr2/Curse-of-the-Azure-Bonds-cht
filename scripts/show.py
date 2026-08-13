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

def next_start(platform, module, ea):
    """下一支函式的起點，用原始位元組找 `55 89 e5` prologue。

    不能用「IDA 認的下一個函式起點」：IDA 會把同一支函式的中段切成獨立函式
    （`overlay-02:0C81h` 的下一個 IDA 函式是 `0CBFh`，但那還在同一支裡）。
    Turbo Pascal 的每一支函式都以 `push bp / mov bp, sp` 開頭，掃位元組比
    問 IDA 可靠。
    """
    blob = open(os.path.join(SWEEP, platform, "overlays", module + ".bin"), "rb").read() \
        if not module.endswith(".EXE") else None
    if blob is None:
        fns = sorted(load(platform, module))
        later = [f for f in fns if f > ea]
        return later[0] if later else ea + 0x1000
    index = blob.find(b"\x55\x89\xe5", ea + 3)
    return index if index > 0 else len(blob)


def show_range(platform, module, start, end):
    """把落在 [start, end) 的所有指令依位址接回來。

    IDA 的函式邊界不可信：Turbo Pascal 的共用 epilogue（`mov sp,bp / pop bp /
    retf N`）與被 `jmp` 跳過的中段，常被切成獨立「函式」。要讀完整的一支，
    唯一可靠的做法是走位址範圍。
    """
    fns = load(platform, module)
    # 同一條指令可能同時屬於好幾個 IDA「函式」（共用出口、被切散的中段），
    # 依位址去重，否則會重複列印並算出負數的缺口。
    seen = {}
    for f in fns.values():
        for it in f["items"]:
            if start <= it["ea"] < end:
                seen.setdefault(it["ea"], it)
    items = [seen[ea] for ea in sorted(seen)]
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
    if sys.argv[1] == "--whole":
        # 讀到「下一個函式起點」為止，而不是相信 IDA 標的 size。
        platform, module = sys.argv[2], sys.argv[3]
        for arg in sys.argv[4:]:
            ea = int(arg, 16)
            show_range(platform, module, ea, next_start(platform, module, ea))
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
