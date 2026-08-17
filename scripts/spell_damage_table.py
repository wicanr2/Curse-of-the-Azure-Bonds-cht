"""抽出「傷害／治療類」法術的骰子。

法術屬性表的 `+0Ah`（效果碼）為 0 的那一批，傷害不在表裡——它在**各自的
handler** 裡。handler 的位址來自法術分派表：`DS:72A0h` 起 101 筆遠指標，
由 `overlay-22` 的一支初始化函式填進去（PC-98 對側見 spec 1071）。
段固定是一個 overlay 的 entry stub 段，所以 `位移 → (位移 − 20h) ÷ 5` 就是
entry 編號，再查 `ovr-manifest.json` 得到 handler 的 code offset。

handler 裡擲骰有兩種形狀，兩種都是 `(骰數, 面數)` 依序推入：

    <overlay-23 entry#9> (目標, 骰數, 面數)
    <overlay-23 entry#10>(來源, 0, 0, 骰數, 面數)

兩支的用途**不是**「治療 vs 傷害」：`entry#9` 既被治療輕傷用，也被解除魔法拿去
擲 `1d100`。所以形狀欄位照 entry 編號記，用途要看該支法術自己。

**骰數 0 代表「用施法者等級當骰數」**（火球 `0d6`、魔法飛彈 `0d4`）。

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


def read_dice(offset):
    """回傳 handler 裡**每一次**擲骰。

    ⚠ 不能只取第一次。像火球那種 handler 會先擲別的東西（範圍、分支），
    第一次擲骰不是傷害；只取第一次會產生「看起來有答案但是錯的」那一列。
    """
    body = instructions(disassemble(offset, offset + 0x300))
    immediates, calls = [], []
    for position, (_, text) in enumerate(body):
        found = re.match(r"^mov\s+al,0x([0-9a-f]+)$", text)
        if found:
            immediates.append((position, int(found.group(1), 16)))
            continue
        shape = DICE_CALLS.get(text)
        if not shape:
            continue
        previous = [value for index, value in immediates if index < position][-2:]
        if len(previous) == 2:
            calls.append({"shape": shape, "dice_count": previous[0],
                          "dice_sides": previous[1]})
    return calls, len(body)


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
        calls, length = read_dice(handler["offset"])
        record = {"name": spell["name"], "level": spell["level"],
                  "caster_class": spell["caster_class"],
                  "camp_only": spell["camp_only"], "instructions": length,
                  "overlay": handler["overlay"], "entry": handler["entry"],
                  "offset": handler["offset"], "dice_calls": calls}
        if len(calls) == 1:
            # 只有一次擲骰時它就是這支法術的骰子，沒有歧義。
            record.update({"shape": calls[0]["shape"],
                           "dice_count": calls[0]["dice_count"],
                           "dice_sides": calls[0]["dice_sides"],
                           "scales_with_caster_level": calls[0]["dice_count"] == 0})
        elif calls:
            # 多次擲骰要人去讀那一支才知道哪一次是傷害。
            record["shape"] = "ambiguous"
        else:
            record["shape"] = "unread"
        records[spell["spell_id"]] = record

    decoded = sum(1 for item in records.values()
                  if item["shape"] not in ("unread", "ambiguous"))
    ambiguous = sum(1 for item in records.values() if item["shape"] == "ambiguous")
    print("`+0Ah = 0` 的法術 %d 支：唯一一次擲骰 %d 支、多次擲骰 %d 支、沒擲骰 %d 支"
          % (len(records), decoded, ambiguous, len(records) - decoded - ambiguous))
    if "--write" not in sys.argv:
        print("（預覽模式；加 --write 才寫報表）")
        return

    payload = {"schema_version": 1,
               "source": "DOS overlay-22 的法術 handler（位址取自 DS:72A0h 的分派表）",
               "spec": "docs/spec/1124-spell-damage-dice.md",
               "spells": {str(key): value for key, value in sorted(records.items())}}
    with open(OUT_JSON, "w", encoding="utf-8") as handle:
        json.dump(payload, handle, ensure_ascii=False, indent=1)
        handle.write("\n")

    lines = ["# 傷害／治療類法術的骰子", "",
             "由 `scripts/spell_damage_table.py` 產生，判讀見 "
             "[spec 1124](../spec/1124-spell-damage-dice.md)。不要手改。", "",
             "屬性表的 `+0Ah` 為 0 的法術，傷害不在表裡而在各自的 handler。",
             "**骰數 0 代表用施法者等級當骰數**（魔法飛彈 `等級d4`）。", "",
             "`ambiguous` 是 handler 裡擲了不只一次骰——**第一次不一定是傷害**，"
             "要人去讀那一支才知道哪一次是。只取第一次會產生看起來有答案但是錯的列。", "",
             "| 編號 | 名稱 | 環 | 職業 | 形狀 | 骰 | handler |",
             "|---:|---|---:|---|---|---|---|"]
    for key in sorted(records):
        item = records[key]
        if item["shape"] == "unread":
            dice = "—"
        elif item["shape"] == "ambiguous":
            dice = "、".join("`%s%dd%d`" % ("等級" if call["dice_count"] == 0 else "",
                                            call["dice_count"], call["dice_sides"])
                             for call in item["dice_calls"])
        elif item["scales_with_caster_level"]:
            dice = "`等級d%d`" % item["dice_sides"]
        else:
            dice = "`%dd%d`" % (item["dice_count"], item["dice_sides"])
        lines.append("| %d | %s%s | %d | %s | %s | %s | `overlay-%d entry#%d` |" % (
            key, item["name"], "（紮營）" if item["camp_only"] else "",
            item["level"], item["caster_class"], item["shape"], dice,
            item["overlay"], item["entry"]))
    open(OUT_MD, "w", encoding="utf-8").write("\n".join(lines) + "\n")
    print("寫出 %s 與 %s" % (OUT_MD, OUT_JSON))


if __name__ == "__main__":
    main()
