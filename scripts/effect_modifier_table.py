"""抽出每個效果 handler 對「規則暫存全域」做的加減。

`CHECKFX(timing, subject)` 只回答「這個時機有哪些效果要介入」
（`docs/audit/checkfx-timing-table.md`）。真正的數字在 handler 本體裡：
它們是一批極短的函式，把修正**寫進一組固定的全域**，呼叫端再讀自己要的那一個。

    效果 01h（祝福）：add ds:6F9Fh, 1   add ds:6FA2h, 5
    效果 02h（詛咒）：dec ds:6F9Fh      sub ds:6FA2h, 5（夾在 0）
    效果 21h（致盲）：sub ds:6F9Fh, 4   sub ds:6F92h, 4   角色^[19Ah]／[19Bh] −4
    效果 2Ah（緩速）：ds:6F96h := ds:6F96h div 2

所以「效果碼 → 修正」是可以機械抽出來的，不必一支一支讀。抽不出來的（有條件
分支、呼叫別的函式、動到角色記錄）一律標 `待解讀` 並附指令數，**不猜**。

handler 的位址取自 `docs/audit/effect-dispatch-table.md`（spec 1005），
反組譯用 objdump 直接讀 overlay 的 `.bin`——IDA 對這一批小函式的邊界常常只認出
前幾條，線性反組譯才看得到完整本體。

用法：python3 scripts/effect_modifier_table.py [--write]
"""

import collections
import json
import os
import re
import subprocess
import sys

ROOT = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
DISPATCH_MD = os.path.join(ROOT, "docs/audit/effect-dispatch-table.md")
TIMING_MD = os.path.join(ROOT, "docs/audit/checkfx-timing-table.md")
OVERLAY_DIR = os.path.join(ROOT, "workplace/re-sweep/dos/overlays")
OUT_MD = os.path.join(ROOT, "docs/audit/effect-modifier-table.md")
OUT_JSON = os.path.join(ROOT, "gamepack/rules/effect-modifiers.json")

# 規則暫存全域。名字只給有獨立佐證的那幾個，其餘保留位址。
SCRATCH = {
    # ⚠ `6F9Fh` 是**共用的修正暫存**：同一支 handler 無條件寫它，而它的意思由
    # 「哪個 timing 讀它」決定（命中、護甲、豁免…）。所以名字保持中性，
    # 不寫成「命中修正」——那會在護甲那條路上變成錯的斷言。
    "6f9f": ("modifier", "共用修正暫存；語意由讀它的 timing 決定"),
    "6f92": ("saving_throw", "豁免總和（`overlay-23 entry#8` 寫它，spec 1119）"),
    "6fa2": ("morale", "士氣（`overlay-09:01388h` 寫它，spec 1122）"),
    "6f96": ("movement", "移動率（`overlay-13 entry#2` 寫它，spec 1122）"),
    "6f9b": ("attack_forced_miss", "攻擊必失手旗標"),
    # 抗寒（`0Ah`）與抗火（`14h`）兩支都把 `6F94h` 折半，而且各自先檢查
    # `6F95h` 的不同位元——兩個獨立的證人指向同一組語意。
    "6f94": ("damage", "傷害值（抗寒／抗火在傷害時機把它折半）"),
    "6f95": ("damage_element", "傷害屬性旗標：bit 0 火、bit 1 冷"),
    "6f9c": ("scratch_6f9c", "（未定）"),
}

PROLOGUE = ("push   bp", "mov    bp,sp", "mov    sp,bp", "pop    bp")

ADD = re.compile(r"^(add|sub)\s+BYTE PTR ds:0x([0-9a-f]{4}),0x([0-9a-f]+)$")
STEP = re.compile(r"^(inc|dec)\s+BYTE PTR ds:0x([0-9a-f]{4})$")
SET = re.compile(r"^mov\s+BYTE PTR ds:0x([0-9a-f]{4}),0x([0-9a-f]+)$")


def read_dispatch():
    """效果碼 → (overlay 編號, 位移)。"""
    pattern = re.compile(
        r"\|\s*(\d+)\s*\|\s*`overlay-(\d+):0*([0-9A-Fa-f]+)h`（entry#\d+）\s*\|")
    table = {}
    for line in open(DISPATCH_MD, encoding="utf-8"):
        found = pattern.match(line)
        if found:
            table[int(found.group(1))] = (int(found.group(2)),
                                          int(found.group(3), 16))
    if not table:
        raise SystemExit("效果分派表解析不出任何一列——格式變了")
    return table


