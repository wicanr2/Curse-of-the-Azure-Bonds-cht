"""抽出「哪一支法術／效果播哪一段演出」。

原作的戰鬥演出只有一個入口：

    <overlay-24 entry#24>(槽)

它對同一個圖組槽連放四格（`entry#23(槽, half, block, mode)`，
`half` 走 0,0,1,1、`block` 走 0,1,2,3）。所以「這支法術長什麼樣」
＝ **它把哪個槽號傳進去**。

槽號怎麼對回圖檔：`overlay-11` 開場跑一個迴圈

    for i := 0 to 11 do 載入圖組('COMSPR', i, i + 13);
    載入圖組('COMSPR', 25, 25);

也就是 **`槽 = COMSPR 區塊 + 13`**（區塊 0..11 → 槽 13..24），另外槽 25 單獨載
區塊 25。載入端每個槽同時放「編號」與「編號 ＋ 80h」兩份（spec 1033），
四格演出就是在這兩份之間切換。

★ 這條換算把既有規格裡的「icon 17h」翻譯得出來：`17h = 23`，`23 − 13 = 10`，
正是 COMSPR `0Ah`／`8Ah` 那一組魔法命中圖。

用法：python3 scripts/spell_visual_table.py [--write]
"""

import json
import os
import re
import subprocess
import sys

ROOT = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
OVDIR = os.path.join(ROOT, "workplace/re-sweep/dos/overlays")
MANIFEST = os.path.join(ROOT, "workplace/re-sweep/dos/ovr-manifest.json")
SPELLS = os.path.join(ROOT, "gamepack/rules/spell-table.json")
OUT_MD = os.path.join(ROOT, "docs/audit/spell-visual-table.md")
OUT_JSON = os.path.join(ROOT, "gamepack/rules/spell-visuals.json")

# 演出入口：`overlay-24 entry#24`。段 14Dh、stub 位移 20h + 24×5 = 98h。
ANIMATE_CALL = "call   0x14d:0x98"
# 槽 → COMSPR 區塊的位移，由 `overlay-11` 的載入迴圈量出來。
SLOT_BASE = 13
STANDALONE_SLOT = 25

DISPATCH_BASE = 0x72A0
DISPATCH_COUNT = 101
# 效果分派表（CALLEFFECT，spec 1005）：基底 6FA6h、效果碼從 1 起算。
# `INITSPELLS`（630Bh）尾端會往這張表寫七格——就是先前「未歸屬」那七支。
EFFECT_DISPATCH_BASE = 0x6FA6
EFFECT_DISPATCH_COUNT = 147
SEGMENT_DELTA = 0x7B
STUB_TABLE_OFFSET = 0x20
STUB_SIZE = 5
# overlay-22 自己的段（`call 0x141:...` 之類要能分辨是不是打回自己）。
OVERLAY22_SEGMENT = 0x11A
# 呼叫圖往下追幾層。實測 handler → 演出常式最多兩層。
CALL_DEPTH = 3


def instructions(overlay, start=0, stop=None):
    path = os.path.join(OVDIR, "overlay-%d.bin" % overlay)
    argv = ["objdump", "-D", "-b", "binary", "-m", "i8086", "-M", "intel"]
    if start:
        argv.append("--start-address=%d" % start)
    if stop:
        argv.append("--stop-address=%d" % stop)
    out = subprocess.run(argv + [path], capture_output=True, text=True).stdout
    body = []
    for line in out.splitlines():
        found = re.match(r"\s+([0-9a-f]+):\s+([0-9a-f ]+)\t(.*)", line)
        if found:
            body.append((int(found.group(1), 16), found.group(3).strip()))
    return body


def entry_ranges(overlay):
    """entry 編號 → (起, 迄)。迄取下一個 entry 的起點。"""
    manifest = json.load(open(MANIFEST, encoding="utf-8"))
    for item in manifest["overlays"]:
        if item["index"] != overlay:
            continue
        entries = sorted((record["code_offset"], record["index"])
                         for record in item["entries"]
                         if record["code_offset"] not in (0xFFFF,))
        ranges = {}
        for position, (offset, index) in enumerate(entries):
            stop = entries[position + 1][0] if position + 1 < len(entries) else 1 << 30
            ranges[index] = (offset, stop)
        return ranges
    raise SystemExit("找不到 overlay-%d" % overlay)


