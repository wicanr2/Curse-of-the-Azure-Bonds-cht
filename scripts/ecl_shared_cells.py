#!/usr/bin/env python3
"""找出 ECL 與引擎共用的記憶體格子（spec 1096 的 RE-14 清冊）。

做法分三段，每一段都只用可驗證的事實，不猜語意：

1. 從 `docs/audit/ecl-event-catalog.json` 取出 ECL 側實際用到的絕對位址，
   依 spec 1096 的分區表換算成 bank 內位移 `(位址 − 區基底) × 2`。
2. 掃 `workplace/re-sweep/<平台>/overlays/filled/*.json`，對每個函式做
   單一暫存器的符號執行：追蹤 `les di, ds:4F99h/4F9Dh/4FA1h` 把 di 綁到
   哪一塊 bank，之後的 `es:[di+XXXXh]` 就歸屬該 bank 的該位移。
   遇到會改動 di 卻無法靜態求值的指令（`add di, ax` 等）就把 di 標成未知，
   **不繼續歸屬**——寧可少報，不要報錯。
3. 兩邊取交集：同時被 ECL 讀寫、又被引擎程式碼讀寫的位移就是共用格子。

用法：
    python3 scripts/ecl_shared_cells.py [dos|pc98] > docs/audit/ecl-shared-cells.md
"""

import collections
import glob
import json
import os
import re
import sys

# spec 1096 §二：位址範圍 → (區號, 基底, bank 指標)。區 3／4 不是 bank 變數。
ZONES = (
    (0x4B00, 0x4EFF, 0, "4F99h", "bank0"),
    (0x7C00, 0x7FFF, 1, "4F9Dh", "bank1"),
    (0x7A00, 0x7BFF, 2, "4FA1h", "bank2"),
)
# 引擎側的 bank 指標。DOS 三個都由 spec 1096 的寫入端直接讀到；
# PC-98 只有 bank0／bank1 有證據（spec 1095 的兩平台對照表），
# 第三塊尚未確認，寧可漏報也不填猜測值。
BANK_POINTERS = {
    "dos": {"4F99h": "bank0", "4F9Dh": "bank1", "4FA1h": "bank2"},
    "pc98": {"7F05h": "bank0", "7F09h": "bank1"},
}

LES_RE = re.compile(r"^les\s+di,\s*ds:([0-9A-Fa-f]+h)$")
MEM_RE = re.compile(r"es:\[di\+([0-9A-Fa-f]+)h\]")
DI_CLOBBER_RE = re.compile(r"^(add|sub|mov|lea|inc|dec|xor|or|and)\s+di\b")


def classify(address):
    """ECL 絕對位址 → (區號, bank 名稱, bank 內位移)；不是 bank 變數就回 None。"""
    for low, high, zone, _pointer, bank in ZONES:
        if low <= address <= high:
            return zone, bank, (address - low) * 2
    return None


def ecl_addresses(catalog_path):
    """ECL 側：位址 → {指令名稱: 次數}。"""
    catalog = json.load(open(catalog_path, encoding="utf-8"))
    used = collections.defaultdict(collections.Counter)
    for member in catalog["members"]:
        for block in member["blocks"]:
            for instruction in block["instructions"]:
                for operand in instruction.get("operands", []):
                    if operand.get("code") != "0x01" or "word" not in operand:
                        continue
                    used[int(operand["word"], 16)][instruction["name"]] += 1
    return used


def engine_accesses(platform):
    """引擎側：(bank, 位移) → [(模組, 函式 ea, 指令位址, 助憶碼)]。"""
    pointers = BANK_POINTERS[platform]
    found = collections.defaultdict(list)
    pattern = "workplace/re-sweep/%s/overlays/filled/*.json" % platform
    for path in sorted(glob.glob(pattern)):
        module = os.path.basename(path).replace("%s-" % platform, "").replace(".json", "")
        for function in json.load(open(path, encoding="utf-8"))["functions"]:
            bank = None                     # di 目前綁在哪一塊 bank；None = 未知
            for item in function.get("items", []):
                text = (item.get("disasm") or "").strip()
                les = LES_RE.match(text)
                if les:
                    bank = pointers.get(les.group(1))
                    continue
                if bank is not None and DI_CLOBBER_RE.match(text):
                    bank = None             # di 被改動且無法靜態求值，停止歸屬
                    continue
                if bank is None:
                    continue
                for offset in MEM_RE.findall(text):
                    found[(bank, int(offset, 16))].append(
                        (module, function["ea"], item.get("ea"), text))
    return found


def main():
    platform = sys.argv[1] if len(sys.argv) > 1 else "dos"
    ecl = ecl_addresses("docs/audit/ecl-event-catalog.json")
    engine = engine_accesses(platform)

    rows = []
    for address, opcodes in ecl.items():
        placed = classify(address)
        if placed is None:
            continue
        zone, bank, offset = placed
        rows.append((address, zone, bank, offset, sum(opcodes.values()),
                     sorted(opcodes), engine.get((bank, offset), [])))
    rows.sort(key=lambda row: -row[4])

    shared = [row for row in rows if row[6]]
    private = [row for row in rows if not row[6]]

    print("# ECL↔引擎共用格子清冊（%s）\n" % platform)
    print("由 `scripts/ecl_shared_cells.py` 產生，不要手改。")
    print("換算依 spec 1096 §二：bank 內位移 ＝ `(位址 − 區基底) × 2`。\n")
    print("- ECL 側變數位址：**%d 個**（%d 次存取）"
          % (len(rows), sum(row[4] for row in rows)))
    print("- **共用格子（引擎側也存取同一位移）：%d 個**（%d 次 ECL 存取）"
          % (len(shared), sum(row[4] for row in shared)))
    print("- ECL 私有（引擎側查無存取）：%d 個（%d 次）\n"
          % (len(private), sum(row[4] for row in private)))
    print("⚠ 引擎側掃描是保守的：`di` 一旦被無法靜態求值的指令改動就停止歸屬，")
    print("所以「查無存取」代表**這個方法沒找到**，不代表一定不存在。\n")

    print("## 共用格子（每一格都必須逐格對上）\n")
    print("| ECL 位址 | 區 | bank 位移 | ECL 存取 | ECL 指令 | 引擎側存取點 |")
    print("|---|---|---|---|---|---|")
    for address, zone, bank, offset, count, opcodes, sites in shared:
        where = "、".join(sorted({"`%s:%04X`" % (module, function)
                                  for module, function, _ea, _text in sites}))
        print("| `%04Xh` | %d | `%s^[%Xh]` | %d | %s | %s |"
              % (address, zone, bank, offset, count, "／".join(opcodes), where))

    print("\n## ECL 私有（自洽即可，優先度低）\n")
    print("| ECL 位址 | 區 | bank 位移 | ECL 存取 | ECL 指令 |")
    print("|---|---|---|---|---|")
    for address, zone, bank, offset, count, opcodes, _sites in private:
        print("| `%04Xh` | %d | `%s^[%Xh]` | %d | %s |"
              % (address, zone, bank, offset, count, "／".join(opcodes)))
    return 0


if __name__ == "__main__":
    sys.exit(main())