def read_timings():
    """timing → 效果碼清單。"""
    row = re.compile(r"^\|\s*`([0-9A-Fa-f]{2})h`\s*\|\s*(.*?)\s*\|$")
    timings = {}
    for line in open(TIMING_MD, encoding="utf-8"):
        found = row.match(line.rstrip())
        if not found:
            continue
        codes = [int(value, 16)
                 for value in re.findall(r"`([0-9A-Fa-f]{2})h`", found.group(2))]
        timings[int(found.group(1), 16)] = codes
    if not timings:
        raise SystemExit("timing 表解析不出任何一列——格式變了")
    return timings


def disassemble(module, offset, limit=0x140):
    path = os.path.join(OVERLAY_DIR, "overlay-%02d.bin" % module)
    output = subprocess.run(
        ["objdump", "-D", "-b", "binary", "-m", "i8086", "-M", "intel",
         "--start-address=%d" % offset, "--stop-address=%d" % (offset + limit),
         path], capture_output=True, text=True).stdout
    body = []
    for line in output.splitlines():
        found = re.match(r"\s+([0-9a-f]+):\s+([0-9a-f ]+)\t(.*)", line)
        if not found:
            continue
        text = found.group(3).strip()
        body.append((int(found.group(1), 16), text))
        if text.startswith("retf") or text == "ret":
            break
    return body


BRANCH = re.compile(r"^j(?!mp)\w+\s+0x([0-9a-f]+)$")
JUMP = re.compile(r"^jmp\s+0x([0-9a-f]+)$")
# 夾底的減法：`cmp ds:G,K / jae L1 / mov ds:G,0 / jmp L2 / L1: sub ds:G,K`。
# 這個慣用法不能拆成兩個獨立的動作——把 `mov ds:G,0` 當成無條件的設定，
# 詛咒就會把士氣直接歸零而不是減 5。
CLAMP_CMP = re.compile(r"^cmp\s+BYTE PTR ds:0x([0-9a-f]{4}),0x([0-9a-f]+)$")
# 折半：`mov al,ds:G / … / mov cx,2 / idiv cx / mov ds:G,al`。
HALF_LOAD = re.compile(r"^mov\s+al,ds:0x([0-9a-f]{4})$")
HALF_STORE = re.compile(r"^mov\s+ds:0x([0-9a-f]{4}),al$")


def conditional_addresses(body):
    """不是無條件執行的指令位址。

    兩種來源都要看：**條件跳躍**之後到目標之前（那一段只在條件成立時跑），
    以及**無條件跳躍**之後到目標之前（那一段是被跳過的另一條分支）。
    ⚠ 少了 `jmp` 那一半，`if 未達上限 then +2 else := 上限` 的 else 分支會被當成
    無條件執行——妖火就會變成「把護甲直接設成上限」而不是「加 2」。
    """
    inside = set()
    for index, (address, text) in enumerate(body):
        found = BRANCH.match(text) or JUMP.match(text)
        if not found:
            continue
        target = int(found.group(1), 16)
        for other, _ in body[index + 1:]:
            if address < other < target:
                inside.add(other)
    return inside


# 封頂的加法：`cmp F,C / jae L1 / add F,K / jmp L2 / L1: mov F,CAP`。
CAP_CMP_RECORD = re.compile(r"^cmp\s+BYTE PTR es:\[di\+0x([0-9a-f]+)\],0x([0-9a-f]+)$")


