"""把 IDA FLIRT 認出的 Turbo Pascal RTL 函式批次寫進覆蓋台帳。

判定依據是**具體證據**：IDA 依 Borland 簽章還原出 mangled 名稱
（`@CLRSCR$qv`、`@MSDOS$qm9REGISTERS`、`__RealMul`…），這些屬於 Turbo Pascal
的 `System`／`Dos`／`Crt`／`Overlay` 單元，不是遊戲程式碼。理由可被推翻：
只要指出某個名稱其實不是 RTL，或該 RTL 函式影響玩家可見結果即可。

**例外清單留 `待解讀`**：RTL 之中仍有直接決定玩家可見結果的：

- `Random`／`Randomize`：原版亂數是戰鬥、遭遇與寶物的來源，專案明確要求
  不得用 remake 的 PRNG 冒稱原版。
- `Sound`／`NoSound`／`Delay`：PC 喇叭發聲與時序，玩家聽得到也看得到。

不在例外清單、但日後發現會影響玩家結果的，改台帳即可，不要改這支腳本的
判定方式去遷就個案。

用法：
    python3 scripts/rtl_ledger_batch.py            # 預覽
    python3 scripts/rtl_ledger_batch.py --write    # 寫入台帳
"""

import glob
import json
import os
import re
import sys

ROOT = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
SWEEP = os.path.join(ROOT, "workplace", "re-sweep")
LEDGER = os.path.join(ROOT, "docs", "audit", "re-function-ledger.json")
SPEC = "docs/spec/566-turbo-pascal-rtl-not-blocking.md"

# 例外：名稱（去掉 Borland 修飾後）落在這裡就維持待解讀。
PLAYER_VISIBLE = {"random", "randomize", "sound", "nosound", "delay"}

MANGLED = re.compile(r"^@(?:__)?(?P<name>[A-Za-z_][A-Za-z0-9_]*)(?:\$|__)")
COMPILER = re.compile(r"^__(?P<name>[A-Za-z][A-Za-z0-9_]*)$")
# Borland 的運算子輔助：`@$basg$`（指派）、`@$brmul$`（乘）、`@Set@…`（集合運算）。
# `$b` 前綴是 Borland 對運算子的編碼，屬編譯器產生碼，不是遊戲程式。
OPERATOR = re.compile(r"^@(?:Set@)?\$b(?P<name>[a-z]+)\$")
SET_HELPER = re.compile(r"^@Set@(?P<name>[A-Za-z]+)\$")


def rtl_name(symbol):
    """回傳去修飾後的 RTL 名稱；不是 RTL 就回 None。"""
    match = MANGLED.match(symbol)
    if match:
        return match.group("name")
    match = COMPILER.match(symbol)
    if match:
        return match.group("name")
    for pattern in (OPERATOR, SET_HELPER):
        match = pattern.match(symbol)
        if match:
            return "operator:" + match.group("name")
    return None


def main():
    write = "--write" in sys.argv
    entries, skipped = [], []
    for platform in ("dos", "pc98"):
        for path in sorted(glob.glob(os.path.join(SWEEP, platform, "out", "*.json"))):
            if path.endswith(".error.log"):
                continue
            data = json.load(open(path, encoding="utf-8"))
            module = data["overlay"]["module"] if data.get("overlay") else data["input"]["name"]
            for function in data["functions"]:
                if function["auto_named"]:
                    continue
                name = rtl_name(function["name"])
                if name is None:
                    continue
                if name.lower() in PLAYER_VISIBLE:
                    skipped.append((platform, module, function["ea"], function["name"]))
                    continue
                entries.append({
                    "platform": platform, "module": module, "ea": function["ea"],
                    "state": "不阻塞", "level": "", "spec": SPEC,
                    "note": ("Turbo Pascal 編譯器運算子輔助：IDA 還原名稱 `%s`，"
                             "`$b` 前綴是 Borland 的運算子編碼"
                             if name.startswith("operator:") else
                             "Turbo Pascal RTL：IDA 依 Borland 簽章還原名稱 `%s`；"
                             "屬 System／Dos／Crt／Overlay 單元，不實作遊戲規則")
                            % function["name"],
                })

    print("可標不阻塞：%d 筆" % len(entries))
    print("保留待解讀的玩家可見 RTL：%d 筆" % len(skipped))
    for platform, module, ea, symbol in skipped:
        print("  %-5s %-12s %04Xh %s" % (platform, module, ea, symbol))

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
