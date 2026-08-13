"""寫台帳條目，覆蓋既有「已解讀」時**擋下來**而不是靜默蓋掉。

會踩到的情境：一支函式在早幾輪就讀過並取好名字（例如 `FINDGUY`），後來因為
別的線索又讀了一次，新寫入把原本的名字與 spec 連結蓋掉——**台帳總數不會變**，
所以從進度數字上完全看不出來。

用法（讀 stdin 的 JSON 陣列）：
    python3 scripts/ledger_write.py < rows.json
    python3 scripts/ledger_write.py --force < rows.json   # 確認要覆蓋

每列必須有 platform／module／ea／state／spec／note，`level` 可省略。
"""

import json
import os
import sys

ROOT = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
LEDGER = os.path.join(ROOT, "docs", "audit", "re-function-ledger.json")
INDEX = os.path.join(ROOT, "docs", "audit", "coab-function-index.json")


def main():
    force = "--force" in sys.argv
    rows = json.load(sys.stdin)
    ledger = json.load(open(LEDGER, encoding="utf-8"))
    existing = {(e["platform"], e["module"], e["ea"]): e for e in ledger["functions"]}

    index = json.load(open(INDEX, encoding="utf-8"))
    known = {(f["platform"], f["module"], f["ea"])
             for f in (index.get("functions") or index)}
    unknown = [r for r in rows
               if (r["platform"], r["module"], r["ea"]) not in known]
    if unknown:
        # 位址打錯時條目會接不上索引：台帳裡多一列、統計卻不動，
        # 從進度數字上完全看不出來。所以這裡直接擋，不給 --force 繞過。
        print("以下位址不在函式索引裡（多半是 ea 打錯）：")
        for row in unknown:
            print("  %-5s %-12s %04Xh" % (row["platform"], row["module"], row["ea"]))
        return 2

    clash = []
    for row in rows:
        old = existing.get((row["platform"], row["module"], row["ea"]))
        if old is not None and old["state"] == "已解讀":
            clash.append((row, old))

    if clash and not force:
        print("以下條目已經是「已解讀」，不覆蓋（要蓋請加 --force）：")
        for row, old in clash:
            print("  %-5s %-12s %04Xh" % (row["platform"], row["module"], row["ea"]))
            print("    既有 spec：%s" % old.get("spec"))
            print("    既有註記：%s" % old["note"][:120])
        return 1

    for row in rows:
        key = (row["platform"], row["module"], row["ea"])
        row.setdefault("level", "exact")
        if key in existing:
            ledger["functions"] = [e for e in ledger["functions"]
                                   if (e["platform"], e["module"], e["ea"]) != key]
        ledger["functions"].append(row)
    json.dump(ledger, open(LEDGER, "w", encoding="utf-8"),
              ensure_ascii=False, indent=1)
    print("寫入 %d 筆" % len(rows))
    return 0


if __name__ == "__main__":
    sys.exit(main())
