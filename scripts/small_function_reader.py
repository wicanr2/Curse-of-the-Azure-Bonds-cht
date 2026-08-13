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
# IDA 對 resident 的具名資料會印成 `mov al, byte_24E6A` 而不是 `ds:XXXX`。
NAMED_LOAD = re.compile(r"^mov (al|ax), ([A-Za-z_][A-Za-z0-9_]*)$")
NAMED_STORE = re.compile(r"^mov ([A-Za-z_][A-Za-z0-9_]*), (al|ax)$")
MOV_ARG = re.compile(r"^mov (al|ax), \[bp\+arg_[0-9A-Fa-f]+\]$")
MOV_VAR_SET = re.compile(r"^mov \[bp\+var_[0-9A-Fa-f]+\], (al|ax)$")
MOV_VAR_GET = re.compile(r"^mov (al|ax), \[bp\+var_[0-9A-Fa-f]+\]$")
CALL_ANY = re.compile(r"^call .+$")
STRING_CONST = re.compile(r"^mov di, offset unk_([0-9A-Fa-f]+)$")
LEA_LOCAL = re.compile(r"^lea di, \[bp\+var_[0-9A-Fa-f]+\]$")
JMP_ANY = re.compile(r"^jmp .+$")
NOISE = {"push cs", "nop", "xor ah, ah", "cbw", "mov sp, bp", "pop bp",
         "push bp", "leave"}


SELF_XOR = re.compile(r"^xor (\w+), (\w+)$")


def is_setup(line):
    """只認資料搬移與自我清零；出現運算、比較、分支就不算參數準備。"""
    mnemonic = line.split()[0]
    if mnemonic in ("mov", "push", "pop", "lea", "les", "lds", "nop"):
        return True
    match = SELF_XOR.match(line)
    return bool(match and match.group(1) == match.group(2))


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

    # 尾呼叫：最後一條是無條件跳躍，控制權交給別的函式後不再返回。
    # ⚠ 仍要求**前面每一條都是參數準備**（資料搬移或自我清零），否則就不是
    # 「整個 body 都被解釋掉」，一律留待解讀。
    if lines and JMP_ANY.match(lines[-1]) and not any(
            EPILOGUE_TAIL.match(line) for line in lines):
        setup = lines[:-1]
        if all(is_setup(line) for line in setup):
            detail = ("；先設定 %s" % "、".join("`%s`" % l for l in setup)) if setup else ""
            return "尾呼叫", "最後一條是 `%s`，控制權轉交後不返回%s" % (lines[-1], detail)
        return None, None

    # ⚠ 沒有 ret 也沒有尾跳躍 ⇒ 這不是完整函式，是 IDA 在 raw overlay 上
    # 建錯的邊界碎片（實測 ON GOTO handler 被切成 3 bytes）。標成獨立狀態，
    # 不混進待解讀，也不冒稱已解讀。
    has_return = any(EPILOGUE_TAIL.match(line) for line in lines)
    if not has_return:
        return "邊界碎片", ("body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `%s`；"
                            "這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀"
                            % (lines[-1] if lines else "(空)"))

    core = [line for line in core_instructions(function["items"])
            if line not in NOISE]

    if not core:
        return "空函式", "prologue／epilogue 之外沒有任何指令，呼叫即返回"

    # 全域搬移：把一個具名／固定位址的值搬到另一個，沒有其他動作。
    named_loads = [NAMED_LOAD.match(l) or MOV_LOAD.match(l) for l in core]
    named_stores = [NAMED_STORE.match(l) or MOV_STORE.match(l) for l in core]
    if (any(named_loads) and any(named_stores)
            and all(named_loads[i] or named_stores[i] or MOV_VAR_SET.match(core[i])
                    or MOV_VAR_GET.match(core[i]) for i in range(len(core)))):
        source = next(m.group(2) for m in named_loads if m)
        destination = next(m.group(1) for m in named_stores if m)
        return "全域搬移", "把 %s 的值搬到 %s（兩者語意另計）" % (source, destination)

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

    # 常數字串函式：把一段 Pascal 字串常數複製進區域緩衝，再呼叫一個 routine。
    # 字串常數就在同一段 overlay 內，可以直接解出來，因此這是完整判讀。
    consts = [STRING_CONST.match(l) for l in core]
    if any(consts):
        allowed = all(
            consts[i] or LEA_LOCAL.match(core[i]) or CALL_ANY.match(core[i])
            or MOV_IMM.match(core[i]) or MOV_LOAD.match(core[i])
            or MOV_ARG.match(core[i]) or MOV_VAR_GET.match(core[i])
            or MOV_VAR_SET.match(core[i]) or core[i].startswith("push ")
            for i in range(len(core)))
        offsets = {m.group(1) for m in consts if m}
        if allowed and len(offsets) == 1:
            return "常數字串", "unk_%s" % offsets.pop()

    # 單一呼叫包裝：整個 body 只有「準備參數」與一個 call／jmp。
    # 允許的準備動作限定為：推參數、推立即數、推全域、載入 far pointer、
    # 暫存搬移與對應的 pop。只要出現算術、比較、分支就不算。
    calls = [l for l in core if CALL_ANY.match(l) or JMP_ANY.match(l)]
    if len(calls) == 1:
        prepare = (MOV_ARG, MOV_VAR_SET, MOV_VAR_GET, MOV_IMM, MOV_LOAD, LEA_LOCAL)
        if all(CALL_ANY.match(l) or JMP_ANY.match(l)
               or any(pattern.match(l) for pattern in prepare)
               or l.startswith(("push ", "pop ", "les ", "lea ", "mov di,", "mov si,"))
               for l in core):
            kind = "單一呼叫包裝" if len(core) > 2 else "轉呼叫"
            # 字串指派：IDA 已在反組譯行尾以 `; "…"` 標出字面值，直接引用它，
            # 不要自己去 overlay 內重算位址（那是另一個位址空間的問題）。
            if "basg" in calls[0]:
                literals = []
                for item in function["items"]:
                    match = re.search(r';\s*"([^"]*)"', item["disasm"])
                    if match:
                        literals.append(match.group(1))
                if literals:
                    return kind, ("字串指派：把字面值「%s」寫入目的字串變數"
                                  % "」「".join(literals))
            return kind, "整個 body 只準備參數並執行 `%s`" % calls[0]

    return None, None


