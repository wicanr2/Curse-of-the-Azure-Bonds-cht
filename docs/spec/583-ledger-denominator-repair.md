# 第五百八十三輪：修正台帳的分母

狀態：`READY`。日期：2026-08-14

## 問題

台帳的分母是「IDA 認出的函式數」。讀 `LOSEDUDE` 時發現 body 裡有兩處位元組
IDA 完全沒認成指令，順著查下去，發現更大的缺口：

- **DOS 側有幾個 overlay 的指令覆蓋率只有 25%～68%**（`11`／`25`／`04`／
  `15`／`17`）。
- **overlay control block 宣告的 entry `code_offset` 有一批沒有成為函式**——
  `overlay-15` 連 `0000h`（unit init）都不是。

entry stub 宣告的入口一定是函式，沒成為函式就是分析漏了。所以分母本身不完整，
「解讀完 N 個」這個說法建立在錯的 N 上。

## 修法

判準只有兩條，都保守：

1. 位址沒有被任何既有函式涵蓋，且開頭是 `55 89 e5`（`push bp / mov bp,sp`），
   Turbo Pascal 的標準 prologue。不猜其他 prologue 形式。
2. overlay control block 宣告的 entry `code_offset`（**排除 `FFFFh`——那是
   未使用的 slot，不是位址**）。

`ida_funcs.add_func(ea)` 對這批位址**一律失敗**，即使該位址已經是 code、不屬於
任何函式、prologue 也完好。原因是 IDA 決定不了函式結尾（這些 overlay 的自動
分析本來就沒走到那裡）。改用 `ida_funcs.find_func_bounds()` 先求邊界再
`add_func(ea, end)`，推不出來才退回「下一個 seed／下一個既有函式起點／段尾」
取最近者。

## 結果

| | 修正前 | 修正後 |
|---|---:|---:|
| 台帳函式數 | 2,825 | **2,875** |
| 指令覆蓋率 | 85.4% | **89.1%** |
| 已解讀 | 571 | 571 |
| 不阻塞 | 162 | 162 |
| 邊界碎片 | 488 | 488 |
| 待解讀 | 1,604 | 1,654 |

既有判讀一筆沒丟。**待解讀從 1,604 變成 1,654**——分母變大，剩下的工作跟著
變多，這是修正的正確方向。

還有 11 個位址補不上（多數是誤報的 prologue，落在字串或資料中間）。這個數字
如實留著，不當成 0。

## 四個「安靜地給出錯結果」的坑

這次每一個都踩到，而且**沒有一個會報錯**：

1. **`ida_pro.qexit()` 不儲存資料庫。** 少了
   `ida_loader.save_database(...)`，補上的函式只活在那次執行的記憶體裡，
   `.i64` 完全沒變，重新匯出看到的還是舊清單。
2. **輸出路徑必須落在掛載點內。** `tools/ida.sh` 把「i64 所在目錄」掛成
   `/work`，所以 `/work/../out/x.json` 在容器裡是 `/out`——不存在，寫入被
   丟棄，`out/` 的時間戳完全不動，看起來像「跑完了但沒變」。
3. **把 `x.bin.i64` 交給 `idat` 會從 `x.bin` 重新載入並覆蓋資料庫。**
   `overlay-25` 因此從 15 個函式變回 8 個。重新匯出時只能跑純讀取的
   `export_module.py`，不能再讓 IDA 重新分析。
4. **`export_module.py` 不寫 `overlay` 欄位**（那是 `analyze_overlay.py` 的
   產物）。少了它，`cmd/re-ledger` 退回用 `input.name`，模組名變成
   `overlay-25.bin`，整個台帳 join 不上——「已解讀」從 571 無聲地掉到 271。
   現在由 `scripts/attach_overlay_field.py` 在主機端從 manifest 補回。

第 4 點特別值得記：**數字掉了一半，程式沒有任何錯誤輸出。** 分母與判讀是
分開產生再 join 的，join 失敗只會表現成「數字變了」。所以每次動到匯出流程，
都要拿四個計數（已解讀／不阻塞／邊界碎片／待解讀）與動之前對照。

## 工具

- `scripts/coverage_audit.py`——量指令覆蓋率。**位元組要取集合聯集再算**，
  各函式長度相加會超過 100%（共用 epilogue 同時屬於好幾支，`dos/overlay-09`
  就是這樣冒出 101.3% 的）。
- `scripts/seed_missing.py` ＋ `tools/ida/seed_missing_functions.py`——找缺口
  並補上，逐一回報失敗原因。
- `tools/reexport-overlays.sh` ＋ `scripts/attach_overlay_field.py`——重新
  匯出並補 `overlay` 欄位。
