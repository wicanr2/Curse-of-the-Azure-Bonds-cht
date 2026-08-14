"""把常駐執行檔的資料段原封不動倒出來，並附上段落清單。

為什麼需要：遊戲有大量表格住在**常駐資料段**（`DS:xxxx`）而不是 overlay 的
`.bin` 裡——法術名稱表（`DS:27BDh`）、金錢欄位名稱（`DS:0F93h`）、方向 dx／dy
表（`DS:2694h` / `DS:269Dh`）、物品與法術屬性表（`DS:5CF6h` / `DS:37DAh`）等。
`scripts/scan_pascal_strings.py` 只掃 overlay，這些一律掃不到，所以先前的判讀
只能寫「位址在這裡、內容沒讀」。

輸出兩份：

- `<out>.json`：每個段的 `name` / `class` / `start` / `end` / `sel`，以及
  「以 `sel` 為基準的段內位移」——**`DS:xxxx` 要對到哪一段，靠的是這個對照，
  不是猜**。
- `<out>-<段名>.bin`：該段的原始位元組，長度就是 `end − start`。
  未初始化的位元組讀出來是 `0`，這一點在 JSON 裡以 `has_bytes` 標明。

用法：
    tools/ida.sh py START.EXE.i64 dump_data_segments.py /work/out/dos-dseg
"""

import json
import os
import sys

import ida_auto
import ida_bytes
import ida_pro
import ida_segment


def main():
    prefix = sys.argv[1]
    ida_auto.auto_wait()

    segments, written = [], []
    segment = ida_segment.get_first_seg()
    while segment is not None:
        name = ida_segment.get_segm_name(segment) or "seg_%X" % segment.start_ea
        klass = ida_segment.get_segm_class(segment) or ""
        size = segment.end_ea - segment.start_ea
        row = {"name": name, "class": klass, "start": segment.start_ea,
               "end": segment.end_ea, "size": size, "sel": segment.sel,
               "has_bytes": False}
        if klass in ("DATA", "BSS", "CONST") and size > 0:
            blob = ida_bytes.get_bytes(segment.start_ea, size)
            if blob:
                path = "%s-%s.bin" % (prefix, name)
                os.makedirs(os.path.dirname(path) or ".", exist_ok=True)
                open(path, "wb").write(blob)
                written.append(os.path.basename(path))
                row["has_bytes"] = True
                row["file"] = os.path.basename(path)
        segments.append(row)
        segment = ida_segment.get_next_seg(segment.start_ea)

    os.makedirs(os.path.dirname(prefix) or ".", exist_ok=True)
    json.dump({"schema": "coab-data-segments/1", "segments": segments,
               "files": written},
              open(prefix + ".json", "w", encoding="utf-8"),
              ensure_ascii=False, indent=1)
    return 0


if __name__ == "__main__":
    try:
        rc = main()
    except BaseException:
        import traceback
        with open(sys.argv[1] + ".error.log", "w", encoding="utf-8") as handle:
            traceback.print_exc(file=handle)
        rc = 3
    ida_pro.qexit(rc)
