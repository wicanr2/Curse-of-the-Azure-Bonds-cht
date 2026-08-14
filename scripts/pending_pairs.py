"""列出待解讀函式中「對側已解讀且相似度高」的，這種只要核差異區塊就能收。"""
import sys, os, json
sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
import module_align

ROOT = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
led = json.load(open(os.path.join(ROOT, "docs/audit/re-function-ledger.json"), encoding="utf-8"))
rows = {(r["platform"], r["module"], r["ea"]): r for r in led["functions"]}

OTHER = {"dos": "pc98", "pc98": "dos"}
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
        la = module_align.clean_seq(left[i]) if hasattr(module_align, "clean_seq") else None
        sa = rows.get(("dos", mod, a))
        sb = rows.get(("pc98", mod, b))
        import difflib
        seqa = [module_align.clean(x).split()[0] for x in left[i]["items"]]
        seqb = [module_align.clean(x).split()[0] for x in right[j]["items"]]
        sim = difflib.SequenceMatcher(None, seqa, seqb).ratio()
        if sim < 0.72:
            continue
        pend_a = (sa is None) or sa["state"] == "待解讀"
        pend_b = (sb is None) or sb["state"] == "待解讀"
        done_a = sa is not None and sa["state"] == "已解讀"
        done_b = sb is not None and sb["state"] == "已解讀"
        if pend_a and done_b:
            out.append((sim, "dos", mod, a, len(left[i]["items"]), sb.get("spec")))
        elif pend_b and done_a:
            out.append((sim, "pc98", mod, b, len(right[j]["items"]), sa.get("spec")))
out.sort(key=lambda t: -t[0])
print("對側已解讀、相似度 >= 0.90 的待解讀函式：%d 筆" % len(out))
for sim, plat, mod, ea, n, spec in out:
    print("  %.3f  %-4s %-12s %05X  %3d 條  對側 spec=%s" % (sim, plat, mod, ea, n, spec))
