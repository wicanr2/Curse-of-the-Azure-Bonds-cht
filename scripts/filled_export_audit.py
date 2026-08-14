"""比對「補洞後」與原本的 prologue 匯出，找出**函式本體真的少讀了**的那些。

`tools/ida/fill_gaps_and_export.py` 會把 undefined 的位元組強制解成指令。這一步
補回了真正漏掉的程式碼，但也會把**函式尾巴後面的字串常數**解成一堆無意義的
指令——`pc98/overlay-22` 補完的 70 支裡有 61 支屬於後者。

所以補洞的輸出不能整份照單全收。判準：

- 新增的指令若**全部落在原本最後一條 `ret` 之後**，那是尾巴的資料被誤解，
  對函式本體的判讀沒有影響。
- 只要**有任何一條新增指令落在最後一條 `ret` 之前**，就代表原本的匯出把本體
  截斷了——那一支先前若已判為「已解讀」，判讀就是根據殘缺的指令序列做的，
  必須重讀。

輸出兩份清單並回報數量；`--list-suspect` 只印需要重讀的那些。

用法：
    python3 scripts/filled_export_audit.py [--list-suspect]
"""

import glob
import json
import os
import sys

ROOT = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
SWEEP = os.path.join(ROOT, "workplace", "re-sweep")
LEDGER = os.path.join(ROOT, "docs", "audit", "re-function-ledger.json")
RET = ("retf", "retn", "ret")


def last_ret(function):
    out = -1
    for item in function["items"]:
        if item["disasm"].strip().split(" ")[0] in RET:
            out = max(out, item["ea"])
    return out


def main():
    only_suspect = "--list-suspect" in sys.argv
    ledger = {(e["platform"], e["module"], e["ea"]): e
              for e in json.load(open(LEDGER, encoding="utf-8"))["functions"]}

    truncated, tail_only, missing = [], 0, 0
    for platform in ("dos", "pc98"):
        pattern = os.path.join(SWEEP, platform, "overlays", "filled", "%s-*.json" % platform)
        for path in sorted(glob.glob(pattern)):
            module = os.path.basename(path)[len(platform) + 1:-5]
            old_path = os.path.join(SWEEP, platform, "overlays", "prologue",
                                    "%s-%s.json" % (platform, module))
            if not os.path.exists(old_path):
                missing += 1
                continue
            old = {f["ea"]: f for f in json.load(open(old_path, encoding="utf-8"))["functions"]}
            new = json.load(open(path, encoding="utf-8"))["functions"]
            for function in new:
                before = old.get(function["ea"])
                if before is None:
                    continue
                seen = {item["ea"] for item in before["items"]}
                added = [item for item in function["items"] if item["ea"] not in seen]
                if not added:
                    continue
                boundary = last_ret(before)
                inside = [item for item in added if item["ea"] < boundary]
                if not inside:
                    tail_only += 1
                    continue
                entry = ledger.get((platform, module, function["ea"]))
                truncated.append({
                    "platform": platform, "module": module, "ea": function["ea"],
                    "old": len(before["items"]), "new": len(function["items"]),
                    "inside": len(inside),
                    "state": entry["state"] if entry else "—",
                    "spec": entry["spec"] if entry else "",
                })

    truncated.sort(key=lambda row: (row["state"] != "已解讀", -row["inside"]))
    if not only_suspect:
        print("尾巴資料被誤解（不影響本體判讀）：%d 支" % tail_only)
        print("缺少 prologue 對照：%d 個模組" % missing)
    print("本體被截斷：%d 支，其中已解讀 %d 支需重讀"
          % (len(truncated), sum(1 for r in truncated if r["state"] == "已解讀")))
    for row in truncated:
        if only_suspect and row["state"] != "已解讀":
            continue
        print("  %-5s %-14s %05Xh  %d → %d（本體內 +%d）  %s  %s"
              % (row["platform"], row["module"], row["ea"], row["old"], row["new"],
                 row["inside"], row["state"], os.path.basename(row["spec"])))
    return 0


if __name__ == "__main__":
    sys.exit(main())