def decode_string(blob, marker, platform):
    """把 `unk_XXXX` 換成實際的 Pascal 字串內容（首位元組是長度）。"""
    offset = int(marker.split("_")[1], 16)
    if blob is None or offset >= len(blob):
        return "字串常數 %s（無法讀取所在 overlay）" % marker
    length = blob[offset]
    raw = blob[offset + 1:offset + 1 + length]
    encoding = "cp932" if platform == "pc98" else "cp437"
    try:
        text = raw.decode(encoding)
    except Exception:
        text = raw.decode("latin-1", "replace")
    return "以固定字串「%s」（%s，長度 %d）呼叫訊息 routine" % (text, marker, length)


def main():
    write = "--write" in sys.argv
    # 已由別的規格判定過的函式不覆蓋：那些條目帶著各自的證據（例如 RTL 的
    # Borland 名稱、跨平台位元組比對），本腳本的形狀判定資訊量較低。
    existing = {}
    if os.path.exists(LEDGER):
        for entry in json.load(open(LEDGER, encoding="utf-8"))["functions"]:
            if entry.get("spec") and entry["spec"] != SPEC:
                existing[(entry["platform"], entry["module"], entry["ea"])] = entry["spec"]
    entries = []
    counts = collections.Counter()
    unmatched = collections.Counter()

    for platform in ("dos", "pc98"):
        for path in sorted(glob.glob(os.path.join(SWEEP, platform, "small", "*.json"))):
            # `*.big.json` 是為跨平台位元組比對另存的全量匯出（上限 4096），
            # 與同名的小函式檔重複，模組名也對不上台帳，必須排除。
            if path.endswith(".error.log") or path.endswith(".big.json"):
                continue
            module = os.path.basename(path)[:-5].replace(".bin", "")
            blob = None
            binary = os.path.join(SWEEP, platform, "overlays", module + ".bin")
            if os.path.exists(binary):
                blob = open(binary, "rb").read()
            for function in json.load(open(path, encoding="utf-8"))["functions"]:
                if (platform, module, function["ea"]) in existing:
                    counts["已由其他規格判定"] += 1
                    continue
                kind, note = classify(function)
                if kind == "常數字串":
                    note = decode_string(blob, note, platform)
                if kind is None:
                    counts["待解讀"] += 1
                    core = core_instructions(function["items"])
                    unmatched[core[0] if core else "(空)"] += 1
                    continue
                counts[kind] += 1
                state = "邊界碎片" if kind == "邊界碎片" else "已解讀"
                entries.append({
                    "platform": platform, "module": module, "ea": function["ea"],
                    "state": state, "level": "" if state != "已解讀" else "exact",
                    "spec": SPEC,
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