def match_capped_add(body, index):
    """認出「未達上限就加 K，否則設成上限」。回傳 (運算, 用掉幾條指令) 或 None。"""
    load = RECORD_LOAD.match(body[index][1])
    if not load or index + 6 >= len(body):
        return None
    compare = CAP_CMP_RECORD.match(body[index + 1][1])
    if not compare or not body[index + 2][1].startswith("jae"):
        return None
    field, threshold = int(compare.group(1), 16), int(compare.group(2), 16)
    if not RECORD_LOAD.match(body[index + 3][1]):
        return None
    added = RECORD_ADD.match(body[index + 4][1])
    if not added or added.group(1) != "add" or int(added.group(2), 16) != field:
        return None
    if not body[index + 5][1].startswith("jmp"):
        return None
    if not RECORD_LOAD.match(body[index + 6][1]):
        return None
    capped = RECORD_SET.match(body[index + 7][1]) if index + 7 < len(body) else None
    if not capped or int(capped.group(1), 16) != field:
        return None
    return {"record": "player", "field": field, "op": "add_capped",
            "value": int(added.group(3), 16),
            "cap": int(capped.group(2), 16), "cap_threshold": threshold}, 8


# 旗標守衛：`mov al, ds:G / and al, K / or al,al / je L`——L 之前的指令只有在
# `G and K <> 0` 時才會跑。抗寒與抗火就是這個形狀（各自看傷害屬性旗標的一個位元）。
GUARD_LOAD = re.compile(r"^mov\s+al,ds:0x([0-9a-f]{4})$")
GUARD_MASK = re.compile(r"^and\s+al,0x([0-9a-f]+)$")


def flag_guards(body):
    """回傳 位址 → (守衛的全域, 位元遮罩)。只認得出一層守衛。"""
    guards = {}
    for index in range(len(body) - 3):
        load = GUARD_LOAD.match(body[index][1])
        mask = GUARD_MASK.match(body[index + 1][1])
        if not load or not mask or body[index + 2][1] != "or     al,al":
            continue
        branch = BRANCH.match(body[index + 3][1])
        if not branch or not body[index + 3][1].startswith("je"):
            continue
        target = int(branch.group(1), 16)
        for address, _ in body[index + 4:]:
            if address >= target:
                break
            guards[address] = (load.group(1), int(mask.group(1), 16))
    return guards


# 寫進**記錄**而不是全域的那一類：
#
#     les di, [bp+0Ch]                 { 對象 }
#     les di, es:[di+18Dh]             { → 戰鬥狀態記錄，可省略 }
#     mov byte ptr es:[di+K], V        { 或 add／sub }
#
# 纏繞術（`88h`）把戰鬥狀態的 `+06h`（移動率）設成 0、妖火（`07h`）把角色的
# `+19Ah`／`+19Bh`（護甲兩格）各加 2。這些不是全域修正，但同樣是規則。
RECORD_LOAD = re.compile(r"^les\s+di,DWORD PTR \[bp\+0x([0-9a-f]+)\]$")
RECORD_DEREF = re.compile(r"^les\s+di,DWORD PTR es:\[di\+0x([0-9a-f]+)\]$")
RECORD_SET = re.compile(r"^mov\s+BYTE PTR es:\[di\+0x([0-9a-f]+)\],0x([0-9a-f]+)$")
RECORD_ADD = re.compile(r"^(add|sub)\s+BYTE PTR es:\[di\+0x([0-9a-f]+)\],0x([0-9a-f]+)$")


def match_record_write(body, index):
    """認出「載入對象記錄再寫一格」。回傳 (運算, 用掉幾條指令) 或 None。"""
    found = RECORD_LOAD.match(body[index][1])
    if not found or index + 1 >= len(body):
        return None
    used, record = 1, "player"
    deref = RECORD_DEREF.match(body[index + 1][1])
    if deref:
        if int(deref.group(1), 16) != 0x18D:
            return None
        record, used = "combat_state", 2
    if index + used >= len(body):
        return None
    text = body[index + used][1]
    written = RECORD_SET.match(text)
    if written:
        return {"record": record, "field": int(written.group(1), 16), "op": "set",
                "value": int(written.group(2), 16)}, used + 1
    written = RECORD_ADD.match(text)
    if written:
        return {"record": record, "field": int(written.group(2), 16),
                "op": written.group(1), "value": int(written.group(3), 16)}, used + 1
    return None


