"""抽出「傷害／治療類」法術的骰子。

法術屬性表的 `+0Ah`（效果碼）為 0 的那一批，傷害不在表裡——它在**各自的
handler** 裡。handler 的位址來自法術分派表：`DS:72A0h` 起 101 筆遠指標，
由 `overlay-22` 的一支初始化函式填進去（PC-98 對側見 spec 1071）。
段固定是一個 overlay 的 entry stub 段，所以 `位移 → (位移 − 20h) ÷ 5` 就是
entry 編號，再查 `ovr-manifest.json` 得到 handler 的 code offset。

handler 裡擲骰有兩種入口，兩支都是 `(骰數, 面數)` 兩個參數（各自 `retf 4`）：

    <overlay-23 entry#9> (骰數, 面數)          { 擲 骰數 次 1..面數 並加總 }
    <overlay-23 entry#10>(骰數, 面數)          { 順手把骰數記進 DS:6F98h，再轉呼 entry#9 }

兩支的用途**不是**「治療 vs 傷害」：`entry#9` 既被治療輕傷用，也被解除魔法拿去
擲 `1d100`。所以形狀欄位照 entry 編號記，用途要看該支法術自己。

收尾一律是 `sub_F06(法術編號, 等級覆寫, ?, 傷害, 傷害屬性旗標, 訊息)`，所以
**傷害與屬性旗標從 `sub_F06` 的呼叫點取**，不是從擲骰取——燃燒之手根本沒擲骰
（傷害直接是施法者等級），寒冰錐則是「擲完再加等級」。

**骰數不一定是立即數**。魔法飛彈是 `(等級＋1) div 2` 顆、火球是 `等級d6`、
寒冰錐是 `等級d4 ＋ 等級`——這些的骰數是算出來的，抽出來只能標 `computed`。
`entry#9` 對骰數 0 的處理是**直接回 0**（`cmp [bp+8],0 / jbe 出口`），所以
把「抽到 0」解讀成「用施法者等級」會在任何真的傳 0 的地方變成錯的。

用法：python3 scripts/spell_damage_table.py [--write]
"""

import json
import os
import re
import subprocess
import sys

ROOT = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
OVERLAY = os.path.join(ROOT, "workplace/re-sweep/dos/overlays/overlay-22.bin")
MANIFEST = os.path.join(ROOT, "workplace/re-sweep/dos/ovr-manifest.json")
SPELLS = os.path.join(ROOT, "gamepack/rules/spell-table.json")
OUT_MD = os.path.join(ROOT, "docs/audit/spell-damage-table.md")
OUT_JSON = os.path.join(ROOT, "gamepack/rules/spell-damage.json")

# 法術分派表的基底與筆數（DOS）。
DISPATCH_BASE = 0x72A0
DISPATCH_COUNT = 101
# 段 → 檔案段落的差，與 `cmd/far-call-map` 量到的同一個常數。
SEGMENT_DELTA = 0x7B
STUB_TABLE_OFFSET = 0x20
STUB_SIZE = 5
# 兩支擲骰入口（`overlay-23` 的 entry stub 位移）。名字照 entry 編號，
# **不照用途**：`entry#9` 既被治療輕傷用，也被解除魔法拿去擲 `1d100`，
# 叫它 `heal` 會在解除魔法那一列變成錯的斷言。
DICE_CALLS = {"call   0x141:0x4d": "entry9", "call   0x141:0x52": "entry10"}
# 兩種收尾。傷害那一支是 `sub_F06(法術編號, 等級覆寫, ?, 傷害, 屬性旗標, 訊息)`；
# 治療那一支是 `<overlay-23 entry#22>(…, 治療量, 0)`。**兩種都要看**：只看傷害
# 那一支的話，治療輕／中／重三支的骰子會整批掉成「沒有傷害」。
DAMAGE_FINISH = "call   0xf06"
HEAL_FINISH = "call   0x141:0x8e"


def disassemble(start, stop):
    return subprocess.run(
        ["objdump", "-D", "-b", "binary", "-m", "i8086", "-M", "intel",
         "--start-address=%d" % start, "--stop-address=%d" % stop, OVERLAY],
        capture_output=True, text=True).stdout


def instructions(text, stop_at_return=True):
    body = []
    for line in text.splitlines():
        found = re.match(r"\s+([0-9a-f]+):\s+([0-9a-f ]+)\t(.*)", line)
        if not found:
            continue
        instruction = found.group(3).strip()
        body.append((int(found.group(1), 16), instruction))
        if stop_at_return and instruction.startswith("retf"):
            break
    return body


