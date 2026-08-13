"""用原版 handler 的 helper 呼叫次數，交叉驗證 `internal/ecl` 宣稱的 operand arity。

Turbo Pascal 的跨 unit 呼叫是 far call，目標是被呼叫 overlay 的 **control block
stub offset**，不是 code offset。本腳本把它一路解回原始函式名：

    INTERPET handler 的 far call `0062:0025`
      → overlay-07（ECL2）control stub offset 0025h
      → entry index 1 → code offset 0296h
      → PC-98 Borland 符號 `ADDRESSVALUE`

因此「這個 opcode 的 handler 呼叫了 ADDRESSVALUE 幾次」是可以直接數出來的，
而它應該與 command table 宣稱的 operand 數量一致。不一致就是待查項目——
可能是 remake 的 arity 錯了，也可能是該 handler 用別的方式取 operand。

**本腳本只產生比對表，不自動判定誰對。**

用法：
    python3 scripts/ecl_handler_operand_audit.py <平台> > docs/audit/ecl-handler-operand-audit.md
"""

import collections
import importlib.util
import json
import os
import re
import sys

ROOT = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
SWEEP = os.path.join(ROOT, "workplace", "re-sweep")
HELPER_UNIT = "overlay-07"   # ECL2；INTERPET 的 helper 全在這裡


def load_module(path, name):
    spec = importlib.util.spec_from_file_location(name, path)
    module = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(module)
    return module


def known_commands():
    source = open(os.path.join(ROOT, "internal", "ecl", "operand.go"),
                  encoding="utf-8").read()
    start = source.index("KnownCommands = map[byte]Command{")
    block = source[start:source.index("\n}\n", start)]
    table = {}
    for match in re.finditer(
            r'0x([0-9A-Fa-f]{2}): \{0x[0-9A-Fa-f]{2}, "([^"]+)", (\d+)\}', block):
        table[int(match.group(1), 16)] = (match.group(2), int(match.group(3)))
    return table


def helper_names(platform):
    """helper 的 stub offset → 原始函式名（只有 PC-98 有符號）。"""
    manifest = json.load(open(os.path.join(SWEEP, platform, "ovr-manifest.json"),
                              encoding="utf-8"))
    overlay = next(o for o in manifest["overlays"] if o["module"] == HELPER_UNIT)
    names = {}
    symbol_path = os.path.join(SWEEP, "pc98", "borland-symbols.json")
    if os.path.exists(symbol_path):
        symbols = json.load(open(symbol_path, encoding="utf-8"))
        for symbol in symbols["symbols"]:
            if symbol.get("overlay_module") == HELPER_UNIT and symbol.get("overlay_code_offset"):
                names.setdefault(symbol["offset"], symbol["name"])
    mapping = {}
    for entry in overlay["entries"]:
        label = names.get(entry["code_offset"])
        if label is None:
            label = "%s#%d@%04Xh" % (HELPER_UNIT, entry["index"], entry["code_offset"])
        mapping["%04X" % entry["stub_offset"]] = label
    return mapping, overlay


def main():
    platform = sys.argv[1] if len(sys.argv) > 1 else "pc98"
    dispatch = load_module(os.path.join(ROOT, "scripts", "ecl_dispatch_table.py"), "dispatch")

    commands = known_commands()
    stub_names, helper_overlay = helper_names(platform)
    calls = json.load(open(os.path.join(SWEEP, platform, "overlays",
                                        "calls-overlay-02.json"), encoding="utf-8"))["calls"]

    # helper segment：INTERPET 對 ECL2 的 far call 全部指向同一個 segment。
    segments = collections.Counter(
        call["far_target"].split(":")[0] for call in calls if call["far_target"])
    helper_segment = segments.most_common(1)[0][0]

    per_function = collections.defaultdict(collections.Counter)
    for call in calls:
        target = call["far_target"]
        if target and target.startswith(helper_segment + ":"):
            per_function[call["function"]][stub_names.get(target.split(":")[1], target)] += 1

    data, table, unhandled, source_address, problems = dispatch.solve(
        os.path.join(SWEEP, platform, "overlays", "dispatch.json"))

    tracked = ["ADDRESSVALUE", "READVAR", "STOREVALUE", "ADDFNC"]
    rows, mismatches = [], []
    for opcode in sorted(table):
        if opcode > 0x40:
            continue
        handler = table[opcode]
        name, arity = commands.get(opcode, ("（不在 KnownCommands）", None))
        counts = per_function.get(handler, collections.Counter())
        fetched = counts.get("ADDRESSVALUE", 0)
        agree = arity is not None and fetched == arity
        rows.append((opcode, name, arity, handler, counts, agree))
        if arity is not None and not agree:
            mismatches.append(opcode)

    print("# ECL handler 的 operand 取用 vs 宣稱 arity")
    print()
    print("由 `scripts/ecl_handler_operand_audit.py` 產生，不要手改。")
    print()
    print("平台 `%s`：dispatcher `%04Xh`，opcode 來源 `ds:%sh`，"
          "helper segment `%sh`（＝%s，ECL2 單元的 control block）。"
          % (platform, data["start"], source_address, helper_segment, HELPER_UNIT))
    print()
    print("`ADDRESSVALUE` 等名稱來自 PC-98 的 Borland 除錯表，經 stub offset →")
    print("entry index → code offset 解析而得，不是猜的。DOS 沒有符號表，兩平台")
    print("的 stub 佈局相同才能沿用同一組名稱——這一點本身是 `strong inference`。")
    print()
    print("| opcode | 指令名（remake） | 宣稱 arity | handler | " +
          " | ".join("`%s`" % name for name in tracked) + " | 一致 |")
    print("|---|---|---:|---|" + "---:|" * len(tracked) + "---|")
    for opcode, name, arity, handler, counts, agree in rows:
        cells = " | ".join(str(counts.get(helper, 0)) for helper in tracked)
        print("| `%02Xh` | %s | %s | `%04Xh` | %s | %s |"
              % (opcode, name, "—" if arity is None else arity, handler, cells,
                 "✓" if agree else "✗"))
    print()
    print("`ADDRESSVALUE` 次數與宣稱 arity 一致的有 %d／%d 個 opcode。"
          % (len(rows) - len(mismatches), len(rows)))
    print()
    print("不一致的 opcode：%s。"
          % "、".join("`%02Xh`" % opcode for opcode in mismatches))
    print()
    print("不一致**不代表 remake 的 arity 錯**：算術類指令（`ADD` 一族）走的是")
    print("`ADDFNC`＋`STOREVALUE` 的組合，menu 類走 `READVAR`，都不是每個 operand")
    print("各呼叫一次 `ADDRESSVALUE`。每一筆都要逐一讀 handler 才能定案，這份表")
    print("只負責把待查清單列出來並排序。")
    return 0


if __name__ == "__main__":
    sys.exit(main())