def match_clamped_sub(body, index):
    """認出夾底減法，回傳 (運算, 用掉幾條指令) 或 None。"""
    found = CLAMP_CMP.match(body[index][1])
    if not found or index + 4 >= len(body):
        return None
    glob, amount = found.group(1), int(found.group(2), 16)
    window = [text for _, text in body[index + 1:index + 5]]
    if not window[0].startswith("jae"):
        return None
    zeroed = SET.match(window[1])
    if not zeroed or zeroed.group(1) != glob or int(zeroed.group(2), 16) != 0:
        return None
    if not window[2].startswith("jmp"):
        return None
    reduced = ADD.match(window[3])
    if not reduced or reduced.group(1) != "sub" or reduced.group(2) != glob:
        return None
    if int(reduced.group(3), 16) != amount:
        return None
    return {"global": glob, "op": "sub_clamped", "value": amount}, 5


def match_halve(body, index):
    """認出「除以常數再寫回同一個全域」。"""
    found = HALF_LOAD.match(body[index][1])
    if not found:
        return None
    glob = found.group(1)
    divisor = None
    for offset in range(index + 1, min(index + 8, len(body))):
        text = body[offset][1]
        divisor_found = re.match(r"^mov\s+cx,0x([0-9a-f]+)$", text)
        if divisor_found:
            divisor = int(divisor_found.group(1), 16)
        if text.startswith("idiv") and divisor:
            store = HALF_STORE.match(body[offset + 1][1]) if offset + 1 < len(body) else None
            if store and store.group(1) == glob:
                return {"global": glob, "op": "div", "value": divisor}, offset + 2 - index
            return None
    return None


def classify(body):
    """回傳 (修正清單, 無法辨識的指令數)。

    只收**無條件執行**的修正。條件分支裡的單一指令一律算成沒解析——
    照字面收下來會產生「看起來完整但語意錯」的表，比空表更糟。
    """
    conditional = conditional_addresses(body)
    guards = flag_guards(body)
    operations, unparsed = [], 0
    index = 0
    while index < len(body):
        address, text = body[index]
        if text in PROLOGUE or text.startswith("retf") or text.startswith("sub    sp,"):
            index += 1
            continue
        # ⚠ 慣用法的比對也要吃條件判斷。少了這一行，抗寒／抗火的「傷害折半」
        # 會被當成無條件——那兩支其實先檢查傷害屬性旗標（`ds:6F95h`），
        # 只對冷／火生效。無條件套用等於讓它們減半**所有**傷害。
        guard = guards.get(address)
        # 被旗標守衛包住的指令仍然收，但把守衛記在運算上——套用端要先確認
        # 那個旗標。少了 `guard`，抗寒會減半**所有**傷害而不只是冷傷害。
        matched = None
        if address not in conditional or guard:
            for matcher in (match_capped_add, match_clamped_sub, match_halve,
                            match_record_write):
                matched = matcher(body, index)
                if matched:
                    break
        if matched:
            operation = matched[0]
            if guard:
                operation["guard_global"], operation["guard_mask"] = guard
            operations.append(operation)
            index += matched[1]
        else:
            if address in conditional and not guard:
                unparsed += 1
                index += 1
                continue
            found = ADD.match(text)
            if found:
                operation = {"global": found.group(2), "op": found.group(1),
                             "value": int(found.group(3), 16)}
                if guard:
                    operation["guard_global"], operation["guard_mask"] = guard
                operations.append(operation)
                index += 1
                continue
            found = STEP.match(text)
            if found:
                operation = {"global": found.group(2),
                             "op": "add" if found.group(1) == "inc" else "sub",
                             "value": 1}
                if guard:
                    operation["guard_global"], operation["guard_mask"] = guard
                operations.append(operation)
                index += 1
                continue
            found = SET.match(text)
            if found:
                operation = {"global": found.group(1), "op": "set",
                             "value": int(found.group(2), 16)}
                if guard:
                    operation["guard_global"], operation["guard_mask"] = guard
                operations.append(operation)
                index += 1
                continue
            unparsed += 1
            index += 1
    # ⚠ 同一個全域同時被「設定」與別的運算動到，代表那是兩條分支而不是兩個
    # 依序執行的動作（`if 不足 then 0 else 折半`）。照順序套會先歸零再折半，
    # 結果永遠是 0。認不出那個形狀就整組退回沒解析——**不猜**。
    touched = collections.Counter(item.get("global") for item in operations)
    conflicted = {item.get("global") for item in operations
                  if item["op"] == "set" and item.get("global")
                  and touched[item.get("global")] > 1}
    if conflicted:
        unparsed += sum(1 for item in operations if item.get("global") in conflicted)
        operations = [item for item in operations
                      if item.get("global") not in conflicted]
    return operations, unparsed


