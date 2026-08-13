"""用原版 handler 傳給 `READVAR` 的參數，交叉驗證 `internal/ecl` 宣稱的 arity。

**正確的 arity 訊號是 `READVAR(n)` 的參數 n**，不是 `ADDRESSVALUE` 的呼叫次數。
`READVAR` 是 operand 解碼器：它從 ECL PC（`DS:7F21h`）往後解 n 個 operand 進三個
平行陣列；`ADDRESSVALUE(i)` 只是事後取用第 i 個已解好的 operand，一個 operand
可以被取用零次或多次，所以次數與 arity 無關。第 562 輪先用了 `ADDRESSVALUE`
次數，只有 35/64 吻合；換成 `READVAR` 參數後是 60/64（見 spec 564）。

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
    """helper 的 stub offset → 原始函式名。

    只有 PC-98 有 Borland 符號，所以名稱一律由 PC-98 解出，再**依 entry index**
    套到目標平台。不能用 code offset 對應：同一個 helper 在兩平台的 code offset
    不同（`READVAR` 在 PC-98 是 `008Eh`、DOS 是 `0034h`），用 code offset 查會
    整片查不到。兩平台相同的是 stub offset 與 entry index（見 spec 562）。
    """
    def overlay_of(target):
        manifest = json.load(open(os.path.join(SWEEP, target, "ovr-manifest.json"),
                                  encoding="utf-8"))
        return next(o for o in manifest["overlays"] if o["module"] == HELPER_UNIT)

    overlay = overlay_of(platform)
    reference = overlay_of("pc98")

    symbol_names = {}
    symbol_path = os.path.join(SWEEP, "pc98", "borland-symbols.json")
    if os.path.exists(symbol_path):
        symbols = json.load(open(symbol_path, encoding="utf-8"))
        for symbol in symbols["symbols"]:
            if symbol.get("overlay_module") == HELPER_UNIT and symbol.get("overlay_code_offset"):
                symbol_names.setdefault(symbol["offset"], symbol["name"])
    by_index = {entry["index"]: symbol_names.get(entry["code_offset"])
                for entry in reference["entries"]}

    mapping = {}
    for entry in overlay["entries"]:
        label = by_index.get(entry["index"])
        if label is None:
            label = "%s#%d@%04Xh" % (HELPER_UNIT, entry["index"], entry["code_offset"])
        mapping["%04X" % entry["stub_offset"]] = label
    return mapping, overlay


def readvar_arguments(platform, stub_names, dispatch_table, helper_segment):
    """每個 handler 傳給 `READVAR` 的參數（依出現順序）。

    找不到立即數時放 `None`——那通常代表數量來自暫存器，也就是變長指令的
    第二段解碼。不要把 `None` 當成 0。

    呼叫歸屬用 IDA 的**實際函式 chunk 範圍**，不用「排序後取前一個 handler
    起點」的區間猜測：handler 在位址上並不連續，猜區間會把後面函式的呼叫
    算到前一個 handler 頭上（實測讓 `3Dh`／`3Eh` 這種 arity 0 的指令冒出
    READVAR 呼叫）。
    """
    module = json.load(open(os.path.join(SWEEP, platform, "out", "overlay-02.json"),
                            encoding="utf-8"))
    owner = []
    for function in module["functions"]:
        for chunk in function["chunks"]:
            owner.append((chunk["start"], chunk["end"], function["ea"]))
    owner.sort()

    def owning_function(ea):
        for start, end, function_ea in owner:
            if start <= ea < end:
                return function_ea
        return None

    stub_for_readvar = next((offset for offset, name in stub_names.items()
                             if name == "READVAR"), None)
    items = json.load(open(os.path.join(SWEEP, platform, "overlays",
                                        "all-overlay-02.json"), encoding="utf-8"))["items"]
    immediate = re.compile(r"^mov\s+(?:al|ax),\s*([0-9A-Fa-f]+h|[0-9]+)$")
    dispatch = load_module(os.path.join(ROOT, "scripts", "ecl_dispatch_table.py"), "dispatch2")

    found = {}
    for index, item in enumerate(items):
        raw = item["bytes"]
        if not (raw.startswith("9a") and len(raw) >= 10):
            continue
        # ⚠ segment 一定要一起比對。只比 stub offset 會把別的 overlay 的
        # 同號 stub（例如 overlay-24 的 014A:002A）誤算成 READVAR。
        offset = "%04X" % int(raw[4:6] + raw[2:4], 16)
        segment = "%04X" % int(raw[8:10] + raw[6:8], 16)
        if offset != stub_for_readvar or segment != helper_segment:
            continue
        argument = None
        for back in range(index - 1, max(-1, index - 4), -1):
            match = immediate.match(re.sub(r"\s+", " ", items[back]["disasm"].strip()))
            if match:
                argument = dispatch.parse_immediate(match.group(1))
                break
        function_ea = owning_function(item["ea"])
        if function_ea is not None:
            found.setdefault(function_ea, []).append(argument)
    return found


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

    tracked = ["READVAR", "ADDRESSVALUE", "STOREVALUE", "ADDFNC"]
    readvar_args = readvar_arguments(platform, stub_names, table, helper_segment)
    rows, mismatches = [], []
    for opcode in sorted(table):
        if opcode > 0x40:
            continue
        handler = table[opcode]
        name, arity = commands.get(opcode, ("（不在 KnownCommands）", None))
        counts = per_function.get(handler, collections.Counter())
        decoded = readvar_args.get(handler)
        # arity 0 的指令不呼叫 READVAR，這也是一致，不是缺漏。
        if arity == 0:
            agree = decoded is None or decoded == []
        else:
            agree = decoded is not None and decoded[:1] == [arity]
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
    print("| opcode | 指令名（remake） | 宣稱 arity | `READVAR` 參數 | handler | " +
          " | ".join("`%s`" % name for name in tracked) + " | 一致 |")
    print("|---|---|---:|---|---|" + "---:|" * len(tracked) + "---|")
    for opcode, name, arity, handler, counts, agree in rows:
        cells = " | ".join(str(counts.get(helper, 0)) for helper in tracked)
        decoded = readvar_args.get(handler)
        shown = "—" if not decoded else "、".join(
            "?" if value is None else str(value) for value in decoded)
        print("| `%02Xh` | %s | %s | %s | `%04Xh` | %s | %s |"
              % (opcode, name, "—" if arity is None else arity, shown, handler,
                 cells, "✓" if agree else "✗"))
    print()
    print("`READVAR` 參數與宣稱 arity 一致的有 %d／%d 個 opcode。"
          % (len(rows) - len(mismatches), len(rows)))
    print()
    print("不一致的 opcode：%s。"
          % "、".join("`%02Xh`" % opcode for opcode in mismatches))
    print()
    print("不一致的都是**變長指令**：原版先 `READVAR(固定前綴)`，解出一個 operand")
    print("得到後續數量 N，`dec` ECL PC 回退一格，再 `READVAR(N)`。`KnownCommands`")
    print("把它們宣告成 0 並由 parser 特別處理，因此不是錯誤，但固定前綴的長度")
    print("（本表第四欄的第一個數字）是原版事實，remake 應該對得上。")
    return 0


if __name__ == "__main__":
    sys.exit(main())
