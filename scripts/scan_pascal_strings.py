"""掃出兩平台每個模組內嵌的 Turbo Pascal 短字串。

Turbo Pascal 的 string 常數是「長度位元組 ＋ 內容」，編譯後直接躺在 code 段
裡，用 `mov di, offset X` ＋ `push cs` 傳位址。所以掃描要同時滿足兩件事：

1. **形狀對**：位址 `i` 的位元組是長度 `n`，其後 `n` 個位元組全是合法字元
   （DOS 為可列印 ASCII；PC-98 另外接受 Shift-JIS 雙位元組序列）。
2. **有人引用**：`BF lo hi`（`mov di, imm16`）的立即值等於 `i`。

只有形狀對、沒人引用的，多半是把普通資料誤讀成字串——一律排除。這條件會
漏掉用其他方式取址的字串，所以**本工具的輸出是下界，不是全集**；不得由某
字串不在輸出裡推論它不存在。

用法：
    python3 scripts/scan_pascal_strings.py            # 掃描並寫出報表
    python3 scripts/scan_pascal_strings.py <module>   # 只看一個模組
"""

import json
import os
import re
import sys

ROOT = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
SWEEP = os.path.join(ROOT, "workplace", "re-sweep")
OUT_JSON = os.path.join(ROOT, "docs", "audit", "embedded-strings.json")
OUT_MD = os.path.join(ROOT, "docs", "audit", "embedded-strings.md")
MIN_LEN = 4
MIN_CHAIN = 6      # 連續幾條都合法才採信整片


def sjis_ok(data):
    i = 0
    while i < len(data):
        b = data[i]
        if 0x20 <= b <= 0x7E:
            i += 1
        elif (0x81 <= b <= 0x9F or 0xE0 <= b <= 0xFC) and i + 1 < len(data):
            trail = data[i + 1]
            if not (0x40 <= trail <= 0x7E or 0x80 <= trail <= 0xFC):
                return False
            i += 2
        elif 0xA1 <= b <= 0xDF:      # 半形片假名
            i += 1
        else:
            return False
    return True


def ascii_ok(data):
    return all(0x20 <= b <= 0x7E for b in data)


def references(blob):
    """所有 `mov di, imm16`（BF lo hi）的立即值。"""
    out = set()
    for match in re.finditer(rb"\xbf(..)", blob, re.S):
        out.add(match.group(1)[0] | (match.group(1)[1] << 8))
    return out


def chains(blob, check, decode, covered):
    """連續鏈判準：一條 Pascal 字串結束後，下一個位元組正好是下一條的長度。

    大片字串常數表（`overlay-11` 的 `Loading...` 那 1705 bytes）裡，多數字串
    沒有任何 `mov di, offset` 直接指著——取址方式是「表基底 ＋ 位移」或由呼叫
    端算出。單條的形狀判準對它們無效。

    但整片是**首尾相接**的：`[len][內容][len][內容]…`。要求連續 `MIN_CHAIN`
    條都合法，誤判機率就低到可用——隨機資料很難連續多次剛好對上長度。

    只從「還沒被引用判準涵蓋」的位址起頭，避免重複。
    """
    out, offset = [], 0
    limit = len(blob)
    while offset < limit:
        run, cursor = [], offset
        while cursor < limit:
            length = blob[cursor]
            if length < MIN_LEN or cursor + 1 + length > limit:
                break
            payload = blob[cursor + 1:cursor + 1 + length]
            if not check(payload):
                break
            text = decode(payload)
            if text is None:
                break
            run.append({"offset": cursor, "length": length, "text": text})
            cursor += 1 + length
        if len(run) >= MIN_CHAIN:
            out.extend(item for item in run if item["offset"] not in covered)
            offset = cursor
        else:
            offset += 1
    return out


def scan(path, platform):
    blob = open(path, "rb").read()
    refs = references(blob)
    check = sjis_ok if platform == "pc98" else ascii_ok
    def decode(payload):
        try:
            return payload.decode("cp932" if platform == "pc98" else "cp437")
        except UnicodeDecodeError:
            return None

    found = []
    for offset in sorted(refs):
        if offset >= len(blob):
            continue
        length = blob[offset]
        if length < MIN_LEN or offset + 1 + length > len(blob):
            continue
        payload = blob[offset + 1:offset + 1 + length]
        if not check(payload):
            continue
        try:
            text = payload.decode("cp932" if platform == "pc98" else "cp437")
        except UnicodeDecodeError:
            continue
        found.append({"offset": offset, "length": length, "text": text, "by": "ref"})
    covered = {item["offset"] for item in found}
    for item in chains(blob, check, decode, covered):
        item["by"] = "chain"
        found.append(item)
    found.sort(key=lambda item: item["offset"])
    return found


def modules():
    for platform in ("dos", "pc98"):
        directory = os.path.join(SWEEP, platform, "overlays")
        for name in sorted(os.listdir(directory)):
            if name.endswith(".bin"):
                yield platform, name[:-4], os.path.join(directory, name)


def main():
    only = sys.argv[1] if len(sys.argv) > 1 else None
    result, total = {}, 0
    for platform, module, path in modules():
        if only and module != only:
            continue
        found = scan(path, platform)
        if found:
            result.setdefault(platform, {})[module] = found
            total += len(found)

    if only:
        for platform in ("dos", "pc98"):
            for item in result.get(platform, {}).get(only, []):
                print("%-5s %04Xh  %s" % (platform, item["offset"], item["text"]))
        return 0

    os.makedirs(os.path.dirname(OUT_JSON), exist_ok=True)
    json.dump({"schema": "coab-embedded-strings/1", "min_length": MIN_LEN,
               "modules": result}, open(OUT_JSON, "w", encoding="utf-8"),
              ensure_ascii=False, indent=1)

    lines = ["# overlay 內嵌字串（掃描結果）", "",
             "由 `scripts/scan_pascal_strings.py` 產生。採信一條字串的依據有兩種：",
             "「有 `mov di, offset` 引用」（標 `ref`），或「連續 %d 條首尾相接的"
             "字串表」（標 `chain`）。" % MIN_CHAIN,
             "兩者都是保守判準，**這是下界不是全集**：用其他方式取址、又不在連續表",
             "裡的字串不會出現在這裡，不得由缺席推論不存在。", ""]
    for platform in ("dos", "pc98"):
        modules_found = result.get(platform, {})
        lines.append("## %s（%d 個模組，%d 條）"
                     % (platform, len(modules_found),
                        sum(len(v) for v in modules_found.values())))
        lines.append("")
        for module in sorted(modules_found):
            lines.append("### %s" % module)
            lines.append("")
            lines.append("| offset | 依據 | 內容 |")
            lines.append("|---|---|---|")
            for item in modules_found[module]:
                text = item["text"].replace("|", "\\|")
                lines.append("| `%04Xh` | %s | %s |" % (item["offset"], item["by"], text))
            lines.append("")
    open(OUT_MD, "w", encoding="utf-8").write("\n".join(lines))
    print("字串 %d 條 → %s" % (total, OUT_MD))
    for platform in ("dos", "pc98"):
        per = result.get(platform, {})
        print("  %-5s %d 模組 %d 條" % (platform, len(per), sum(len(v) for v in per.values())))
    return 0


if __name__ == "__main__":
    sys.exit(main())
