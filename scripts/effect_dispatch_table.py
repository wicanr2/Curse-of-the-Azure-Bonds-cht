"""把效果分派表（DOS `DS:6FA6h` ／ PC-98 `DS:A040h`）解碼成可查的對照表。

表的內容不是資料，是 `overlay-12` 的初始化函式（DOS `02E3Ch`、PC-98 `02ED4h`）
用一連串 `mov ds:xxxx, ax` ／ `mov ds:xxxx, dx` 寫進去的 **far pointer**。
每個 far pointer 指向 Borland overlay 的 5-byte entry stub（`CD 3F` ＋ 進入點
位移 ＋ 旗標），所以 `段:位移` 可以直接對回 `ovr-manifest.json` 裡的
`executable_offset`，再換成「哪個 overlay 的第幾個 entry」。

換算：IDA 的載入基底是 `1000h`，
`段:位移` → 線性 `(1000h ＋ 段) × 10h ＋ 位移`；
stub 的線性位址 ＝ `10000h ＋ executable_offset − MZ 標頭長度`。

消費端是 `overlay-23:00C9h`（PC-98 符號名 `CALLEFFECT`，spec 576）：
`call dword ptr [di + 6FA6h]`，`di ＝ 效果碼 × 4` ⇒ **效果碼是 1-based**。

用法：
    python3 scripts/effect_dispatch_table.py [--write]
"""

import json
import os
import re
import struct
import subprocess
import sys

ROOT = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
SWEEP = os.path.join(ROOT, "workplace", "re-sweep")
LEDGER = os.path.join(ROOT, "docs", "audit", "re-function-ledger.json")
OUT = os.path.join(ROOT, "docs", "audit", "effect-dispatch-table.md")

# 平台 → (可執行檔, 初始化函式所在模組, 初始化函式 ea, 表基底（效果碼 0 的位置）)
PLATFORMS = {
    "dos": ("START.EXE", "overlay-12", "02E3C", 0x6FA6),
    "pc98": ("PC98-GAME.EXE", "overlay-12", "02ED4", 0xA040),
}
# 第二支會寫這張表的初始化：`INITSPELLS` 的尾段（spec 1202）。
# 它把七個特殊攻擊效果碼指到 overlay-22 尾端的 handler。
SECOND_WRITER = {
    "dos": ("overlay-22", "0630B"),
    "pc98": ("overlay-22", "066EB"),
}


def stub_index(platform, executable):
    """線性位址 → (模組, entry 編號, 進入點 ea)。"""
    blob = open(os.path.join(SWEEP, platform, executable), "rb").read()
    header = struct.unpack_from("<H", blob, 8)[0] * 16
    manifest = json.load(open(os.path.join(SWEEP, platform, "ovr-manifest.json"),
                              encoding="utf-8"))
    out = {}
    for overlay in manifest["overlays"]:
        for entry in overlay["entries"]:
            linear = 0x10000 + entry["executable_offset"] - header
            out[linear] = (overlay["module"], entry["index"],
                           entry["code_offset"])
    return out


def slots(platform, module, ea):
    """從初始化函式的 dump 抽出「槽位 → (段, 位移)」。"""
    dump = subprocess.run(
        [sys.executable, os.path.join(ROOT, "scripts", "annotated_dump.py"),
         platform, module, ea],
        capture_output=True, text=True, check=True).stdout
    table, seg, off = {}, None, None
    for line in dump.splitlines():
        m = re.search(r"mov\s+ax,\s*([0-9A-Fa-f]+)h", line)
        if m and "ds:" not in line:
            off = int(m.group(1), 16)
            continue
        m = re.search(r"mov\s+dx,\s*([0-9A-Fa-f]+)h", line)
        if m and "ds:" not in line:
            seg = int(m.group(1), 16)
            continue
        m = re.search(r"mov\s+ds:([0-9A-Fa-f]+)h,\s*ax", line)
        if m:
            table[int(m.group(1), 16)] = (seg, off)
    return table


def decode(platform):
    executable, module, ea, base = PLATFORMS[platform]
    stubs = stub_index(platform, executable)
    raw = slots(platform, module, ea)
    out = {}
    for slot, (seg, off) in raw.items():
        code = (slot - base) // 4
        linear = (0x1000 + seg) * 0x10 + off
        out[code] = stubs[linear]
    second_module, second_ea = SECOND_WRITER[platform]
    second = {}
    for slot, (seg, off) in slots(platform, second_module, second_ea).items():
        if not base <= slot:
            continue
        code = (slot - base) // 4
        if code > (max(raw) - base) // 4 + 8:
            continue
        linear = (0x1000 + seg) * 0x10 + off
        if linear in stubs:
            second[code] = stubs[linear]
    return out, second, base, max(raw)


def ledger():
    rows = {}
    for row in json.load(open(LEDGER, encoding="utf-8"))["functions"]:
        rows[(row["platform"], row["module"], row["ea"])] = row
    return rows


def main():
    rows = ledger()
    dos, dos2, dos_base, dos_last = decode("dos")
    pc98, pc982, pc98_base, _ = decode("pc98")
    codes = range(1, (dos_last - dos_base) // 4 + 1)

    lines = [
        "# 效果分派表（DOS `DS:6FA6h` ／ PC-98 `DS:A040h`）",
        "",
        "本檔由 `scripts/effect_dispatch_table.py` 產生，判讀見 "
        "`docs/spec/1005-effect-dispatch-table.md`。",
        "",
        "分派端 `overlay-23:00C9h`（`CALLEFFECT`，spec 576）以 "
        "**效果碼 × 4** 索引，效果碼從 **1** 起算。",
        "",
        "| 效果碼 | DOS 處理常式 | PC-98 處理常式 | 台帳 | 判讀 |",
        "|---|---|---|---|---|",
    ]
    for code in codes:
        a, b = dos.get(code), pc98.get(code)
        a2, b2 = dos2.get(code), pc982.get(code)
        if a is None and a2 is None and b is None and b2 is None:
            lines.append(f"| {code} | *（未指定，維持 NIL）* | "
                         f"*（未指定，維持 NIL）* | — | — |")
            continue
        note = ""
        if a is not None and a2 is not None:
            # 雙重寫入：本支之外 INITSPELLS 也寫了這一格；生效者看執行順序
            # （spec 1202 的宣稱邊界）。
            note = (f" ⚠ INITSPELLS 又寫成 `{a2[0]}:{a2[2]:05X}h`"
                    f"（entry#{a2[1]}），生效者未判定（spec 1202）")
        primary = a or a2
        primary_b = b or b2
        row = rows.get(("dos", primary[0], primary[2])) if primary else None
        state = row["state"] if row else "無列"
        spec = (row or {}).get("spec", "") or ""
        spec = os.path.basename(spec).split("-")[0] if spec else ""
        source = "" if a is not None else "（INITSPELLS，spec 1202）"
        lines.append(
            f"| {code} | `{primary[0]}:{primary[2]:05X}h`（entry#{primary[1]}）{source}{note} | "
            f"`{primary_b[0]}:{primary_b[2]:05X}h`（entry#{primary_b[1]}）{source} | {state} | {spec} |")

    text = "\n".join(lines) + "\n"
    if "--write" in sys.argv:
        open(OUT, "w", encoding="utf-8").write(text)
        print(f"寫入 {OUT}（{len(list(codes))} 個效果碼）")
    else:
        print(text)


if __name__ == "__main__":
    main()
