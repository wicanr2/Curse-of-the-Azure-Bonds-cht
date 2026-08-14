"""列出每個「待解讀」函式引用到的 overlay 內嵌字串。

用途：字串是判讀一支函式最快的錨。`overlay-12` 這種「一個 entry 一個怪物特殊
攻擊」的模組，光看 `'Spits Acid'` 就知道該從哪裡下手（spec 819）。

作法：掃匯出裡的 `mov di, offset loc_XXXX` / `mov di, (offset loc_XXXX+N)` /
`mov di, offset unk_XXXX`，把那個 overlay-local 位址當成 Pascal 字串
（長度位元組 ＋ 內容）去 `.bin` 裡取。取不到合理內容的就跳過。

⚠ 這只是**線索**，不是判讀：位址被當成字串取出來不代表它真的是字串，也不代表
那支函式的語意就是字串講的那件事。一律要把本體讀完才能標「已解讀」。

用法：
    python3 scripts/function_strings.py <平台> <模組>
    python3 scripts/function_strings.py <平台>            # 掃該平台全部 overlay
"""

import glob
import json
import os
import re
import sys

ROOT = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
SWEEP = os.path.join(ROOT, "workplace", "re-sweep")
LEDGER = os.path.join(ROOT, "docs", "audit", "re-function-ledger.json")
RET = ("retf", "retn", "ret")
PATTERN = re.compile(r"offset\s+(?:loc|unk|sub|byte|word|asc|a)_([0-9A-F]+)"
                     r"(?:\+([0-9A-F]+))?")


def load(platform, module):
    for sub in ("filled", "prologue"):
        path = os.path.join(SWEEP, platform, "overlays", sub,
                            "%s-%s.json" % (platform, module))
        if os.path.exists(path):
            data = json.load(open(path, encoding="utf-8"))
            return data["functions"] if isinstance(data, dict) else data
    return None


def text(raw, offset):
    if offset >= len(raw):
        return None
    length = raw[offset]
    if not 2 <= length <= 60 or offset + 1 + length > len(raw):
        return None
    body = raw[offset + 1:offset + 1 + length]
    if not all(0x20 <= b < 0x7F or b >= 0x81 for b in body):
        return None
    try:
        return body.decode("cp932")
    except UnicodeDecodeError:
        return None


def run(platform, module, done):
    functions = load(platform, module)
    if not functions:
        return
    raw = open(os.path.join(SWEEP, platform, "overlays", "%s.bin" % module),
               "rb").read()
    for function in functions:
        items = function["items"]
        if not items:
            continue
        ea = items[0]["ea"]
        if (platform, module, ea) in done:
            continue
        found = []
        for item in items:
            for match in PATTERN.finditer(item["disasm"]):
                offset = int(match.group(1), 16) + int(match.group(2) or "0", 16)
                got = text(raw, offset)
                if got and got not in found:
                    found.append(got)
        if found:
            print("  %s:%05Xh（%d 條） %s"
                  % (module, ea, len(items),
                     " ／ ".join(repr(s) for s in found[:6])))


def main():
    if len(sys.argv) < 2:
        raise SystemExit(__doc__)
    platform = sys.argv[1]
    ledger = json.load(open(LEDGER, encoding="utf-8"))
    done = {(e["platform"], e["module"], e["ea"])
            for e in ledger["functions"] if e["state"] == "已解讀"}
    if len(sys.argv) >= 3:
        modules = [sys.argv[2]]
    else:
        modules = sorted(
            "overlay-" + os.path.basename(p).split("overlay-")[1].split(".")[0]
            for p in glob.glob(os.path.join(SWEEP, platform, "overlays",
                                            "filled", "%s-overlay-*.json" % platform)))
    for module in modules:
        print("== %s %s" % (platform, module))
        run(platform, module, done)
    return 0


if __name__ == "__main__":
    sys.exit(main())
