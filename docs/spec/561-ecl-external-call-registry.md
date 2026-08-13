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
| `B200h` | `3201h` | `2FCFh` | `3193h` | 0 |
| `C018h` | `4019h` | `300Eh` | `31D2h` | 0 |
| `C01Eh` | `401Fh` | `3002h` | `31C6h` | 0 |
| `2E10h` | `AE11h` | `2F31h` | `30F5h` | 12 |
| `6803h` | `E804h` | `3035h` | `31F9h` | 11 |

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

`docs/audit/ecl-event-catalog.json` 的靜態清冊裡共有 23 個 `CALL` 指令，
operand 只有兩種：`2E10h`（12 次）與 `6803h`（11 次）。這正是完整度矩陣裡
「23 個靜態可達 CALL」的來源。

所以本輪把兩件事分開了：

- **engine 認得 7 個**（本規格，`exact`）。
- **CoAB 的 ECL1–ECL6 靜態使用 2 個**（清冊，`exact`）。

先前文件寫「CoAB 已接三個 observed address」（`2E10h`、`C01Eh`、`B200h`）。
後兩者不在靜態 corpus 裡，但確實被 handler 認得。它們的來源是別條路徑
（動態觀察或其他成員），本規格不推翻也不採信該敘述，只把可重生的事實分成
上面兩層；要把 `C01Eh`／`B200h` 寫成 CoAB 實際會走到的路徑，必須另外提出
runtime trace 或指出它們在哪個 DAX 成員出現。

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

- **任何 routine 的效果**。分支主體位址是 `exact`，但每個分支做什麼——
  `2E10h` 的地圖／重繪語意、`6803h` 的用途、`B200h`／`C018h`／`C01Eh`／
  `8000h`／`8001h` 是什麼——全部仍是 `待解讀`。第 519／520 輪已對 `2E10h`
  追到 overlay-30 的 cell-layer accessor，那條結論不受本規格影響也不被它擴大。
- **`8000h`／`8001h` 兩個分支只差一個常數**（PC-98`3173h` 推 `1`、`3183h`
  推 `0`，之後呼叫同一個 far routine）。這看起來像同一功能的開關，但在證明
  那個 routine 是什麼之前不得命名。
- **跨作品通用性**。其他 Gold Box 作品的 selector 表必須各自解，不得沿用。
