"""把 `overlay` 欄位補進 export_module.py 的輸出。

`analyze_overlay.py` 會寫這個欄位（模組名、entry 種子、code_size、SHA-256…），
`export_module.py` 不會。少了它，`cmd/re-ledger` 會退回用 `input.name`，模組名
變成 `overlay-25.bin`，整個台帳 join 不上——結果是「已解讀」數字無聲地掉一半。

只補 re-ledger 會用到的欄位，並如實標明來源是 manifest 而非這次分析：
`seeded_entries` 是**控制區塊宣告的 entry**，不是這次 IDA 實際種進去的，
所以 `origin` 標為 `manifest`。

用法：python3 scripts/attach_overlay_field.py <platform> <module> <in.json> <out.json>
"""
import json, os, sys

def main():
    platform, module, src, dst = sys.argv[1:5]
    root = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
    manifest = json.load(open(os.path.join(
        root, "workplace", "re-sweep", platform, "ovr-manifest.json"), encoding="utf-8"))
    record = next(o for o in manifest["overlays"] if o["module"] == module)
    payload = json.load(open(src, encoding="utf-8"))
    payload["overlay"] = {
        "platform": manifest.get("platform", platform),
        "module": module,
        "index": record["index"],
        "file_offset": record["file_offset"],
        "code_size": record["code_size"],
        "code_sha256": record["code_sha256"],
        "entry_count": record["entry_count"],
        "relocation_offsets": record["relocation_offsets"],
        # 控制區塊宣告的 entry；FFFFh 是未使用的 slot，不是位址。
        "seeded_entries": [
            {"index": e["index"], "stub_offset": e["stub_offset"],
             "code_offset": e["code_offset"], "flags": e["flags"], "origin": ""}
            for e in record["entries"] if e["code_offset"] != 0xFFFF],
        "field_source": "ovr-manifest.json（本檔由 export_module.py 匯出，"
                        "overlay 欄位為事後補上）",
    }
    json.dump(payload, open(dst, "w", encoding="utf-8"), ensure_ascii=False)
    return 0

if __name__ == "__main__":
    sys.exit(main())
