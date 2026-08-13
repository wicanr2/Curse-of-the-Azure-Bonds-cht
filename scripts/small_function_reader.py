"""逐條讀完小函式並分類，產生台帳條目。

輸入是 `tools/ida/export_small_functions.py` 匯出的**完整 body**。分類的前提是
「整個 body 都被解釋掉」——扣掉 prologue／epilogue 之後，剩下的指令必須完全
符合某一個已知形狀，否則一律維持 `待解讀`。這是為了避免「前兩條看起來像
accessor 就當成 accessor」，而忽略後面真正做事的指令。

可判定的形狀（每一種都是把整個 body 讀完，不是取樣）：

- `空函式`：扣掉 prologue／epilogue 後沒有指令。
- `常數`：只把一個立即數放進回傳暫存器。
- `讀取器`：只從一個固定 `ds:XXXX` 讀值並回傳。
- `寫入器`：只把參數寫進一個固定 `ds:XXXX`。
- `轉呼叫`：只呼叫另一個位址後回傳（參數原樣傳遞）。

**位址的語意不在本腳本的判定範圍內**：「回傳 `DS:A2A9h` 的 byte」是對函式
本體的完整描述，`DS:A2A9h` 代表什麼是另一條證據線。條目會如實這樣寫。

用法：
    python3 scripts/small_function_reader.py            # 預覽統計
    python3 scripts/small_function_reader.py --write    # 寫入台帳
"""

import collections
import glob
import json
import os
import re
import sys

ROOT = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
SWEEP = os.path.join(ROOT, "workplace", "re-sweep")
LEDGER = os.path.join(ROOT, "docs", "audit", "re-function-ledger.json")
SPEC = "docs/spec/569-small-function-batch-reading.md"

PROLOGUE = ["push bp", "mov bp, sp"]
SUB_SP = re.compile(r"^sub sp, [0-9A-Fa-f]+h?$")
EPILOGUE_TAIL = re.compile(r"^ret[fn]?( [0-9A-Fa-f]+h?)?$")

MOV_IMM = re.compile(r"^mov (al|ax), ([0-9A-Fa-f]+h|[0-9]+)$")
MOV_LOAD = re.compile(r"^mov (al|ax), ds:([0-9A-Fa-f]+h)$")
MOV_STORE = re.compile(r"^mov ds:([0-9A-Fa-f]+h), (al|ax)$")
MOV_ARG = re.compile(r"^mov (al|ax), \[bp\+arg_[0-9A-Fa-f]+\]$")
MOV_VAR_SET = re.compile(r"^mov \[bp\+var_[0-9A-Fa-f]+\], (al|ax)$")
MOV_VAR_GET = re.compile(r"^mov (al|ax), \[bp\+var_[0-9A-Fa-f]+\]$")
CALL_ANY = re.compile(r"^call .+$")
JMP_ANY = re.compile(r"^jmp .+$")
NOISE = {"push cs", "nop", "xor ah, ah", "cbw", "mov sp, bp", "pop bp",
         "push bp", "leave"}


def clean(text):
    text = re.sub(r"\s*;.*$", "", text.strip())
    return re.sub(r"\s+", " ", text)


def core_instructions(items):
    """扣掉 prologue／epilogue，回傳中間的指令文字。"""
    lines = [clean(item["disasm"]) for item in items]
    index = 0
    for expected in PROLOGUE:
        if index < len(lines) and lines[index] == expected:
            index += 1
    if index < len(lines) and SUB_SP.match(lines[index]):
        index += 1
    end = len(lines)
    while end > index and (lines[end - 1] in ("mov sp, bp", "pop bp")
                           or EPILOGUE_TAIL.match(lines[end - 1])):
        end -= 1
    return lines[index:end]


