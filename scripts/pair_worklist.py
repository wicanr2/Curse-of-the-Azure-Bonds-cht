"""列出「已跨平台配對但兩邊都還沒讀」的 PC-98 函式，依價值排序。

讀這些函式的邊際效益是雙倍：讀完 PC-98 這一支，
`cross_platform_pair.py --write` 會把判讀轉移到 DOS 的對應支。

排序依據：被呼叫次數（callers）多的先讀——它是該單元的共用邏輯，
讀懂它同時降低其餘函式的閱讀成本。

用法：
    python3 scripts/pair_worklist.py <overlay> [數量]
    python3 scripts/pair_worklist.py --summary
"""
import json, os, sys, collections
sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
from cross_platform_pair import full, unique_index, LEDGER, ROOT

def symbols():
    """符號名直接取自全函式索引的 `borland_symbol` 欄位。

    不要自己再解一次 `borland-symbols.json`：那份用的是 segment 而非模組名，
    segment→overlay 的歸屬是 `cmd/borland-symbols` 做的，索引已經是歸屬後的
    結果。重解一次只會多一個對不齊的來源。
    """
    path = os.path.join(ROOT, "docs", "audit", "coab-function-index.json")
    data = json.load(open(path, encoding="utf-8"))
    return {(f["platform"], f["module"], f["ea"]): f.get("borland_symbol") or f["ida_name"]
            for f in data["functions"]}


def worklist(module):
    ledger = json.load(open(LEDGER, encoding="utf-8"))
    state = {(e["platform"], e["module"], e["ea"]): e["state"] for e in ledger["functions"]}
    d, p = unique_index(full("dos", module)), unique_index(full("pc98", module))
    sym = symbols()
    rows = []
    for seq in set(d) & set(p):
        pf, df = p[seq], d[seq]
        if state.get(("pc98", module, pf["ea"]), "待解讀") != "待解讀":
            continue
        if state.get(("dos", module, df["ea"]), "待解讀") != "待解讀":
            continue
        rows.append((len(pf["callers"]), pf["size"], pf["ea"], df["ea"],
                     sym.get(("pc98", module, pf["ea"]), pf["name"]), len(seq)))
    rows.sort(key=lambda r: (-r[0], r[1]))
    return rows

def main():
    if "--summary" in sys.argv:
        total = 0
        for i in list(range(36)):
            m = "overlay-%02d" % i
            rows = worklist(m)
            if rows:
                total += len(rows)
                print("  %-12s %4d 對待讀  最大 %d bytes" % (m, len(rows), max(r[1] for r in rows)))
        print("合計 %d 對（= %d 個函式）" % (total, total * 2))
        return 0
    module = sys.argv[1]
    limit = int(sys.argv[2]) if len(sys.argv) > 2 else 20
    rows = worklist(module)
    print("%s：%d 對待讀" % (module, len(rows)))
    print("callers  size   pc98    dos    指令數  名稱")
    for c, s, pe, de, name, n in rows[:limit]:
        print("  %4d  %5d  %04Xh  %04Xh  %5d  %s" % (c, s, pe, de, n, name))
    return 0

if __name__ == "__main__":
    sys.exit(main())
