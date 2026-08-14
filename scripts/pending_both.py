"""列出「兩側都待解讀」的配對——讀一次可以收兩筆。"""
import sys, os, json, difflib
sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
import module_align

ROOT = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
led = json.load(open(os.path.join(ROOT, "docs/audit/re-function-ledger.json"), encoding="utf-8"))
rows = {(r["platform"], r["module"], r["ea"]): r for r in led["functions"]}
def pend(p, m, e):
    r = rows.get((p, m, e))
    return r is None or r["state"] == "待解讀"

out = []
mods = sorted({m for (p, m, e) in rows if m.startswith("overlay-")})
for mod in mods:
    try:
        left, right, pairs = module_align.align(mod)
    except Exception:
        continue
    for i, j in pairs:
        if i is None or j is None:
            continue
        a = left[i]["items"][0]["ea"]
        b = right[j]["items"][0]["ea"]
        if not (pend("dos", mod, a) and pend("pc98", mod, b)):
            continue
        sa = [module_align.clean(x).split()[0] for x in left[i]["items"]]
        sb = [module_align.clean(x).split()[0] for x in right[j]["items"]]
        sim = difflib.SequenceMatcher(None, sa, sb).ratio()
        out.append((sim, mod, a, len(sa), b, len(sb)))
out.sort(key=lambda t: -t[0])
print("兩側都待解讀的配對：%d 對（收一次算兩筆）" % len(out))
for sim, mod, a, na, b, nb in out:
    print("  %.3f  %-12s dos %05X(%d 條)  pc98 %05X(%d 條)" % (sim, mod, a, na, b, nb))