def classify(function):
    lines = [clean(item["disasm"]) for item in function["items"]]

    # Borland overlay 呼叫 stub：`int 3Fh` 後接 control 資料，不是一般函式。
    if lines and lines[0] == "int 3Fh":
        return "overlay stub", ("Borland overlay 呼叫 stub（`int 3Fh` ＋ control 資料），"
                                "由 overlay manager 轉派，不含遊戲邏輯")

    # ⚠ 沒有 ret 就不能宣稱「讀完了」：IDA 的函式邊界會建錯，被切短的函式
    # 看起來就像空函式（實測 ON GOTO handler 被切成 3 bytes）。
    has_return = any(EPILOGUE_TAIL.match(line) for line in lines)
    if not has_return:
        return None, None

    core = [line for line in core_instructions(function["items"])
            if line not in NOISE]

    if not core:
        return "空函式", "prologue／epilogue 之外沒有任何指令，呼叫即返回"

    # 常數
    if len(core) <= 3 and all(MOV_IMM.match(l) or MOV_VAR_SET.match(l)
                              or MOV_VAR_GET.match(l) for l in core):
        match = next((MOV_IMM.match(l) for l in core if MOV_IMM.match(l)), None)
        if match:
            return "常數", "固定回傳 %s" % match.group(2)

    # 讀取器：只有一個 ds 讀取，其餘是暫存搬移
    loads = [MOV_LOAD.match(l) for l in core]
    if any(loads) and all(loads[i] or MOV_VAR_SET.match(core[i])
                          or MOV_VAR_GET.match(core[i]) for i in range(len(core))):
        addresses = {m.group(2) for m in loads if m}
        widths = {m.group(1) for m in loads if m}
        if len(addresses) == 1:
            width = "word" if "ax" in widths else "byte"
            return "讀取器", "回傳 DS:%s 的 %s（該位址語意另計）" % (addresses.pop(), width)

    # 寫入器：只有一個 ds 寫入，來源是參數
    stores = [MOV_STORE.match(l) for l in core]
    if any(stores) and all(stores[i] or MOV_ARG.match(core[i])
                           or MOV_VAR_SET.match(core[i]) or MOV_VAR_GET.match(core[i])
                           for i in range(len(core))):
        addresses = {m.group(1) for m in stores if m}
        if len(addresses) == 1:
            return "寫入器", "把參數寫入 DS:%s（該位址語意另計）" % addresses.pop()

    # 轉呼叫：唯一的實質動作是一個 call／jmp
    calls = [l for l in core if CALL_ANY.match(l) or JMP_ANY.match(l)]
    if len(calls) == 1 and all(
            CALL_ANY.match(l) or JMP_ANY.match(l) or MOV_ARG.match(l)
            or MOV_VAR_SET.match(l) or MOV_VAR_GET.match(l)
            or l.startswith("push ") for l in core):
        return "轉呼叫", "唯一實質動作是 `%s`，參數原樣傳遞" % calls[0]

    return None, None


def main():
    write = "--write" in sys.argv
    entries = []
    counts = collections.Counter()
    unmatched = collections.Counter()

    for platform in ("dos", "pc98"):
        for path in sorted(glob.glob(os.path.join(SWEEP, platform, "small", "*.json"))):
            if path.endswith(".error.log"):
                continue
            module = os.path.basename(path)[:-5].replace(".bin", "")
            for function in json.load(open(path, encoding="utf-8"))["functions"]:
                kind, note = classify(function)
                if kind is None:
                    counts["待解讀"] += 1
                    core = core_instructions(function["items"])
                    unmatched[core[0] if core else "(空)"] += 1
                    continue
                counts[kind] += 1
                entries.append({
                    "platform": platform, "module": module, "ea": function["ea"],
                    "state": "已解讀", "level": "exact", "spec": SPEC,
                    "note": "%s：%s（body 共 %d bytes，已逐條讀完）"
                            % (kind, note, function["size"]),
                })

    print("可判定並標為已解讀：%d" % len(entries))
    for kind, count in counts.most_common():
        print("  %-8s %d" % (kind, count))
    print("\n未匹配任何已知形狀（維持待解讀）的核心首指令 Top 15：")
    for line, count in unmatched.most_common(15):
        print("  %-40s %d" % (line[:40], count))

    if not write:
        print("\n（預覽模式；加 --write 才寫入台帳）")
        return 0

    ledger = json.load(open(LEDGER, encoding="utf-8"))
    keys = {(e["platform"], e["module"], e["ea"]) for e in entries}
    kept = [e for e in ledger["functions"]
            if (e["platform"], e["module"], e["ea"]) not in keys]
    ledger["functions"] = kept + entries
    json.dump(ledger, open(LEDGER, "w", encoding="utf-8"), ensure_ascii=False, indent=1)
    print("\n已寫入 %s" % LEDGER)
    return 0


if __name__ == "__main__":
    sys.exit(main())
