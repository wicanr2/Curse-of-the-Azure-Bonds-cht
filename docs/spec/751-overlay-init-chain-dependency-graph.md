# 751 — overlay 的 unit 初始化鏈與模組相依圖

- 平台／模組：DOS 與 PC-98 全 36 個 overlay
- 證據等級：`exact`（直接讀原始 bytes，不經 IDA 匯出）
- 工具：`scripts/overlay_init_graph.py` → `docs/audit/overlay-init-graph.md`

## 形狀

Turbo Pascal 把 unit 的初始化段編在模組的 `0000h`，形狀固定：

```
55 89 E5            push bp / mov bp, sp
9A off seg  × N     依序呼叫每個相依 unit 的初始化段
89 EC 5D CB         mov sp, bp / pop bp / retf
```

判準要求三段全中：起頭必須是 `55 89 E5`，中間是**連續**的 `9A`，緊接著必須正好
是 `89 EC 5D CB`。兩平台各 36 個模組，**各有 32 個完全符合**；不符的四個兩平台
相同：`overlay-00`、`overlay-01`、`overlay-13`、`overlay-18`。

## 每個目標都是別的 overlay 的 `0000h`

把 far call 的 `seg:off` 拿去查 VROOMM stub 表，除了一個例外，**全部落在某個
overlay 的 entry 且 `code_offset` 為 `0000h`**。也就是說目標就是那個模組的初始化
段——這正是 unit 相依關係，所以 `0000h` 的呼叫清單可以直接當成「這個模組直接
uses 哪些模組」。

唯一例外：`overlay-08`／`overlay-09`／`overlay-10` 都呼叫
`overlay-13` 的 `+1D9Dh` 而不是 `0000h`——而 `overlay-13` 的 `0000h` 正好是四個
「形狀不符」之一。兩件事互相呼應：`overlay-13` 的初始化段不在 `0000h`。

## 結果

完整表在 `docs/audit/overlay-init-graph.md`。幾個直接看得出來的結構：

- `overlay-34` 沒有任何相依，而幾乎每個模組都相依它——它是最底層的 unit。
- `overlay-27`、`overlay-31`、`overlay-35` 同樣沒有相依（葉節點）。
- 相依最多的是 `overlay-02`（15 個），其次 `overlay-08`（14）、`overlay-10`（13）。
- **兩平台的相依圖逐列相同**，包含順序。移植沒有改動 unit 結構。

## 為什麼一定要讀原始 bytes

IDA 的匯出在這段會出錯：`89 EC`（`mov sp, bp`）常被吃掉一個 byte 而變成
`EC`＝`in al, dx`，連帶把後面的位元組錯開，於是 `overlay-04:0000h` 在匯出裡
只看得到 7 個 far call ＋ 兩條垃圾指令，實際是 9 個（見 spec 736 對同一種漏
byte 的紀錄）。本輪的相依圖全部由 `.bin` 的原始 bytes 解出。

## 明確不宣稱

- 沒有宣稱這張圖等於「執行期的載入順序」。它是**編譯期的 uses 關係**；
  VROOMM 什麼時候把哪個 overlay 換進記憶體是另一回事。
- 沒有宣稱四個「形狀不符」的模組沒有初始化段——只是不在 `0000h`、或形狀不同，
  本輪沒有逐一追。
- 沒有宣稱相依清單的**順序**有語意（Turbo Pascal 依 uses 子句順序產生，但沒有
  在本輪驗證）。