def animation_sites(overlay):
    """回傳 [(位址, 槽)]；槽不是立即數就記 None。"""
    body = instructions(overlay)
    sites = []
    for position, (address, text) in enumerate(body):
        if text != ANIMATE_CALL:
            continue
        slot = None
        for _, previous in reversed(body[max(0, position - 6):position]):
            found = re.match(r"^mov\s+al,0x([0-9a-f]+)$", previous)
            if found:
                slot = int(found.group(1), 16)
                break
            if previous.startswith("mov    al,"):
                break
        sites.append((address, slot))
    return sites


def call_graph(overlay, ranges):
    """overlay 內部的呼叫圖：函式 → 它呼叫的函式。

    ★ 法術 handler 幾乎都**不直接**播演出，而是叫 overlay-22 尾端那幾支共用的
    演出常式。只看 handler 本體會得到「沒有任何法術有演出」——那是把中間那一層
    當成不存在。
    """
    body = instructions(overlay)
    owner_of = {}
    for index, (start, stop) in ranges.items():
        owner_of[index] = (start, stop)

    def owner(address):
        for index, (start, stop) in owner_of.items():
            if start <= address < stop:
                return index
        return None

    graph = {}
    for address, text in body:
        source = owner(address)
        if source is None:
            continue
        found = re.match(r"^call\s+0x([0-9a-f]+)$", text)
        if found:
            target = owner(int(found.group(1), 16))
            if target is not None and target != source:
                graph.setdefault(source, set()).add(target)
            continue
        found = re.match(r"^call\s+0x([0-9a-f]+):0x([0-9a-f]+)$", text)
        if found:
            segment, offset = int(found.group(1), 16), int(found.group(2), 16)
            stub = offset - STUB_TABLE_OFFSET
            if stub < 0 or stub % STUB_SIZE:
                continue
            # 只收「打回自己這個 overlay」的遠呼叫；別的 overlay 不在這張圖裡。
            if segment == OVERLAY22_SEGMENT:
                graph.setdefault(source, set()).add(stub // STUB_SIZE)
    return graph


def reachable(graph, start, depth):
    seen, frontier = {start}, {start}
    for _ in range(depth):
        nxt = set()
        for node in frontier:
            nxt |= graph.get(node, set())
        nxt -= seen
        if not nxt:
            break
        seen |= nxt
        frontier = nxt
    return seen


def _pointer_stores(base, count):
    """掃 overlay-22 對 `[base, base+count*4)` 的 far pointer 賦值，回 index → entry。"""
    body = instructions(22)
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
        if not base <= address < base + count * 4:
            continue
        index = (address - base) // 4
        half = "offset" if (address - base) % 4 == 0 else "segment"
        slots.setdefault(index, {})[half] = registers.get(found.group(2))
    handlers = {}
    for index, slot in slots.items():
        if "offset" not in slot or "segment" not in slot:
            continue
        stub = slot["offset"] - STUB_TABLE_OFFSET
        if stub < 0 or stub % STUB_SIZE:
            continue
        handlers[index] = stub // STUB_SIZE
    return handlers


def dispatch_table():
    """法術編號 → entry。與 spell_damage_table.py 同一條路。"""
    handlers = _pointer_stores(DISPATCH_BASE, DISPATCH_COUNT)
    if not handlers:
        raise SystemExit("法術分派表解不出任何一筆")
    return handlers


def effect_assignments():
    """效果碼 → entry：`INITSPELLS` 對 CALLEFFECT 分派表（6FA6h）的賦值。"""
    return _pointer_stores(EFFECT_DISPATCH_BASE, EFFECT_DISPATCH_COUNT)


def comspr_block(slot):
    if slot is None:
        return None
    if slot == STANDALONE_SLOT:
        return STANDALONE_SLOT
    return slot - SLOT_BASE


def main():
    ranges22 = entry_ranges(22)
    handlers = dispatch_table()
    by_entry = {}
    for spell_id, entry in handlers.items():
        by_entry.setdefault(entry, []).append(spell_id)

    table = json.load(open(SPELLS, encoding="utf-8"))
    names = {item["spell_id"]: item for item in table["spells"]}

    graph = call_graph(22, ranges22)
    sites = {}
    for address, slot in animation_sites(22):
        owner = None
        for entry, (start, stop) in ranges22.items():
            if start <= address < stop:
                owner = entry
                break
        sites.setdefault(owner, []).append(
            {"address": address, "slot": slot, "comspr_block": comspr_block(slot)})

    # `CASTSPELL` 那一支（分派表就在它裡面）播的是每支法術都會有的演出。
    cast_entry = None
    for entry, (start, stop) in ranges22.items():
        if start <= 0x1563 < stop:
            cast_entry = entry
            break
    shared = list(sites.get(cast_entry, []))

    per_spell = {}
    for spell_id, entry in sorted(handlers.items()):
        found = []
        for node in sorted(reachable(graph, entry, CALL_DEPTH)):
            if node == cast_entry:
                continue
            for item in sites.get(node, []):
                found.append(dict(item, via=node))
        if found:
            per_spell[spell_id] = found

    # 剩下的演出點：既不在 `CASTSPELL`、也搆不到任何法術 handler。
    # ★ `INITSPELLS` 尾端另外把七支寫進 CALLEFFECT 分派表（6FA6h，spec 1005）
    # ——那就是「特殊攻擊」的效果 handler（吐酸、龍息、凝視、丟電光，
    # spec 1202）。先用它歸屬；真的誰都不指的才進 orphans。
    attributed = {cast_entry}
    for spell_id, entry in handlers.items():
        attributed |= reachable(graph, entry, CALL_DEPTH)
    effect_entries = effect_assignments()
    by_effect_entry = {}
    for code, entry in effect_entries.items():
        by_effect_entry.setdefault(entry, []).append(code)
    special = []
    orphans = []
    for entry in sorted(sites):
        if entry in attributed:
            continue
        codes = by_effect_entry.get(entry)
        for item in sites[entry]:
            if codes:
                special.append(dict(item, entry=entry, effect_codes=sorted(codes)))
            else:
                orphans.append(dict(item, entry=entry))

    effects = [{"address": address, "slot": slot, "comspr_block": comspr_block(slot)}
               for address, slot in animation_sites(12)]

    print("共用施法路 %d 處、法術自己的 %d 支、效果 handler %d 處、"
          "特殊攻擊 %d 處、未歸屬 %d 處"
          % (len(shared), len(per_spell), len(effects), len(special), len(orphans)))
    used = sorted({item["slot"] for item in shared + effects
                   + [row for rows in per_spell.values() for row in rows]
                   if item["slot"] is not None} if True else set())
    print("用到的槽：%s（COMSPR 區塊 %s）"
          % (used, [comspr_block(slot) for slot in used]))
    if "--write" not in sys.argv:
        print("（預覽模式；加 --write 才寫報表）")
        return

    payload = {"schema_version": 1,
               "source": "DOS overlay-22／overlay-12 的 <overlay-24 entry#24> 呼叫點",
               "spec": "docs/spec/1126-spell-visual-slots.md",
               "slot_base": SLOT_BASE,
               "shared": shared,
               "special_attacks": special,
               "unattributed": orphans,
               "effects": effects,
               "spells": {str(key): value for key, value in sorted(per_spell.items())}}
    with open(OUT_JSON, "w", encoding="utf-8") as handle:
        json.dump(payload, handle, ensure_ascii=False, indent=1)
        handle.write("\n")

    lines = ["# 法術演出用哪個圖組槽", "",
             "由 `scripts/spell_visual_table.py` 產生，判讀見 "
             "[spec 1126](../spec/1126-spell-visual-slots.md)。不要手改。", "",
             "演出入口是 `<overlay-24 entry#24>(槽)`，對同一個槽連放四格。",
             "**`槽 = COMSPR 區塊 + 13`**（`overlay-11` 開場的載入迴圈量出來的）。", "",
             "⚠ 這張表只涵蓋 `entry#24` 這條路。同一個 blit（`entry#23`）還有兩類"
             "使用者不在這裡：兩種雲的持續區域格（COMSPR 區塊 4／2）與弓箭八方向"
             "（區塊 0／1／2、128..130，`overlay-13` 的 11 個直接呼叫點）。", "",
             "## 共用施法路", "",
             "所有法術都會經過這裡，所以這幾格是**每一支法術都有的演出**。", "",
             "| overlay-22 位址 | 槽 | COMSPR 區塊 |", "|---|---:|---:|"]
    for item in shared:
        lines.append("| `%04Xh` | %s | %s |" % (
            item["address"],
            item["slot"] if item["slot"] is not None else "（算出來的）",
            item["comspr_block"] if item["comspr_block"] is not None else "—"))
    lines += ["", "## 法術自己多播的", "",
              "| 編號 | 名稱 | overlay-22 位址 | 槽 | COMSPR 區塊 |",
              "|---:|---|---|---:|---:|"]
    for spell_id in sorted(per_spell):
        entry = names.get(spell_id, {})
        for item in per_spell[spell_id]:
            lines.append("| %d | %s | `%04Xh` | %s | %s |" % (
                spell_id, entry.get("name", "—"), item["address"],
                item["slot"] if item["slot"] is not None else "（算出來的）",
                item["comspr_block"] if item["comspr_block"] is not None else "—"))
    lines += ["", "## 特殊攻擊效果 handler 裡的", "",
              "overlay-22 尾端這幾支由 `INITSPELLS` 尾段寫進 CALLEFFECT 分派表"
              "（`6FA6h`，效果碼從 1 起算），是怪物特殊攻擊的演出"
              "（吐酸／龍息／凝視／丟電光，spec 1202；各 handler 的語意見"
              " spec 720／722／723／725／735／847／981）。", "",
              "| 效果碼 | overlay-22 位址 | entry | 槽 | COMSPR 區塊 |",
              "|---|---|---:|---:|---:|"]
    for item in special:
        lines.append("| %s | `%04Xh` | %d | %s | %s |" % (
            "／".join("%d (0x%02X)" % (code, code) for code in item["effect_codes"]),
            item["address"], item["entry"],
            item["slot"] if item["slot"] is not None else "（算出來的）",
            item["comspr_block"] if item["comspr_block"] is not None else "—"))
    if orphans:
        lines += ["", "## 未歸屬的演出常式", "",
                  "兩張分派表都不指、也沒有任何 overlay 呼叫端。", "",
                  "| overlay-22 位址 | entry | 槽 | COMSPR 區塊 |", "|---|---:|---:|---:|"]
        for item in orphans:
            lines.append("| `%04Xh` | %d | %s | %s |" % (
                item["address"], item["entry"],
                item["slot"] if item["slot"] is not None else "（算出來的）",
                item["comspr_block"] if item["comspr_block"] is not None else "—"))
    lines += ["", "## 效果 handler 裡的", "",
              "這幾格由效果碼觸發，不是法術直接播的（`overlay-12`）。", "",
              "| overlay-12 位址 | 槽 | COMSPR 區塊 |", "|---|---:|---:|"]
    for item in effects:
        lines.append("| `%04Xh` | %s | %s |" % (
            item["address"],
            item["slot"] if item["slot"] is not None else "（算出來的）",
            item["comspr_block"] if item["comspr_block"] is not None else "—"))
    open(OUT_MD, "w", encoding="utf-8").write("\n".join(lines) + "\n")
    print("寫出 %s 與 %s" % (OUT_MD, OUT_JSON))


if __name__ == "__main__":
    main()
