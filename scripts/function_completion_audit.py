"""盤點函式 RE 完成度，並把 caller=0 依可追溯理由互斥分類。

輸入是 cmd/re-ledger 產生的 coab-function-index.json；本工具不改台帳。
"""

import argparse
import collections
import json


def is_rtl(function):
    note = function.get("note") or ""
    name = function.get("ida_name") or ""
    return "Turbo Pascal RTL" in note or "Turbo Pascal 編譯器運算子輔助" in note or name.startswith("@")


def is_empty(function):
    note = function.get("note") or ""
    return "空函式" in note or "空程序" in note


def caller_zero_class(function):
    if function.get("is_overlay_entry"):
        return "overlay_entry"
    if is_rtl(function):
        return "borland_rtl"
    if is_empty(function):
        return "empty_function"
    if function.get("state") == "邊界碎片":
        return "ida_boundary_fragment"
    return "unexplained"


def render(functions):
    states = collections.Counter(f.get("state") or "待解讀" for f in functions)
    by_platform = collections.defaultdict(collections.Counter)
    for function in functions:
        by_platform[function["platform"]][function.get("state") or "待解讀"] += 1

    caller_zero = [f for f in functions if f.get("callers") == 0]
    caller_classes = collections.Counter(caller_zero_class(f) for f in caller_zero)
    pending = [f for f in functions if (f.get("state") or "待解讀") == "待解讀"]
    unexplained = [f for f in caller_zero if caller_zero_class(f) == "unexplained"]
    rtl = [f for f in functions if is_rtl(f)]

    fragments = collections.Counter()
    for function in functions:
        if function.get("state") == "邊界碎片":
            fragments[(function["platform"], function["module"])] += 1

    out = [
        "# CoAB 函式 RE 完成與無直接 caller 分類",
        "",
        "由 `scripts/function_completion_audit.py` 從 `docs/audit/coab-function-index.json` 產生；不要手改。",
        "",
        "## 結論",
        "",
        f"- 函式台帳：**{len(functions)}** 筆。",
        f"- 真正尚待 RE 的函式（`待解讀`）：**{len(pending)}**。",
        f"- `callers = 0`：**{len(caller_zero)}**；排除 entry、RTL、空函式與錯切邊界後，未解釋 **{len(unexplained)}**。",
        f"- Borland／Turbo Pascal RTL 或編譯器輔助：**{len(rtl)}**；library 身分不自動等於玩家可見語意可忽略。",
        f"- IDA `邊界碎片`：**{states['邊界碎片']}**；它們不是可獨立宣稱『尚未 RE』的函式。",
        "",
        "## 狀態",
        "",
        "| 平台 | 已解讀 | 不阻塞 | 邊界碎片 | 待解讀 |",
        "|---|---:|---:|---:|---:|",
    ]
    for platform in ("dos", "pc98"):
        row = by_platform[platform]
        out.append(f"| {platform} | {row['已解讀']} | {row['不阻塞']} | {row['邊界碎片']} | {row['待解讀']} |")

    labels = {
        "overlay_entry": "overlay entry（entry table／manager 間接分派）",
        "borland_rtl": "Borland RTL／編譯器輔助",
        "empty_function": "空函式／空程序",
        "ida_boundary_fragment": "IDA 錯切邊界碎片",
        "unexplained": "仍無解釋",
    }
    out.extend(["", "## `callers = 0` 互斥分類", "", "| 類別 | 數量 |", "|---|---:|"])
    for key in ("overlay_entry", "borland_rtl", "empty_function", "ida_boundary_fragment", "unexplained"):
        out.append(f"| {labels[key]} | {caller_classes[key]} |")

    out.extend([
        "",
        "直接 code xref 是下界；overlay entry、callback、relocation 與 far pointer 必須先排除。",
        "Borland pattern 的證據與限制見 Codex 知識庫 `local/borland-turbo-pascal-runtime-patterns.md`。",
        "",
        "## 邊界碎片分布",
        "",
        "這一節是 IDA 資料品質台帳，不是待實作玩法清單。",
        "",
        "| 平台／模組 | 筆數 |",
        "|---|---:|",
    ])
    for (platform, module), count in sorted(fragments.items()):
        out.append(f"| {platform}/{module} | {count} |")

    if pending:
        out.extend(["", "## 真正尚待 RE", "", "| 平台／模組／位址 | 原始名稱 |", "|---|---|"])
        for f in pending:
            out.append(f"| {f['platform']}/{f['module']}/{f['ea']:X}h | `{f['ida_name']}` |")
    if unexplained:
        out.extend(["", "## 無直接 caller 且仍無解釋", "", "| 平台／模組／位址 | 原始名稱 |", "|---|---|"])
        for f in unexplained:
            out.append(f"| {f['platform']}/{f['module']}/{f['ea']:X}h | `{f['ida_name']}` |")
    return "\n".join(out) + "\n", len(pending), len(unexplained)


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("--input", default="docs/audit/coab-function-index.json")
    parser.add_argument("--output", default="docs/audit/function-completion-audit.md")
    args = parser.parse_args()
    with open(args.input, encoding="utf-8") as handle:
        functions = json.load(handle)["functions"]
    report, pending, unexplained = render(functions)
    with open(args.output, "w", encoding="utf-8") as handle:
        handle.write(report)
    print(f"functions={len(functions)} pending={pending} unexplained_caller0={unexplained}")
    return 1 if pending or unexplained else 0


if __name__ == "__main__":
    raise SystemExit(main())