def main():
    dispatch, timings = read_dispatch(), read_timings()
    records = {}
    for code in sorted(dispatch):
        module, offset = dispatch[code]
        body = disassemble(module, offset)
        operations, unparsed = classify(body)
        records[code] = {"overlay": module, "offset": offset,
                         "instructions": len(body), "modifiers": operations,
                         "unparsed": unparsed,
                         "status": "decoded" if operations and unparsed == 0
                         else ("partial" if operations else
                               ("inert" if len(body) <= 5 else "unread"))}

    counts = collections.Counter(item["status"] for item in records.values())
    print("效果碼 %d 個：%s" % (len(records), dict(counts)))
    if "--write" not in sys.argv:
        print("（預覽模式；加 --write 才寫報表）")
        return

    payload = {
        "schema_version": 1,
        "source": "DOS overlay handler 本體（位址取自 docs/audit/effect-dispatch-table.md）",
        "spec": "docs/spec/1123-effect-modifier-table.md",
        "scratch_globals": {key: {"name": name, "note": note}
                            for key, (name, note) in sorted(SCRATCH.items())},
        "timings": {"%02X" % timing: codes for timing, codes in sorted(timings.items())},
        "effects": {"%02X" % code: item for code, item in sorted(records.items())},
    }
    with open(OUT_JSON, "w", encoding="utf-8") as handle:
        json.dump(payload, handle, ensure_ascii=False, indent=1)
        handle.write("\n")

    lines = ["# 效果 handler 的修正表", "",
             "由 `scripts/effect_modifier_table.py` 產生，判讀見 "
             "[spec 1123](../spec/1123-effect-modifier-table.md)。不要手改。", "",
             "`CHECKFX(timing)` 只說「這個時機有哪些效果要介入」"
             "（[timing 表](checkfx-timing-table.md)）；本表是**數字**："
             "每個 handler 把修正寫進哪個全域、加多少。", "",
             "| 狀態 | 意思 |", "|---|---|",
             "| `decoded` | 整支都是對全域的加減／設定，沒有其他指令 |",
             "| `partial` | 有加減，但還有沒解析的指令（條件、呼叫、動角色記錄）|",
             "| `inert` | 只有序幕與 `retf`，什麼都不做 |",
             "| `unread` | 有內容但沒有可辨識的加減 |", "",
             "統計：" + "、".join("%s %d" % (key, value)
                                  for key, value in sorted(counts.items())), "",
             "## 暫存全域", "", "| 位址 | 名稱 | 佐證 |", "|---|---|---|"]
    for key, (name, note) in sorted(SCRATCH.items()):
        lines.append("| `DS:%sh` | `%s` | %s |" % (key.upper(), name, note))
    lines += ["", "## 逐效果碼", "",
              "| 碼 | overlay:位移 | 狀態 | 修正 | 出現在哪些 timing |",
              "|---:|---|---|---|---|"]
    for code in sorted(records):
        item = records[code]
        modifiers = "、".join(
            ("`%s+%02Xh` %s %d" % (entry["record"], entry["field"],
                                   {"add": "＋", "sub": "−", "set": "＝"}[entry["op"]],
                                   entry["value"])
             if "record" in entry else
             "`%s` %s %d" % (SCRATCH.get(entry["global"], (entry["global"],))[0],
                            {"add": "＋", "sub": "−", "set": "＝", "sub_clamped": "−（夾底）", "div": "÷"}[entry["op"]],
                            entry["value"]))
            for entry in item["modifiers"]) or "—"
        appears = "、".join("`%02Xh`" % timing for timing in sorted(timings)
                            if code in timings[timing]) or "—"
        lines.append("| `%02Xh` | `overlay-%d:%04Xh` | %s | %s | %s |" % (
            code, item["overlay"], item["offset"], item["status"],
            modifiers, appears))
    open(OUT_MD, "w", encoding="utf-8").write("\n".join(lines) + "\n")
    print("寫出 %s 與 %s" % (OUT_MD, OUT_JSON))


if __name__ == "__main__":
    main()
