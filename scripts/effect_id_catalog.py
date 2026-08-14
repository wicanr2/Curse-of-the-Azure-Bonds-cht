"""把「效果編號」和同一支函式裡的訊息字串配起來。

`overlay-22` 的每一條法術 opcode 大致是「對某個效果編號做事，然後顯示一句話」。
編號都是立即數（`mov al, XXh` ＋ `push`），訊息都是 `mov di, offset` 指到的
Pascal 字串。兩者在同一支函式裡，配起來就得到一張編號↔語意的對照。

判準刻意保守：
- 只收 `overlay-23` entry#3／#16 與 `overlay-24` entry#27 這三支的呼叫點——
  它們的第二個參數位置放的是效果編號（spec 704／713／714 已逐一讀過）。
- 編號取「該 `call` 之前最近的一個 `mov al, imm8` ＋ `push ax`」。
- 訊息取整支函式裡所有 `mov di, offset` 指到、且解得出 Pascal 字串的位址。

**同一支函式有多個編號或多句訊息時不強配**，一律照實列出讓人判斷。所以本表是
線索不是結論；引用某個編號的語意前要回去讀那支函式。

用法：
    python3 scripts/effect_id_catalog.py [platform] [module]
"""

import collections
import json
import os
import re
import sys

ROOT = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
SWEEP = os.path.join(ROOT, "workplace", "re-sweep")
OUT = os.path.join(ROOT, "docs", "audit", "effect-id-catalog.md")
sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
from overlay_call_graph import modules, resolve   # noqa: E402

WATCHED = {("overlay-23", 3), ("overlay-23", 16), ("overlay-24", 27)}
MOV_AL = re.compile(r"^mov\s+al,\s*([0-9A-Fa-f]+)h?$")


def clean(text):
    """去掉反組譯的行末註解再比對。

    IDA 會替可列印的立即數加上 `; '7'` 這種註解。忘了去掉的話，正則會**只漏掉
    值落在可列印 ASCII 範圍的編號**——`0Fh`、`16h` 照樣抓到，`24h`、`37h`、`52h`
    全部消失。漏得有規律，所以結果看起來完全正常。
    """
    return re.sub(r"\s*;.*$", "", re.sub(r"\s+", " ", text.strip()))
MOV_DI = re.compile(r"^mov\s+di,\s*(?:offset\s+\w+|([0-9A-Fa-f]+h))")


def pascal(blob, offset):
    if offset >= len(blob):
        return None
    length = blob[offset]
    if length < 3 or offset + 1 + length > len(blob):
        return None
    body = blob[offset + 1:offset + 1 + length]
    if not all(0x20 <= b <= 0x7E for b in body):
        return None
    return body.decode("cp437")


def main():
    platform = sys.argv[1] if len(sys.argv) > 1 else "dos"
    module = sys.argv[2] if len(sys.argv) > 2 else "overlay-22"
    table = modules(platform)
    blob = open(os.path.join(SWEEP, platform, "overlays",
                             module + ".bin"), "rb").read()
    path = os.path.join(SWEEP, platform, "overlays", "prologue",
                        "%s-%s.json" % (platform, module))
    rows = []
    for function in json.load(open(path, encoding="utf-8"))["functions"]:
        items = function["items"]
        texts, ids = [], []
        for index, item in enumerate(items):
            text = clean(item["disasm"])
            raw = item["bytes"]
            if raw.startswith("bf") and len(raw) == 6:
                offset = int(raw[2:4], 16) | int(raw[4:6], 16) << 8
                found = pascal(blob, offset)
                if found and found not in texts:
                    texts.append(found)
            if raw.startswith("9a") and len(raw) == 10:
                off = int(raw[2:4], 16) | int(raw[4:6], 16) << 8
                seg = int(raw[6:8], 16) | int(raw[8:10], 16) << 8
                hit = resolve(table, seg, off)
                if hit and (hit["module"], hit["entry"]) in WATCHED:
                    for back in range(index - 1, max(-1, index - 12), -1):
                        match = MOV_AL.match(clean(items[back]["disasm"]))
                        if match:
                            value = int(match.group(1), 16)
                            label = "%s#%d" % (hit["module"][-2:], hit["entry"])
                            if (value, label) not in ids:
                                ids.append((value, label))
                            break
        if ids:
            rows.append((function["ea"], ids, texts))

    lines = ["# 效果編號 ↔ 訊息 對照（線索表）", "",
             "由 `scripts/effect_id_catalog.py` 產生，範圍 `%s %s`。" % (platform, module),
             "每一列是一支函式：它對哪些效果編號做了事，以及它裡面有哪些訊息字串。", "",
             "**同一支有多個編號或多句訊息時本表不強配**——這是線索不是結論，",
             "引用某個編號的語意之前要回去讀那支函式。", "",
             "`23#3` ＝ `overlay-23` entry#3（解除）、`23#16` ＝ entry#16（查詢）、",
             "`24#27` ＝ `overlay-24` entry#27（找節點）。", "",
             "| 函式 | 編號 | 訊息 |", "|---|---|---|"]
    for ea, ids, texts in rows:
        lines.append("| `%04Xh` | %s | %s |"
                     % (ea,
                        " ".join("`%02Xh`(%s)" % (v, k) for v, k in ids),
                        " / ".join(t.replace("|", "\\|") for t in texts) or "—"))
    counter = collections.Counter(v for _, ids, _ in rows for v, _ in ids)
    lines += ["", "## 各編號出現次數", "",
              " ".join("`%02Xh`×%d" % (v, n) for v, n in sorted(counter.items()))]
    open(OUT, "w", encoding="utf-8").write("\n".join(lines) + "\n")
    print("%d 支函式用到效果編號，%d 個相異編號 → %s"
          % (len(rows), len(counter), OUT))
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
