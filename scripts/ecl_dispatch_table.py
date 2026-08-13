"""由 ECL dispatcher 的逐指令 dump 推出 opcode → handler 全表。

做法是對每個 opcode 值 0x00..0xFF 做一次符號執行：把 dispatcher 當成「只對
單一暫存器做比較與分支」的鏈，從函式起點走到第一個 `call`，記下 handler。
遇到會影響結果卻不在支援集合內的指令就**停下並標記**，不猜語意——這是為了
避免用文字比對硬湊出一張看似完整的假表。

輸入是 tools/ida/dump_function.py 的輸出 JSON。

用法：
    python3 scripts/ecl_dispatch_table.py <dump.json>
    python3 scripts/ecl_dispatch_table.py --merge dos=<dump> pc98=<dump> \
        > docs/audit/ecl-opcode-dispatch.md
"""

import json
import re
import sys

# 只支援這幾種：其餘一律視為「需要人工讀」。
CMP_RE = re.compile(r"^cmp\s+(al|ax),\s*([0-9A-Fa-f]+)h?$")
JCC_RE = re.compile(r"^(jz|jnz|je|jne)\s+short\s+\S+$")
JMP_RE = re.compile(r"^jmp\s+(?:short\s+)?\S+$")
CALL_RE = re.compile(r"^call\s+(?:near ptr\s+)?(\S+)$")
LOAD_RE = re.compile(r"^mov\s+(al|ax),\s*ds:([0-9A-Fa-f]+)h$")
IGNORED = ("push", "nop", "xor ah, ah", "mov bp, sp")
# 走到 epilogue 代表這個 opcode 在本 dispatcher 沒有 handler。
EPILOGUE = ("mov sp, bp", "pop bp", "retn", "leave", "ret")


def parse_immediate(text):
    if text.endswith("h"):
        return int(text[:-1], 16)
    # IDA 對純十進位可讀值不加 h，例如 `cmp al, 0`
    if re.fullmatch(r"[0-9]+", text):
        return int(text, 10)
    return int(text, 16)


def load(path):
    with open(path, encoding="utf-8") as handle:
        data = json.load(handle)
    items = {}
    order = []
    for item in data["items"]:
        items[item["ea"]] = item
        order.append(item["ea"])
    return data, items, order


def next_ea(order, ea):
    index = order.index(ea)
    return order[index + 1] if index + 1 < len(order) else None


def branch_target(item):
    """分支目標一律取 IDA 的 code ref，不從助憶碼字串解析標籤名。"""
    refs = [ref for ref in item["code_refs"]]
    return refs[0] if refs else None


def trace(order, items, start, value, source):
    ea = start
    steps = 0
    while ea is not None and steps < 4096:
        steps += 1
        item = items.get(ea)
        if item is None:
            return None, "位址不在 dump 範圍 %04X" % ea
        text = item["disasm"].strip()
        text = re.sub(r"\s*;.*$", "", text)
        text = re.sub(r"\s+", " ", text)

        match = CALL_RE.match(text)
        if match:
            target = branch_target(item)
            return target, None

        match = CMP_RE.match(text)
        if match:
            immediate = parse_immediate(match.group(2))
            flag_zero = (value == immediate)
            ea = next_ea(order, ea)
            item = items.get(ea)
            if item is None:
                return None, "cmp 之後沒有指令"
            text = re.sub(r"\s+", " ", item["disasm"].strip())
            match = JCC_RE.match(text)
            if not match:
                return None, "cmp 之後不是條件跳躍：%s" % text
            taken = flag_zero if match.group(1) in ("jz", "je") else not flag_zero
            ea = branch_target(item) if taken else next_ea(order, ea)
            continue

        if JMP_RE.match(text):
            target = branch_target(item)
            if target is None:
                return None, "jmp 沒有 code ref"
            ea = target
            continue

        # `mov al, ds:XXXXh` 是把 opcode 讀進暫存器，也就是本次符號執行的輸入。
        # 它是唯一被允許改變被追蹤值的指令；出現第二個不同來源就要停下。
        match = LOAD_RE.match(text)
        if match:
            if source["address"] is None:
                source["address"] = match.group(2)
            elif source["address"] != match.group(2):
                return None, "追蹤值有第二個來源：%s" % text
            ea = next_ea(order, ea)
            continue

        if any(text.startswith(prefix) for prefix in EPILOGUE):
            return None, None

        if any(text.startswith(prefix) for prefix in IGNORED) or text in IGNORED:
            ea = next_ea(order, ea)
            continue

        return None, "未支援的指令：%s @%04X" % (text, item["ea"])
    return None, "步數超過上限"


