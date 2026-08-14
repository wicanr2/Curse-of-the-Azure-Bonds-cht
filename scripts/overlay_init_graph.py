"""解出每個 overlay 的 unit 初始化鏈，產出模組相依圖。

Turbo Pascal 的 unit 初始化段編譯後放在模組的 `0000h`，形狀固定：

    55 89 E5            push bp / mov bp, sp
    9A off seg × N      依序呼叫每個相依 unit 的初始化段
    89 EC 5D CB         mov sp, bp / pop bp / retf

每個目標都是「另一個 overlay 的 `@0000h`」——這正是 unit 相依關係。所以
`0000h` 的 far call 清單就是**該模組直接 uses 的其他模組**。

判準刻意嚴格：`0000h` 起必須是 `55 89 E5`，之後是連續的 `9A`，再之後必須正好
是 `89 EC 5D CB`。有一項不符就不採信，列為「形狀不符」。目標若不是別的 overlay
的 `0000h`，也會單獨標出來——那代表這個假設在該處不成立。

**不要用 IDA 的匯出讀這段**：`89 EC` 常被吃掉一個 byte 而變成 `in al, dx`
（見 spec 736），連帶把後面的 far call 也錯開。本工具一律讀原始 bytes。

用法：
    python3 scripts/overlay_init_graph.py [--write]
"""

import json
import os
import struct
import sys

sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
import overlay_call_graph as graph

ROOT = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
SWEEP = os.path.join(ROOT, "workplace", "re-sweep")
OUT = os.path.join(ROOT, "docs", "audit", "overlay-init-graph.md")

PROLOGUE = bytes((0x55, 0x89, 0xE5))
EPILOGUE = bytes((0x89, 0xEC, 0x5D, 0xCB))


def scan(platform):
    table = graph.modules(platform)
    directory = os.path.join(SWEEP, platform, "overlays")
    result = {}
    for name in sorted(os.listdir(directory)):
        if not name.endswith(".bin"):
            continue
        module = name[:-4]
        blob = open(os.path.join(directory, name), "rb").read()
        if blob[:3] != PROLOGUE:
            result[module] = {"shape": "形狀不符", "calls": []}
            continue
        cursor, calls = 3, []
        while cursor + 5 <= len(blob) and blob[cursor] == 0x9A:
            offset, segment = struct.unpack_from("<HH", blob, cursor + 1)
            target = graph.resolve(table, segment, offset)
            calls.append({"segment": segment, "offset": offset, "target": target})
            cursor += 5
        shape = "ok" if blob[cursor:cursor + 4] == EPILOGUE else "尾巴不符"
        result[module] = {"shape": shape, "calls": calls}
    return result


def main():
    write = "--write" in sys.argv
    lines = ["# overlay unit 初始化鏈（模組相依圖）", "",
             "由 `scripts/overlay_init_graph.py` 從**原始 bytes** 解出。每個 overlay 的",
             "`0000h` 是它的 unit 初始化段，裡面依序呼叫每個相依 unit 的初始化段，因此",
             "這張表就是模組層級的 uses 關係。", ""]
    for platform in ("dos", "pc98"):
        result = scan(platform)
        bad = [m for m, r in result.items() if r["shape"] != "ok"]
        lines += ["## %s" % platform, "",
                  "形狀符合 %d 個模組；不符 %d 個（%s）。"
                  % (len(result) - len(bad), len(bad), "、".join(bad) or "無"), "",
                  "| 模組 | 直接相依 |", "|---|---|"]
        for module in sorted(result):
            row = result[module]
            if row["shape"] != "ok":
                lines.append("| `%s` | *（%s）* |" % (module, row["shape"]))
                continue
            names = []
            for call in row["calls"]:
                target = call["target"]
                if target and target["ea"] == 0:
                    names.append("`%s`" % target["module"])
                elif target:
                    names.append("`%s`+%04Xh ⚠" % (target["module"], target["ea"]))
                else:
                    names.append("`%04X:%04X` ⚠" % (call["segment"], call["offset"]))
            lines.append("| `%s` | %s |" % (module, "、".join(names) or "*（無）*"))
        lines.append("")
        print("%-5s 形狀符合 %d／%d" % (platform, len(result) - len(bad), len(result)))

    if not write:
        print("（預覽模式；加 --write 才寫檔）")
        return 0
    open(OUT, "w", encoding="utf-8").write("\n".join(lines))
    print("已寫入 %s" % OUT)
    return 0


if __name__ == "__main__":
    sys.exit(main())
