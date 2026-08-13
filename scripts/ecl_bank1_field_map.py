"""抽出 ECL bank 1（`7C00h..7FFFh`）的「計算值」對照表。

bank 1 的讀取不是直接讀陣列：`ADDRESSVALUE` → 記憶體讀取 routine → bank 1
會先呼叫一個計算 routine，它是一條 `cmp ax, N` 鏈，把 `7C00h + N` 對映到
某個 record 的欄位偏移（`es:[di + off]`，base 由 `DS:9594h` 的 far pointer 提供）。
算不出來時回傳無效旗標，呼叫端才回退去讀陣列。

輸出每筆：ECL 位址、record 欄位偏移、存取寬度、是否符號延伸。

⚠ IDA 的反組譯文字會帶 `; 'r'` 這種註解，比對前一定要剝掉；不剝的話
`cmp ax, 72h ; 'r'` 這種會整筆漏掉（實測 52 個 cmp 只抽到 20 筆）。

用法：
    python3 scripts/ecl_bank1_field_map.py <dump.json> [bank_base_hex]
"""

import json
import re
import sys

CMP = re.compile(r"^cmp\s+ax,\s*([0-9A-Fa-f]+h|[0-9]+)$")
FIELD = re.compile(r"^mov\s+(al|ax),\s*es:\[di\+([0-9A-Fa-f]+h|[0-9]+)\]$")


def clean(text):
    text = re.sub(r"\s*;.*$", "", text.strip())
    return re.sub(r"\s+", " ", text)


def immediate(text):
    return int(text[:-1], 16) if text.endswith("h") else int(text)


def main():
    if len(sys.argv) < 2:
        print(__doc__)
        return 2
    base = int(sys.argv[2], 16) if len(sys.argv) > 2 else 0x7C00
    items = json.load(open(sys.argv[1], encoding="utf-8"))["items"]

    rows, pending, comparisons = [], None, 0
    for index, item in enumerate(items):
        text = clean(item["disasm"])
        match = CMP.match(text)
        if match:
            comparisons += 1
            pending = immediate(match.group(1))
            continue
        match = FIELD.match(text)
        if match and pending is not None:
            following = clean(items[index + 1]["disasm"]) if index + 1 < len(items) else ""
            rows.append({
                "ecl_address": base + pending,
                "field_offset": immediate(match.group(2)),
                "width": "byte" if match.group(1) == "al" else "word",
                "sign_extended": following == "cbw",
            })
            pending = None

    print("# ECL bank 1 計算值 → record 欄位對照")
    print()
    print("由 `scripts/ecl_bank1_field_map.py` 產生，不要手改。")
    print("base pointer 是 `DS:9594h` 的 far pointer；偏移是該 record 內的位移。")
    print("**本表只證明位址對映與存取寬度，不證明欄位語意。**")
    print()
    print("鏈上共 %d 個比較，抽出 %d 筆對映。" % (comparisons, len(rows)))
    print()
    print("| ECL 位址 | record 偏移 | 寬度 | 延伸 |")
    print("|---|---|---|---|")
    for row in rows:
        print("| `%04Xh` | `+%03Xh` | %s | %s |"
              % (row["ecl_address"], row["field_offset"], row["width"],
                 "符號" if row["sign_extended"] else "零"))
    if comparisons != len(rows):
        print()
        print("⚠ 比較數與抽出數不同（%d vs %d）：有分支不是單純的欄位讀取，"
              % (comparisons, len(rows)))
        print("必須逐一人工確認，不得假設漏掉的那幾筆不存在。")
    return 0


if __name__ == "__main__":
    sys.exit(main())
