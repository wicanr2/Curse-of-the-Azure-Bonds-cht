# 第五百六十一輪：ECL `CALL` external routine registry（兩平台）

狀態：`READY`（限 selector 集合與 operand 換算）；每個 routine 的實際效果仍是
`待解讀`。日期：2026-08-14

## 結論先行

ECL `CALL`（opcode `2Dh`）的 handler 已在兩平台完整解出，**engine 認得的
external routine 共 7 個，兩平台完全相同**：

| ECL operand | selector | DOS 分支 | PC-98 分支 | CoAB corpus 使用次數 |
|---|---|---|---|---:|
| `8000h` | `0001h` | `2FAFh` | `3173h` | 0 |
| `8001h` | `0002h` | `2FBFh` | `3183h` | 0 |
| `B200h` | `3201h` | `2FCFh` | `3193h` | 19 |
| `C018h` | `4019h` | `300Eh` | `31D2h` | 0 |
| `C01Eh` | `401Fh` | `3002h` | `31C6h` | 13 |
| `2E10h` | `AE11h` | `2F31h` | `30F5h` | 125 |
| `6803h` | `E804h` | `3035h` | `31F9h` | 11 |

使用次數是沿 `ecl.TraceGraph` 走完 25 個 block 的**可達**條數，合計 168，與
`cmd/ecl-effect-coverage` 的 `0x2D` 分母一致。每一支實際做什麼見 spec 1150。

其餘 operand 一律走到 handler 的 epilogue：**不做事直接返回**，不是錯誤路徑。

產物：[`docs/audit/ecl-external-call-registry.md`](../audit/ecl-external-call-registry.md)。

## operand 與 selector 的換算

handler 先取得 operand word，再 `sub ax, 7FFFh`，之後才逐一比對 selector：

```text
DOS   overlay-02 2F02h：… call far ptr（取 operand）→ 2F23h `sub ax, 7FFFh`
PC-98 overlay-02 30C6h：… call far ptr（取 operand）→ 30E7h `sub ax, 7FFFh`
```

因此

```text
selector = (operand - 7FFFh) mod 10000h
operand  = (selector + 7FFFh) mod 10000h
```

`7FFFh` 是從指令的立即數直接讀出的，不是湊出來的常數。這條換算同時**獨立
驗證了第 519 輪**的結論：該輪由另一條路徑得到「ECL operand `2E10h` →
selector `AE11h`」，與本輪一致。

## 與 CoAB corpus 的關係

兩件事要分開：

- **engine 認得 7 個**（本規格，`exact`）。
- **CoAB 的 ECL1–ECL6 用到 4 個**：`2E10h`、`B200h`、`C01Eh`、`6803h`
  （上表，`exact`）。`8000h`／`8001h`／`C018h` 認得但沒有腳本用。

⚠ 舊產物曾少報為 4,222 條指令、55 個 opcode、`2Dh` 78 條；2026-08-27 已依
spec 1219 重生 `docs/audit/ecl-event-catalog.json`，現為 14,177 條、61 個 opcode、
`2Dh` 168 條。要分母就用 `ecl.TraceGraph`
（`cmd/ecl-window`、`cmd/ecl-cell-refs`、`cmd/ecl-effect-coverage` 走同一條路）。

## 方法

與 spec 560 相同的逐值符號執行，但終止條件不同：這裡要的是「selector 落到
哪個分支主體」，所以走到**第一個不屬於比較鏈的指令**就停，而不是走到第一個
`call`（CALL handler 的每個分支內部都有數個 far call）。

```sh
tools/ida.sh py workplace/re-sweep/dos/overlays/overlay-02.bin.i64 \
  dump_function.py /work/call-handler.json 2F02
tools/ida.sh py workplace/re-sweep/pc98/overlays/overlay-02.bin.i64 \
  dump_function.py /work/call-handler.json 30C6
python3 scripts/ecl_call_registry.py --merge \
  dos=workplace/re-sweep/dos/overlays/call-handler.json \
  pc98=workplace/re-sweep/pc98/overlays/call-handler.json \
  > docs/audit/ecl-external-call-registry.md
```

工具曾用「selector 最多的分支＝default」的啟發式來過濾未識別值，結果在每個
分支各只有一個 selector 時，把真的 selector `0001h` 當成 default 刪掉。現在
改成：走到 epilogue 的值根本不會進表，進表的就是具名 selector。
**規則：不要用「以量取勝」去猜哪個分支是預設值。**

## 這份規格明確不宣稱

- **任何 routine 的效果**。分支主體位址是 `exact`；每個分支做什麼在 spec 1150
  才逐條讀出（`2E10h` 髒了才重畫、`6803h` 推圖片序列一格、`B200h` 播 10 或 11
  號音效、`C01Eh` 是 `MOVEFORWARD`、`8000h`／`8001h` 是 `GODUEL(1)`／`GODUEL(0)`、
  `C018h` 有條件重算牆面碼）。
- **`8000h`／`8001h` 兩個分支只差一個常數**（PC-98`3173h` 推 `1`、`3183h`
  推 `0`，之後呼叫同一個 far routine ＝ `ECL2.GODUEL`）。這看起來像同一功能的
  開關，但 `GODUEL` 內容本身仍是 `待解讀`。
- **跨作品通用性**。其他 Gold Box 作品的 selector 表必須各自解，不得沿用。
