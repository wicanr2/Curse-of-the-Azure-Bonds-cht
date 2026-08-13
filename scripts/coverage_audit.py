"""量測「IDA 到底把多少 code 位元組認成指令」，作為台帳分母的可信度檢查。

為什麼要量：`LOSEDUDE`（[spec 582]）裡有兩處位元組 IDA 完全沒認成指令，
所以不在任何匯出裡——靠函式匯出讀會**安靜地漏掉指令**。缺口有多大，決定
「已解讀」這個狀態能宣稱到什麼程度。

計算方式：把每個模組所有函式的每一條指令攤成**位元組集合**再取聯集。
不能直接把各函式的長度相加——Turbo Pascal 的共用 epilogue 會同時屬於好幾支，
相加會超過 100%（`dos/overlay-09` 就是這樣冒出 101.3% 的）。

未涵蓋的位元組**不全是漏認的指令**：Pascal 的 string 常數與 32 bytes 的 set
常數本來就在 code 段裡，是資料。所以這裡量的是上界，要判斷某一段究竟是資料
還是漏認的指令，只能個別看。

用法：python3 scripts/coverage_audit.py [--gaps <platform> <module>]
"""

import glob
import json
import os
import sys

ROOT = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
SWEEP = os.path.join(ROOT, "workplace", "re-sweep")


def covered(path):
    """該模組被認成指令的位元組集合。"""
    out = set()
    for function in json.load(open(path, encoding="utf-8"))["functions"]:
        for item in function["items"]:
            size = len(item["bytes"]) // 2
            out.update(range(item["ea"], item["ea"] + size))
    return out


def code_sizes(platform):
    manifest = json.load(open(os.path.join(SWEEP, platform, "ovr-manifest.json"),
                              encoding="utf-8"))
    return {o["module"]: o["code_size"] for o in manifest["overlays"]}


def gaps(platform, module):
    path = os.path.join(SWEEP, platform, "overlays", "full",
                        "%s-%s.json" % (platform, module))
    seen = covered(path)
    size = code_sizes(platform)[module]
    runs, start = [], None
    for offset in range(size):
        if offset in seen:
            if start is not None:
                runs.append((start, offset)); start = None
        elif start is None:
            start = offset
    if start is not None:
        runs.append((start, size))
    return runs


def main():
    if "--gaps" in sys.argv:
        index = sys.argv.index("--gaps")
        platform, module = sys.argv[index + 1], sys.argv[index + 2]
        blob = open(os.path.join(SWEEP, platform, "overlays", module + ".bin"), "rb").read()
        print("%s %s 未認成指令的區段：" % (platform, module))
        for start, end in gaps(platform, module):
            head = blob[start:min(end, start + 12)].hex(" ")
            print("  %04Xh..%04Xh  %4d bytes  %s%s"
                  % (start, end, end - start, head, "…" if end - start > 12 else ""))
        return 0

    total_code = total_seen = 0
    rows = []
    for platform in ("dos", "pc98"):
        sizes = code_sizes(platform)
        for path in sorted(glob.glob(os.path.join(
                SWEEP, platform, "overlays", "full", "%s-overlay-*.json" % platform))):
            module = os.path.basename(path)[len(platform) + 1:-5]
            size = sizes.get(module, 0)
            if not size:
                continue
            seen = len(covered(path) & set(range(size)))
            total_code += size
            total_seen += seen
            rows.append((platform, module, size, seen, 100.0 * seen / size))

    rows.sort(key=lambda r: r[4])
    print("兩平台 overlay code 合計 %d bytes，認成指令 %d bytes（%.1f%%）"
          % (total_code, total_seen, 100.0 * total_seen / total_code))
    print("\n覆蓋率最低的 12 個模組：")
    print("  平台   模組         code_size  已認指令   覆蓋率")
    for row in rows[:12]:
        print("  %-5s %-12s %8d %9d   %5.1f%%" % row)
    over = [r for r in rows if r[4] > 100.0]
    print("\n超過 100%% 的模組：%d 個（應為 0；集合去重後不該發生）" % len(over))
    return 0


if __name__ == "__main__":
    sys.exit(main())
