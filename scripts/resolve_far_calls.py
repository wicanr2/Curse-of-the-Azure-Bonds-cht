"""把 overlay 內的 far call `seg:off` 解析成「resident 位址」或「overlay-N entry#i」。

**為什麼需要這一支**：raw overlay 丟進 IDA 後，`9A off seg` 的 segment 只是尚未
套用重定位的 addend。IDA 會把它當成同一個平坦映像的位址算出一個 `loc_XXXX`，
那個數字**落在 overlay 自己的 code 裡**，看起來像個合理的本地函式——但它是巧合。
（AGENTS.md：重定位未解析前，raw addend 只能標為「未解析候選」。）

正確換算：Borland 的段內 addend 是「相對於程式載入段的段號」，所以
    載入映像位移 L = seg×16 + off
    START.EXE 檔案位移 F = L + MZ header 大小（`e_cparhdr`×16）
若 `F` 落在某個 overlay 控制記錄的 entry stub 表內（每筆 5 bytes、`CD 3F` 開頭），
它就是「呼叫 overlay N 的第 i 個 entry」；否則是 resident 常駐碼。

用法：
    python3 scripts/resolve_far_calls.py <START.EXE> <ovr-manifest.json> \\
        [--calls <calls-overlay-NN.json>] [--target SEG:OFF ...] [--json out.json]

輸出每筆：`seg:off`、載入映像位移、檔案位移、判定（`overlay_entry` 或
`resident`）、overlay 索引與 entry 索引（0-based，與既有規格一致）、
entry 的 handler-local `code_offset`。
"""

import argparse
import json
import struct
import sys


def header_size(data):
    if data[:2] not in (b"MZ", b"ZM"):
        raise ValueError("不是 MZ 執行檔")
    return struct.unpack_from("<H", data, 0x08)[0] * 16


def build_stub_index(manifest):
    """entry stub 的檔案位移 → (overlay index, entry index, code_offset)。"""
    index = {}
    for ov in manifest["overlays"]:
        for entry in ov["entries"]:
            index[entry["executable_offset"]] = (
                ov["index"], entry["index"], entry["code_offset"])
    return index


def resolve(seg, off, data, hdr, stubs):
    load_offset = seg * 16 + off
    file_offset = load_offset + hdr
    record = {
        "target": "%04X:%04X" % (seg, off),
        "load_offset": load_offset,
        "file_offset": file_offset,
    }
    hit = stubs.get(file_offset)
    if hit is not None:
        overlay, entry, code_offset = hit
        record.update({
            "kind": "overlay_entry",
            "overlay": overlay,
            "entry": entry,
            "entry_code_offset": code_offset,
            "label": "overlay-%02d entry#%d" % (overlay, entry),
        })
    else:
        record.update({
            "kind": "resident",
            "label": "resident:%06X" % file_offset,
        })
    if 0 <= file_offset < len(data) - 5:
        record["bytes"] = data[file_offset:file_offset + 5].hex()
    return record


def parse_far_call(raw_hex):
    raw = bytes.fromhex(raw_hex)
    if len(raw) >= 5 and raw[0] == 0x9A:
        return raw[3] | (raw[4] << 8), raw[1] | (raw[2] << 8)
    return None


def main(argv):
    ap = argparse.ArgumentParser()
    ap.add_argument("executable")
    ap.add_argument("manifest")
    ap.add_argument("--calls", action="append", default=[])
    ap.add_argument("--target", action="append", default=[])
    ap.add_argument("--json")
    args = ap.parse_args(argv)

    data = open(args.executable, "rb").read()
    hdr = header_size(data)
    manifest = json.load(open(args.manifest, encoding="utf-8"))
    stubs = build_stub_index(manifest)

    out = []
    for text in args.target:
        seg_text, off_text = text.split(":")
        out.append(resolve(int(seg_text, 16), int(off_text, 16), data, hdr, stubs))
    for path in args.calls:
        doc = json.load(open(path, encoding="utf-8"))
        for call in doc.get("calls", []):
            parsed = parse_far_call(call.get("bytes", ""))
            if parsed is None:
                continue
            record = resolve(parsed[0], parsed[1], data, hdr, stubs)
            record["from_function"] = call.get("function")
            record["from_ea"] = call.get("ea")
            out.append(record)

    if args.json:
        with open(args.json, "w", encoding="utf-8") as fh:
            json.dump({"header_size": hdr, "resolved": out}, fh,
                      ensure_ascii=False, indent=1)
    for record in out:
        origin = ""
        if "from_ea" in record:
            origin = " @%04X" % record["from_ea"]
        print("%s%s -> %s%s" % (
            record["target"], origin, record["label"],
            "" if record["kind"] == "resident"
            else " (code_offset=%04X)" % record["entry_code_offset"]))
    return 0


if __name__ == "__main__":
    sys.exit(main(sys.argv[1:]))
