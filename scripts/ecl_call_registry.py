"""列出 ECL `CALL`（opcode `2Dh`）handler 認得的**全部** selector。

CALL handler 先取得 operand，再 `sub ax, 7FFFh` 得到 selector，然後是一串
`cmp ax, <selector>` 分支。因此原始 ECL operand 與 selector 的關係是

    operand = (selector + 7FFFh) mod 10000h

（驗證：selector `AE11h` → operand `2E10h`，與 spec 519 一致。）

做法與 `ecl_dispatch_table.py` 相同：對 selector 逐值符號執行比較鏈，走到
第一個不屬於比較鏈的指令就停，該位址即這個 selector 的分支主體。分支目標
一律取 IDA 的 code ref。任何不在支援集合內、可能改變結果的指令都會停下並
標記，不猜。

用法：
    python3 scripts/ecl_call_registry.py <call-handler-dump.json>
    python3 scripts/ecl_call_registry.py --merge dos=<dump> pc98=<dump>
"""

import json
import re
import sys

CMP_RE = re.compile(r"^cmp\s+(al|ax),\s*([0-9A-Fa-f]+h|[0-9]+)$")
JCC_RE = re.compile(r"^(jz|jnz|je|jne)\s+(?:short\s+)?\S+$")
JMP_RE = re.compile(r"^jmp\s+(?:short\s+)?\S+$")
BIAS_RE = re.compile(r"^sub\s+ax,\s*([0-9A-Fa-f]+)h$")
RELOAD_RE = re.compile(r"^mov\s+ax,\s*\[bp\+var_[0-9A-Fa-f]+\]$")
IGNORED_PREFIX = ("nop", "push ax", "mov [bp+var_")
EPILOGUE = ("mov sp, bp", "pop bp", "retn", "leave", "ret")


def normalise(item):
    text = re.sub(r"\s*;.*$", "", item["disasm"].strip())
    return re.sub(r"\s+", " ", text)


def load(path):
    with open(path, encoding="utf-8") as handle:
        data = json.load(handle)
    items = {item["ea"]: item for item in data["items"]}
    order = [item["ea"] for item in data["items"]]
    return data, items, order


def find_chain_start(items, order):
    """回傳 (比較鏈起點, bias)。bias 是 `sub ax, NNNNh` 的立即數。"""
    for index, ea in enumerate(order):
        match = BIAS_RE.match(normalise(items[ea]))
        if match:
            bias = int(match.group(1), 16)
            # bias 之後可能還有暫存來回搬運，跳過到第一個 cmp 之前的那個 reload。
            for follow in order[index + 1:]:
                text = normalise(items[follow])
                if CMP_RE.match(text):
                    return follow, bias
                if RELOAD_RE.match(text) or text.startswith("mov [bp+var_"):
                    continue
                return follow, bias
    return None, None


def parse_immediate(text):
    """帶 `h` 後綴才是十六進位。後綴必須留給本函式判斷，不能在正規表示式吃掉，
    否則 `cmp ax, 3201h` 會被當成十進位 3201（＝0C81h）。"""
    if text.endswith("h"):
        return int(text[:-1], 16)
    return int(text, 10)


def next_ea(order, ea):
    index = order.index(ea)
    return order[index + 1] if index + 1 < len(order) else None


def target_of(item):
    return item["code_refs"][0] if item["code_refs"] else None


