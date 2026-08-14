"""印出一支函式，並把每個 far call 換成「模組 entry#N ＋ 對面平台已讀完的判讀」。

沒有這一層的話，讀 overlay 反組譯要同時對付兩個問題：far call 那行的名字是
IDA 攤平 `seg×16+off` 之後套上的**假名字**（spec 687），而真名要人工翻
manifest。結果就是一支函式讀完只知道「它呼叫了六個地方」。

本工具把兩件事接起來：
1. 用 `overlay_call_graph` 把 `9A` 的 `seg:off` 解回（模組, entry 編號, 目標位址）。
2. 拿該 entry 在**另一個平台**的位址去查台帳——兩平台的 entry 編號一致
   （spec 562），所以對面讀完的那筆就是這裡的語意來源。同平台已讀完的話
   一併顯示。

因此輸出裡的註解**不是本函式的證據**，是被呼叫者的。引用時要標明來源那一筆。

用法：
    python3 scripts/annotated_dump.py <platform> <module> <ea16> [end16]
"""

import json
import os
import sys

ROOT = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
SWEEP = os.path.join(ROOT, "workplace", "re-sweep")
sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
from overlay_call_graph import modules, resolve   # noqa: E402

LEDGER = os.path.join(ROOT, "docs", "audit", "re-function-ledger.json")
ARITY = os.path.join(ROOT, "docs", "audit", "callee-arity.json")


def arity():
    """被呼叫者的參數個數，來源是它結尾的 `retf N`（spec 711）。

    印在呼叫點旁邊是為了擋掉「數呼叫端連續幾個 push」這個錯法：推入數量會因為
    堆疊殘留（spec 690／710）而多於或少於實際參數，而錯誤會沿著推導鏈傳染。
    """
    if not os.path.exists(ARITY):
        return {}
    return json.load(open(ARITY, encoding="utf-8"))["functions"]


def ledger():
    out = {}
    for entry in json.load(open(LEDGER, encoding="utf-8"))["functions"]:
        if entry["state"] != "待解讀":
            out[(entry["platform"], entry["module"], entry["ea"])] = entry
    return out


def entry_offsets(platform, module):
    manifest = json.load(open(os.path.join(SWEEP, platform, "ovr-manifest.json"),
                              encoding="utf-8"))
    overlay = next((o for o in manifest["overlays"] if o["module"] == module), None)
    if overlay is None:
        return {}
    return {e["index"]: e["code_offset"] for e in overlay["entries"]}


def main():
    platform, module = sys.argv[1], sys.argv[2]
    start = int(sys.argv[3], 16)
    end = int(sys.argv[4], 16) if len(sys.argv) > 4 else None
    other = "pc98" if platform == "dos" else "dos"

    # prologue 匯出只收「push bp / mov bp, sp」開場的函式；自己管堆疊的
    # （例如 PC-98 直接叫磁碟 BIOS 的那幾支）只在 full 匯出裡，見 spec 1080。
    chosen = None
    for flavour in ("prologue", "full"):
        path = os.path.join(SWEEP, platform, "overlays", flavour,
                            "%s-%s.json" % (platform, module))
        if not os.path.exists(path):
            continue
        functions = json.load(open(path, encoding="utf-8"))["functions"]
        chosen = next((f for f in functions
                       if f["ea"] == start and f.get("items")), None)
        if chosen is not None:
            break
    if chosen is None:
        print("找不到起點 %04Xh（prologue 與 full 匯出裡都沒有這個位址）" % start)
        return 2
    limit = end if end is not None else 1 << 30

    table = modules(platform)
    known = ledger()
    arities = arity()
    cache = {}
    for item in chosen["items"]:
        if item["ea"] >= limit:
            break
        line = "  %04X  %s" % (item["ea"], item["disasm"])
        raw = item["bytes"]
        if raw.startswith("9a") and len(raw) == 10:
            offset = int(raw[2:4], 16) | int(raw[4:6], 16) << 8
            segment = int(raw[6:8], 16) | int(raw[8:10], 16) << 8
            hit = resolve(table, segment, offset)
            if hit is None:
                line += "        ; → resident %04X:%04X" % (segment, offset)
            else:
                info = arities.get("%s/%s/%04X" % (platform, hit["module"], hit["ea"]))
                shape = ""
                if info and info["words"] is not None:
                    shape = "，收 %d word" % info["words"]
                elif info:
                    shape = "，參數個數未知（%s）" % info["kind"]
                line += "        ; → %s entry#%d @%04Xh%s" % (
                    hit["module"], hit["entry"], hit["ea"], shape)
                key = (hit["module"], hit["entry"])
                if key not in cache:
                    notes = []
                    for plat in (platform, other):
                        ea = entry_offsets(plat, hit["module"]).get(hit["entry"])
                        found = known.get((plat, hit["module"], ea))
                        if found:
                            notes.append("      [%s] %s" % (plat, found["note"]))
                    cache[key] = notes
                for note in cache[key]:
                    line += "\n" + note
        print(line)
    return 0


if __name__ == "__main__":
    sys.exit(main())
