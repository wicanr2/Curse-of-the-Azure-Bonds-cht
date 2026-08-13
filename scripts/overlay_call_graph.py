"""把 overlay 之間的 far call 解回「模組 ＋ entry 編號 ＋ 目標位址」。

反組譯裡的 `9A off seg` far call，IDA 會把 `seg:off` **攤平成線性位址**再去找
名字，於是印出來的是同一個 overlay 內某個位址的標籤——例如
`call sub_1953` 實際上是 `call far 018Ch:0093h`，兩者只是 `seg×16+off` 剛好
相等。**照著這種標籤畫呼叫圖會得到完全錯誤的結果**，本工具存在的理由就是這個。

解法用的是 VROOMM 的版面（spec 562）：每個模組在 EXE 裡有一塊控制區，
`+20h` 起是 stub 陣列，一個 stub 5 bytes（`CD 3F` ＋ 3 bytes 描述子）。所以

    entry 編號 = (off − 20h) / 5      且   off ≥ 20h、(off − 20h) mod 5 = 0

段值則是「控制區在 EXE image 裡的位置」除以 16。基底 H 由 manifest 的
`executable_offset` 反解（讓最多 far call 命中），DOS 得 1968、PC-98 得 2464。

自我檢查有三關：offset 必須落在 stub 陣列上、編號必須小於 `entry_count`、
目標 `code_offset` 不得是 `0FFFFh`（未使用的 entry）。三關都過才算解出；
沒解出的一律歸類為「resident 或未知」，不硬湊。

用法：
    python3 scripts/overlay_call_graph.py            # 兩平台統計 ＋ 寫出 JSON
    python3 scripts/overlay_call_graph.py dos overlay-02 179A   # 查單一函式
"""

import collections
import glob
import json
import os
import sys

ROOT = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
SWEEP = os.path.join(ROOT, "workplace", "re-sweep")
OUT = os.path.join(ROOT, "docs", "audit", "overlay-call-graph.json")
BASE = {"dos": 1968, "pc98": 2464}
STUB_AREA = 0x20
STUB_STRIDE = 5


def modules(platform):
    manifest = json.load(open(os.path.join(SWEEP, platform, "ovr-manifest.json"),
                              encoding="utf-8"))
    out = {}
    for overlay in manifest["overlays"]:
        segment = (overlay["executable_offset"] - BASE[platform]) // 16
        out[segment] = overlay
    return out


def resolve(table, segment, offset):
    overlay = table.get(segment)
    if overlay is None:
        return None
    if offset < STUB_AREA or (offset - STUB_AREA) % STUB_STRIDE:
        return None
    index = (offset - STUB_AREA) // STUB_STRIDE
    entries = {e["index"]: e for e in overlay["entries"]}
    entry = entries.get(index)
    if entry is None or entry["code_offset"] == 0xFFFF:
        return None
    return {"module": overlay["module"], "entry": index,
            "ea": entry["code_offset"]}


def far_calls(path):
    for function in json.load(open(path, encoding="utf-8"))["functions"]:
        for item in function["items"]:
            raw = item["bytes"]
            if raw.startswith("9a") and len(raw) == 10:
                offset = int(raw[2:4], 16) | int(raw[4:6], 16) << 8
                segment = int(raw[6:8], 16) | int(raw[8:10], 16) << 8
                yield function["ea"], item["ea"], segment, offset


def main():
    if len(sys.argv) == 4:
        platform, module, ea = sys.argv[1], sys.argv[2], int(sys.argv[3], 16)
        table = modules(platform)
        path = os.path.join(SWEEP, platform, "overlays", "prologue",
                            "%s-%s.json" % (platform, module))
        seen = []
        for owner, site, segment, offset in far_calls(path):
            if owner != ea:
                continue
            hit = resolve(table, segment, offset)
            label = ("%s entry#%d → %04Xh" % (hit["module"], hit["entry"], hit["ea"])
                     if hit else "resident/未知")
            seen.append("  %04X  call far %04X:%04X   %s"
                        % (site, segment, offset, label))
        print("\n".join(seen) if seen else "（此函式沒有 far call）")
        return 0

    result = {}
    for platform in ("dos", "pc98"):
        table = modules(platform)
        edges = collections.defaultdict(list)
        resolved = unresolved = 0
        for path in sorted(glob.glob(os.path.join(
                SWEEP, platform, "overlays", "prologue", "*.json"))):
            module = os.path.basename(path)[len(platform) + 1:-5]
            for owner, site, segment, offset in far_calls(path):
                hit = resolve(table, segment, offset)
                if hit is None:
                    unresolved += 1
                    continue
                resolved += 1
                key = "%s:%04X" % (module, owner)
                target = "%s:%04X" % (hit["module"], hit["ea"])
                if target not in edges[key]:
                    edges[key].append(target)
        result[platform] = {"edges": dict(edges), "resolved": resolved,
                            "unresolved": unresolved}
        print("%-5s 解出 overlay→overlay 呼叫 %d 條（%d 個呼叫點），"
              "resident／未知 %d" % (platform, sum(len(v) for v in edges.values()),
                                    resolved, unresolved))
    json.dump({"schema": "coab-overlay-call-graph/1", "base": BASE,
               "platforms": result}, open(OUT, "w", encoding="utf-8"),
              ensure_ascii=False, indent=1)
    print("→ %s" % OUT)
    return 0


if __name__ == "__main__":
    sys.exit(main())