def dispatch_table():
    """法術編號 → handler 的 code offset。"""
    body = instructions(disassemble(0, os.path.getsize(OVERLAY)), stop_at_return=False)
    registers, slots = {}, {}
    for _, text in body:
        found = re.match(r"^mov\s+(ax|dx),0x([0-9a-f]+)$", text)
        if found:
            registers[found.group(1)] = int(found.group(2), 16)
            continue
        found = re.match(r"^mov\s+(?:WORD PTR )?ds:0x([0-9a-f]{4}),(ax|dx)$", text)
        if not found:
            continue
        address = int(found.group(1), 16)
        if not DISPATCH_BASE <= address < DISPATCH_BASE + DISPATCH_COUNT * 4:
            continue
        index = (address - DISPATCH_BASE) // 4
        half = "offset" if (address - DISPATCH_BASE) % 4 == 0 else "segment"
        slots.setdefault(index, {})[half] = registers.get(found.group(2))
    manifest = json.load(open(MANIFEST, encoding="utf-8"))
    handlers = {}
    for index, slot in slots.items():
        if "offset" not in slot or "segment" not in slot:
            continue
        stub = slot["offset"] - STUB_TABLE_OFFSET
        if stub < 0 or stub % STUB_SIZE:
            continue
        entry = stub // STUB_SIZE
        control = (slot["segment"] + SEGMENT_DELTA) * 16
        for overlay in manifest["overlays"]:
            if overlay["executable_offset"] != control:
                continue
            for record in overlay["entries"]:
                if record["index"] == entry:
                    handlers[index] = {"overlay": overlay["index"], "entry": entry,
                                       "offset": record["code_offset"]}
    if not handlers:
        raise SystemExit("法術分派表解不出任何一筆——基底或段差變了")
    return handlers


def push_source(body, position):
    """把「這個 push ax 的值哪來的」分類。

    位置在 `push ax` 上。前一條是 `mov al,立即數` 就是常數；否則就是算出來的
    （區域變數、除法、加法…），只能記下那一條的原文。
    ⚠ 這一層是必要的：骰數常常**不是**立即數（火球 `等級d6`、魔法飛彈
    `(等級＋1) div 2` 顆），把附近撿到的立即數當成骰數會得到自洽但錯的表。
    """
    if position <= 0:
        return {"kind": "unknown"}
    text = body[position - 1][1]
    found = re.match(r"^mov\s+al,0x([0-9a-f]+)$", text)
    if found:
        return {"kind": "literal", "value": int(found.group(1), 16)}
    if text in DICE_CALLS:
        return {"kind": "dice"}
    if re.match(r"^mov\s+al,BYTE PTR ds:0x([0-9a-f]+)$", text):
        return {"kind": "computed", "via": text}
    return {"kind": "computed", "via": text}


# 擲骰之後、推進收尾之前，只調整 ax 而不改變「這個值哪來的」的那些指令。
TRANSPARENT = re.compile(r"^(xor\s+ah,ah|cbw|cwd|inc\s+ax|add\s+ax,0x[0-9a-f]+"
                         r"|mov\s+(dx|bx),ax|pop\s+(dx|bx)|add\s+ax,(dx|bx))$")


def value_origin(body, position):
    """收尾拿到的那個值是哪來的。

    ★ 要**跨過**加值那幾條再判斷。治療中傷是 `2d8` 之後 `inc ax`，只看前一條
    會判成「算出來的」，那三支治療法術的骰子就整批掉出表外。
    """
    index = position - 1
    while index > 0 and TRANSPARENT.match(body[index][1]):
        index -= 1
    return push_source(body, index + 1)


