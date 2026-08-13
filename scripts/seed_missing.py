"""算出每個模組「IDA 漏認的函式起點」，並驅動 IDA 補上。

判準：位址沒有被任何既有函式涵蓋，且開頭是 `55 89 e5`
（`push bp / mov bp,sp`）——Turbo Pascal 的標準 prologue。不猜其他形式。

用法：
    python3 scripts/seed_missing.py            # 只列出
    python3 scripts/seed_missing.py --apply    # 補進 .i64 並重新匯出
"""
import json, os, re, subprocess, sys
sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
from coverage_audit import covered, code_sizes, SWEEP, ROOT

PROLOGUE = re.compile(rb"\x55\x89\xe5")


def entry_offsets(platform, module):
    """overlay control block 宣告的 entry code_offset。

    `FFFFh` 是**未使用的 slot**，不是位址——控制區塊的 entry 陣列長度固定，
    沒用到的填 `FFFFh`。當成位址去 seed 會產生假的失敗紀錄。
    """
    manifest = json.load(open(os.path.join(SWEEP, platform, "ovr-manifest.json"),
                              encoding="utf-8"))
    overlay = next(o for o in manifest["overlays"] if o["module"] == module)
    return [e["code_offset"] for e in overlay["entries"] if e["code_offset"] != 0xFFFF]


def missing(platform, module, size):
    full = os.path.join(SWEEP, platform, "overlays", "full", "%s-%s.json" % (platform, module))
    if not os.path.exists(full):
        return []
    seen = covered(full)
    blob = open(os.path.join(SWEEP, platform, "overlays", module + ".bin"), "rb").read()
    found = {f["ea"] for f in json.load(open(full, encoding="utf-8"))["functions"]}
    out = [m.start() for m in PROLOGUE.finditer(blob)
           if m.start() < size and m.start() not in seen]
    # entry stub 宣告的入口一定是函式；沒成為函式就是分析漏了。
    out += [ea for ea in entry_offsets(platform, module)
            if ea < size and ea not in found and ea not in out]
    return sorted(set(out))


def main():
    apply_it = "--apply" in sys.argv
    total = 0
    for platform in ("dos", "pc98"):
        sizes = code_sizes(platform)
        directory = os.path.join(SWEEP, platform, "overlays")
        for module, size in sorted(sizes.items()):
            seeds = missing(platform, module, size)
            if not seeds:
                continue
            total += len(seeds)
            print("%-5s %-12s %d 個：%s" % (platform, module, len(seeds),
                                          " ".join("%04Xh" % s for s in seeds)))
            if not apply_it:
                continue
            path = os.path.join(directory, "%s.seeds.json" % module)
            json.dump(seeds, open(path, "w", encoding="utf-8"))
            subprocess.run([os.path.join(ROOT, "tools", "ida.sh"), "py",
                            os.path.join(directory, "%s.bin.i64" % module),
                            "seed_missing_functions.py",
                            "/work/%s.seeds.json" % module,
                            "/work/%s.seedreport.json" % module],
                           capture_output=True)
            report = os.path.join(directory, "%s.seedreport.json" % module)
            if os.path.exists(report):
                data = json.load(open(report, encoding="utf-8"))
                reasons = {}
                for item in data["failed"]:
                    reasons[item["reason"]] = reasons.get(item["reason"], 0) + 1
                print("      補上 %d，函式總數 %d%s"
                      % (len(data["added"]), data["total_functions"],
                         ("；" + "、".join("%s×%d" % kv for kv in reasons.items()))
                         if reasons else ""))
            else:
                print("      !! 沒有產生報告")
    print("\n合計 %d 個" % total)
    return 0


if __name__ == "__main__":
    sys.exit(main())