def solve(path):
    """回傳 (opcode → handler, 沒有 handler 的 opcode, opcode 來源位址, 疑難)。"""
    data, items, order = load(path)
    source = {"address": None}
    table, unhandled, problems = {}, [], {}
    for value in range(0x100):
        target, error = trace(order, items, data["start"], value, source)
        if error:
            problems[value] = error
        elif target is None:
            unhandled.append(value)
        else:
            table[value] = target
    return data, table, unhandled, source["address"], problems


def merge(pairs):
    """輸出兩平台對照表。位址是各自 overlay-02 的 code-local offset。"""
    solved = {}
    for label, path in pairs:
        solved[label] = solve(path)
    labels = [label for label, _ in pairs]

    print("# ECL opcode → handler 對照表（overlay-02 `INTERPET`）")
    print()
    print("由 `scripts/ecl_dispatch_table.py` 產生，不要手改。做法是對 opcode")
    print("`00h..FFh` 逐值符號執行 dispatcher 的比較／分支鏈，走到第一個 `call`；")
    print("分支目標一律取 IDA 的 code ref，不從助憶碼字串解析標籤。")
    print("遇到不在支援集合內的指令會停下並標記，本表的『需人工讀』為 0。")
    print()
    for label in labels:
        data, table, unhandled, address, problems = solved[label]
        print("- **%s**：dispatcher `%04Xh..%04Xh`，opcode 來源 `ds:%sh`，"
              "handler %d 個／涵蓋 opcode %d 個／需人工讀 %d 個。"
              % (label.upper(), data["start"], data["end"], address,
                 len(set(table.values())), len(table), len(problems)))
    print()
    print("| opcode | " + " | ".join(l.upper() + " handler" for l in labels) + " | 兩平台位址 |")
    print("|---|" + "---|" * (len(labels) + 1))
    every = sorted({value for label in labels for value in solved[label][1]})
    same = 0
    for value in every:
        cells = []
        targets = []
        for label in labels:
            target = solved[label][1].get(value)
            targets.append(target)
            cells.append("`%04Xh`" % target if target is not None else "—")
        equal = len(set(targets)) == 1 and targets[0] is not None
        same += 1 if equal else 0
        print("| `%02Xh` | %s | %s |" % (value, " | ".join(cells),
                                          "相同" if equal else "不同"))
    print()
    print("位址相同的 opcode：%d／%d。" % (same, len(every)))
    print()
    unhandled_sets = {label: set(solved[label][2]) for label in labels}
    common = set.intersection(*unhandled_sets.values())
    print("兩平台都沒有 handler（走到 epilogue）的 opcode：`%s`。"
          % "`、`".join("%02Xh" % v for v in sorted(common) if v <= 0x40))
    print("（0x41 以上同樣沒有 handler，本表只列到 ECL 指令集上界 0x40。）")
    return 0


def main():
    if len(sys.argv) < 2:
        print(__doc__)
        return 2
    if sys.argv[1] == "--merge":
        pairs = []
        for argument in sys.argv[2:]:
            label, _, path = argument.partition("=")
            pairs.append((label, path))
        return merge(pairs)
    data, items, order = load(sys.argv[1])
    start = data["start"]

    handlers = {}
    problems = {}
    source = {"address": None}
    unhandled = []
    for value in range(0x100):
        target, error = trace(order, items, start, value, source)
        if error:
            problems[value] = error
        elif target is not None:
            handlers.setdefault(target, []).append(value)
        else:
            unhandled.append(value)

    print("dump=%s 函式=%04X..%04X opcode 來源=ds:%sh"
          % (sys.argv[1], data["start"], data["end"], source["address"]))
    print("解出的 handler 數：%d，覆蓋 opcode %d 個；需人工讀 %d 個值"
          % (len(handlers), sum(len(v) for v in handlers.values()), len(problems)))
    print()
    print("| opcode | handler |")
    print("|---|---|")
    for target in sorted(handlers):
        values = handlers[target]
        if len(values) > 8:
            label = "%02X..%02X（%d 個）" % (min(values), max(values), len(values))
        else:
            label = ", ".join("%02X" % v for v in values)
        print("| %s | `%04Xh` |" % (label, target))
    if unhandled:
        print()
        print("走到 epilogue、本 dispatcher 沒有 handler 的 opcode（%d 個）：%s"
              % (len(unhandled), ", ".join("%02X" % v for v in unhandled)))
    if problems:
        print()
        print("需人工讀：")
        seen = {}
        for value, error in problems.items():
            seen.setdefault(error, []).append(value)
        for error, values in seen.items():
            print("  %s ← opcode %s" % (error, ", ".join("%02X" % v for v in values[:12])))
    return 0


if __name__ == "__main__":
    sys.exit(main())