def bonus_between(body, start, stop):
    """擲骰之後、把值推進收尾之前，加了多少。

    ★ 這一段不能省。致重傷是 `3d8` 之後 `add ax,3`、致中傷是 `inc ax`、
    電擊觸手是 `add ax,dx`（dx 是施法者等級）。只記骰子的話這三支的傷害
    每一支都偏低，而且低得剛好像是「合理的數字」——看不出錯。
    """
    total = 0
    for _, text in body[start + 1:stop]:
        if text in ("xor    ah,ah", "cbw", "cwd"):
            continue
        if text == "inc    ax":
            total += 1
            continue
        found = re.match(r"^add\s+ax,0x([0-9a-f]+)$", text)
        if found:
            total += int(found.group(1), 16)
            continue
        if re.match(r"^(mov|push|pop)\s+", text):
            if re.match(r"^mov\s+(dx|bx|cx),ax$", text) or re.match(r"^(push|pop)\s+(ax|dx|bx)$", text):
                continue
        if text.startswith("add    ax,") or text.startswith("mul") or text.startswith("shl"):
            return {"kind": "computed", "via": text}
        # 其餘（讀區域變數、算別的東西）都不影響 ax，忽略。
    return {"kind": "literal", "value": total}


def read_handler(offset):
    """讀一支法術 handler：擲骰、收尾傳出去的傷害／治療量，以及傷害屬性旗標。

    ★ 值從**收尾的呼叫點**取，不是從擲骰取。燃燒之手根本沒擲骰（傷害就是
    施法者等級），寒冰錐是「擲完再加等級」，致中傷是「擲完 ＋1」——只看擲骰
    的話第一支整支漏掉，後兩支數字偏低。

    ⚠ 也不能只取第一次擲骰：火球在分支裡先擲了 `1d3`，那一次不是傷害。
    """
    body = instructions(disassemble(offset, offset + 0x300))
    calls = []
    for position, (_, text) in enumerate(body):
        shape = DICE_CALLS.get(text)
        if not shape:
            continue
        pushes = [index for index in range(position) if body[index][1] == "push   ax"][-2:]
        if len(pushes) != 2:
            continue
        calls.append({"shape": shape, "at": position,
                      "count": push_source(body, pushes[0]),
                      "sides": push_source(body, pushes[1])})

    value, element, outcome, taken = {"kind": "absent"}, None, "none", None
    # 傷害收尾：`push cs / call 0xf06`。往回找字串那一段就定位得出兩個參數：
    #   [push 傷害] [mov al,旗標] [push ax] [lea di,區域] [push ss] [push di] …
    finish = [index for index, (_, text) in enumerate(body) if text == DAMAGE_FINISH]
    if finish:
        head = [index for index in range(finish[0])
                if body[index][1].startswith("lea    di,[bp")]
        if head and head[-1] >= 3 and body[head[-1] - 1][1] == "push   ax":
            anchor = head[-1]
            element_source = push_source(body, anchor - 1)
            if element_source["kind"] == "literal":
                element = element_source["value"]
            if body[anchor - 3][1] == "push   ax":
                value, outcome, taken = value_origin(body, anchor - 3), "damage", anchor - 3
    else:
        # 治療收尾：`push 治療量 / mov al,0 / push ax / <entry#22>`。
        finish = [index for index, (_, text) in enumerate(body) if text == HEAL_FINISH]
        if finish and finish[0] >= 3 and body[finish[0] - 3][1] == "push   ax":
            value = value_origin(body, finish[0] - 3)
            outcome, taken = "heal", finish[0] - 3

    bonus = {"kind": "literal", "value": 0}
    if value["kind"] == "dice" and taken is not None and calls:
        bonus = bonus_between(body, calls[-1]["at"], taken)
    return calls, value, bonus, element, outcome, len(body)


