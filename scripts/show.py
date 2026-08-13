"""印出指定函式的完整逐指令 body（從 full 匯出讀，不重開 IDA）。

用法：python3 scripts/show.py <platform> <module> <ea16> [ea16...]
"""
import json, os, sys
ROOT = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
SWEEP = os.path.join(ROOT, "workplace", "re-sweep")

def load(platform, module):
    for base in (os.path.join(SWEEP, platform, "overlays", "full"),
                 os.path.join(SWEEP, platform, "full")):
        p = os.path.join(base, "%s-%s.json" % (platform, module))
        if os.path.exists(p):
            return {f["ea"]: f for f in json.load(open(p, encoding="utf-8"))["functions"]}
    raise SystemExit("找不到 full 匯出：%s %s" % (platform, module))

def main():
    platform, module = sys.argv[1], sys.argv[2]
    fns = load(platform, module)
    for arg in sys.argv[3:]:
        ea = int(arg, 16)
        f = fns.get(ea)
        if f is None:
            print("!! %s %s %04Xh 不在匯出裡" % (platform, module, ea)); continue
        print("=== %s %s %04Xh %s  size=%d args=%d callers=%d"
              % (platform, module, ea, f["name"], f["size"], f["arg_bytes"], len(f["callers"])))
        for it in f["items"]:
            print("  %04X  %-16s %s" % (it["ea"], it["bytes"], it["disasm"]))
        print()

if __name__ == "__main__":
    main()
