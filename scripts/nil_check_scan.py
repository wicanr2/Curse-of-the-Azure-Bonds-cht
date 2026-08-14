"""逐一檢查每個「取用目標陣列」的地方前面有沒有 NIL 判斷。

spec 716 第一版是**以函式為單位**掃的：只要函式裡出現過一次
`or ax, [di+7433h]` 就算「有檢查」。spec 725 的 `5AA8h` 推翻了這個做法——
同一支裡兩個迴圈，前一個直接取欄位、後一個才檢查，於是它被算成「有檢查」而
前半段的漏洞被蓋掉。

本工具改成**以取用點為單位**：找每一個 `les di, [di+7431h]`（把某一格當成 far
指標解開），往回看有限的幾條指令內有沒有出現 `[di+7433h]` 參與的 `or`。
沒有就標為未檢查。

限制寫在前面：往回看的窗格是固定的，用別種寫法（例如先搬到區域變數再判斷）
一樣會被算成未檢查。所以**輸出是「需要人工確認的清單」，不是缺陷清單**。

用法：
    python3 scripts/nil_check_scan.py [platform] [module]
"""

import json
import os
import re
import sys

ROOT = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
SWEEP = os.path.join(ROOT, "workplace", "re-sweep")
BASE_LOW, BASE_HIGH = "7431h", "7433h"
WINDOW = 14


def main():
    platform = sys.argv[1] if len(sys.argv) > 1 else "dos"
    module = sys.argv[2] if len(sys.argv) > 2 else "overlay-22"
    path = os.path.join(SWEEP, platform, "overlays", "prologue",
                        "%s-%s.json" % (platform, module))
    checked = unchecked = 0
    for function in json.load(open(path, encoding="utf-8"))["functions"]:
        items = function["items"]
        for index, item in enumerate(items):
            text = re.sub(r"\s+", " ", item["disasm"].strip())
            if not re.match(r"^les\s+\w+, \[di\+%s\]$" % BASE_LOW, text):
                continue
            guarded = False
            for back in range(index - 1, max(-1, index - WINDOW), -1):
                previous = re.sub(r"\s+", " ", items[back]["disasm"].strip())
                if BASE_HIGH in previous and previous.startswith("or "):
                    guarded = True
                    break
            if guarded:
                checked += 1
            else:
                unchecked += 1
                print("  %04Xh 內 %04Xh  取用目標陣列前 %d 條指令內沒有 NIL 判斷"
                      % (function["ea"], item["ea"], WINDOW))
    print("%s %s：取用點 %d 個，其中前方有 NIL 判斷的 %d 個、沒有的 %d 個"
          % (platform, module, checked + unchecked, checked, unchecked))
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