def main():
    handlers = dispatch_table()
    table = json.load(open(SPELLS, encoding="utf-8"))
    records = {}
    for spell in table["spells"]:
        if spell["effect_id"] != 0 or spell["placeholder"]:
            continue
        handler = handlers.get(spell["spell_id"])
        if not handler:
            continue
        calls, value, bonus, element, outcome, length = read_handler(handler["offset"])
        record = {"name": spell["name"], "level": spell["level"],
                  "caster_class": spell["caster_class"],
                  "camp_only": spell["camp_only"], "instructions": length,
                  "overlay": handler["overlay"], "entry": handler["entry"],
                  "offset": handler["offset"], "outcome": outcome,
                  "value_source": value["kind"], "element": element,
                  "dice_calls": [{key: call[key] for key in ("shape", "count", "sides")}
                                 for call in calls]}
        if value["kind"] == "literal":
            record["flat_value"] = value["value"]
        elif value["kind"] == "computed":
            record["value_via"] = value.get("via")
        # 形狀：只有「值就是那一次擲骰、骰數面數與加值全是立即數」才給數字。
        if value["kind"] == "dice" and len(calls) == 1 and bonus["kind"] == "literal":
            count, sides = calls[0]["count"], calls[0]["sides"]
            if count["kind"] == "literal" and sides["kind"] == "literal":
                record.update({"shape": calls[0]["shape"],
                               "dice_count": count["value"],
                               "dice_sides": sides["value"],
                               "bonus": bonus["value"]})
            else:
                record["shape"] = "computed"
        elif value["kind"] == "literal":
            record["shape"] = "flat"
        elif value["kind"] == "absent":
            # 有擲骰但收尾不是那兩支標準收尾（屠殺活物走豁免、火焰護盾是反擊、
            # 解除魔法是對抗）。**不要**把這裡的骰子當成傷害填進去。
            record["shape"] = "other_finish"
        else:
            record["shape"] = "computed"
        records[spell["spell_id"]] = record

    usable = sum(1 for item in records.values() if item["shape"] in ("entry9", "entry10"))
    computed = sum(1 for item in records.values() if item["shape"] == "computed")
    flat = sum(1 for item in records.values() if item["shape"] == "flat")
    print("`+0Ah = 0` 的法術 %d 支：數字可直接用 %d 支、算出來的 %d 支、"
          "固定值 %d 支、收尾不是標準那兩支 %d 支"
          % (len(records), usable, computed, flat,
             len(records) - usable - computed - flat))
    if "--write" not in sys.argv:
        print("（預覽模式；加 --write 才寫報表）")
        return

    payload = {"schema_version": 2,
               "source": "DOS overlay-22 的法術 handler（位址取自 DS:72A0h 的分派表）",
               "spec": "docs/spec/1124-spell-damage-dice.md",
               "spells": {str(key): value for key, value in sorted(records.items())}}
    with open(OUT_JSON, "w", encoding="utf-8") as handle:
        json.dump(payload, handle, ensure_ascii=False, indent=1)
        handle.write("\n")

    lines = ["# 傷害／治療類法術的骰子", "",
             "由 `scripts/spell_damage_table.py` 產生，判讀見 "
             "[spec 1124](../spec/1124-spell-damage-dice.md)。不要手改。", "",
             "屬性表的 `+0Ah` 為 0 的法術，傷害不在表裡而在各自的 handler。"
             "數字取自**收尾的呼叫點**（傷害走 `sub_F06`、治療走 `<overlay-23 "
             "entry#22>`），不是取自擲骰——燃燒之手沒擲骰，治療中傷擲完還加 1。", "",
             "形狀欄位：`entry9`／`entry10` 是**骰數、面數、加值全是立即數**的那一種，"
             "數字可以直接用；`computed` 是有一段由程式算出來（火球 `等級d6`、"
             "魔法飛彈 `(等級＋1) div 2` 顆、寒冰錐 `等級d4 ＋ 等級`、"
             "電擊觸手 `1d8 ＋ 等級`），要人去讀那一支；`flat` 是不擲骰的固定值；"
             "`other_finish` 是有擲骰但收尾不是那兩支（豁免即死、反擊、對抗）。", "",
             "| 編號 | 名稱 | 環 | 職業 | 收尾 | 形狀 | 值 | 屬性旗標 | handler |",
             "|---:|---|---:|---|---|---|---|---|---|"]
    for key in sorted(records):
        item = records[key]
        if item["shape"] in ("entry9", "entry10"):
            dice = "`%dd%d%s`" % (item["dice_count"], item["dice_sides"],
                                  " ＋ %d" % item["bonus"] if item["bonus"] else "")
        elif item["shape"] == "flat":
            dice = "`%d`" % item["flat_value"]
        else:
            dice = "、".join(
                "`%sd%s`" % (describe(call["count"]), describe(call["sides"]))
                for call in item["dice_calls"]) or "（不擲骰）"
        element = item.get("element")
        lines.append("| %d | %s%s | %d | %s | %s | %s | %s | %s | `overlay-%d entry#%d` |" % (
            key, item["name"], "（紮營）" if item["camp_only"] else "",
            item["level"], item["caster_class"], item["outcome"], item["shape"], dice,
            "`%02Xh`" % element if element is not None else "—",
            item["overlay"], item["entry"]))
    open(OUT_MD, "w", encoding="utf-8").write("\n".join(lines) + "\n")
    print("寫出 %s 與 %s" % (OUT_MD, OUT_JSON))


def describe(source):
    if source["kind"] == "literal":
        return str(source["value"])
    return "?"


if __name__ == "__main__":
    main()
