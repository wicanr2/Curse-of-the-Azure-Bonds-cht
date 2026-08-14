"""找出反組譯輸出裡「位元組數對不上位址」的地方。

`overlay-12:356Ch` 的匯出是：

    356C 55     push bp
    356D 89e5   mov bp, sp
    3570 ec     in al, dx        ← 位址從 356F 跳到 3570
    3571 5d     pop bp
    3572 cb     retf

原始位元組是 `55 89 E5 89 EC 5D CB`——`89 EC`（`mov sp, bp`）被切成兩半，
`89` 整個消失，剩下的 `EC` 被解成 `in al, dx`。實際上那是一支**空函式**。

危險的地方在於**假指令看起來很合理**：`in al, dx` 是連接埠讀取，照著寫下去會
變成「這支函式在讀 I/O 埠」，然後有人去追根本不存在的硬體存取。

判準：同一支函式裡，前一條的 `ea + 位元組數` 應該等於下一條的 `ea`。不相等就
是漏掉了位元組。輸出的每一筆都要人工回去對原始 `.bin`。

用法：
    python3 scripts/decode_gap_scan.py
"""

import glob
import json
import os

ROOT = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
SWEEP = os.path.join(ROOT, "workplace", "re-sweep")


def main():
    total = 0
    for platform in ("dos", "pc98"):
        for path in sorted(glob.glob(os.path.join(
                SWEEP, platform, "overlays", "prologue", "*.json"))):
            module = os.path.basename(path)[len(platform) + 1:-5]
            for function in json.load(open(path, encoding="utf-8"))["functions"]:
                items = function["items"]
                for index in range(len(items) - 1):
                    size = len(items[index]["bytes"]) // 2
                    expected = items[index]["ea"] + size
                    actual = items[index + 1]["ea"]
                    if expected != actual:
                        total += 1
                        print("  %-5s %-12s %04Xh 內：%04Xh(%d bytes) 之後應為 "
                              "%04Xh，實際 %04Xh，少了 %d bytes"
                              % (platform, module, function["ea"],
                                 items[index]["ea"], size, expected, actual,
                                 actual - expected))
    print("位元組數對不上的地方共 %d 處" % total)
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