def trace(order, items, start, value):
    ea = start
    steps = 0
    while ea is not None and steps < 4096:
        steps += 1
        item = items.get(ea)
        if item is None:
            return None, "位址不在 dump 範圍 %04X" % ea
        text = normalise(item)

        match = CMP_RE.match(text)
        if match:
            equal = (value == parse_immediate(match.group(2)))
            ea = next_ea(order, ea)
            item = items.get(ea)
            if item is None:
                return None, "cmp 之後沒有指令"
            text = normalise(item)
            match = JCC_RE.match(text)
            if not match:
                # cmp 後直接進主體（少見但合法）
                return ea, None
            taken = equal if match.group(1) in ("jz", "je") else not equal
            ea = target_of(item) if taken else next_ea(order, ea)
            continue

        if JMP_RE.match(text):
            destination = target_of(item)
            if destination is None:
                return None, "jmp 沒有 code ref"
            ea = destination
            continue

        if any(text.startswith(prefix) for prefix in EPILOGUE):
            return None, None

        if RELOAD_RE.match(text) or any(text.startswith(p) for p in IGNORED_PREFIX):
            ea = next_ea(order, ea)
            continue

        # 第一個不屬於比較鏈的指令＝這個 selector 的分支主體
        return ea, None
    return None, "步數超過上限"


def solve(path):
    data, items, order = load(path)
    start, bias = find_chain_start(items, order)
    if start is None:
        raise SystemExit("%s：找不到 `sub ax, NNNNh` 偏移" % path)
    table, problems = {}, {}
    for value in range(0x10000):
        body, error = trace(order, items, start, value)
        if error:
            problems[value] = error
        elif body is not None:
            table[value] = body
    return data, bias, table, problems


def report(label, data, bias, table, problems):
    """未被任何 `cmp` 匹配的 selector 會走到 epilogue，trace 直接回 None，
    因此不會進 table。凡是進了 table 的就是具名 selector。

    ⚠ 這裡曾經用「selector 最多的分支＝default」的啟發式，但每個分支各只有
    一個 selector 時，它會把第一筆真的 selector 當成 default 刪掉（實測少報了
    `0001h`）。不要再引入這種以量取勝的猜測。
    """
    return {
        "label": label, "start": data["start"], "end": data["end"], "bias": bias,
        "named": dict(table), "problems": len(problems),
    }


def main():
    if len(sys.argv) < 2:
        print(__doc__)
        return 2
    pairs = []
    if sys.argv[1] == "--merge":
        for argument in sys.argv[2:]:
            label, _, path = argument.partition("=")
            pairs.append((label, path))
    else:
        pairs.append(("", sys.argv[1]))

    results = []
    for label, path in pairs:
        data, bias, table, problems = solve(path)
        results.append(report(label or path, data, bias, table, problems))

    print("# ECL `CALL`（opcode `2Dh`）external routine registry")
    print()
    print("由 `scripts/ecl_call_registry.py` 產生，不要手改。")
    print("selector 與原始 ECL operand 的關係是 `operand = (selector + bias) mod 10000h`；")
    print("bias 由 handler 內的 `sub ax, NNNNh` 直接讀出，不是假設值。")
    print()
    for result in results:
        print("- **%s**：handler `%04Xh..%04Xh`，bias `%04Xh`，認得的 selector %d 個，"
              "需人工讀 %d 個。其餘 selector 走到 epilogue（不做事就返回）。"
              % (result["label"].upper(), result["start"], result["end"],
                 result["bias"], len(result["named"]), result["problems"]))
    print()

    labels = [result["label"] for result in results]
    selectors = sorted({s for result in results for s in result["named"]})
    print("| ECL operand | selector | " + " | ".join(l.upper() + " 分支" for l in labels) + " |")
    print("|---|---|" + "---|" * len(labels))
    for selector in selectors:
        operand = (selector + results[0]["bias"]) & 0xFFFF
        cells = []
        for result in results:
            body = result["named"].get(selector)
            cells.append("`%04Xh`" % body if body is not None else "—")
        print("| `%04Xh` | `%04Xh` | %s |" % (operand, selector, " | ".join(cells)))
    print()
    print("`分支` 是該 selector 在各自 overlay-02 內的 code-local 位址；兩平台位址")
    print("不同是正常的（見 spec 560），對應依據是 selector 值本身。")
    return 0


if __name__ == "__main__":
    sys.exit(main())
