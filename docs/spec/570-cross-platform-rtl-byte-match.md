# 第五百七十輪：以逐位元組比對把 RTL 名稱轉移到 PC-98

狀態：`READY`（限本輪的比對判準與結果）。日期：2026-08-14

## 結論先行

PC-98 `GAME.EXE` 沒有可被 FLIRT 辨識的 RTL 名稱（333 個函式只有 `start` 一個
具名）。但兩個版本用同一套 Turbo Pascal RTL 編譯，因此可以用**逐位元組比對**
把 DOS 側的識別轉移過去：

- 44 個 PC-98 resident 函式的 body 與 DOS `START.EXE` 的某個函式**完全相同**。
- 其中 31 個在 DOS 側有 Borland 還原的名稱。

判準刻意保守：body bytes 完全相同、長度 `>= 16` bytes、DOS 側名稱唯一。
太短的 body 可能碰巧相同（只有 prologue 加一個 `call` 的樁），不足以當證據。

轉移的是**識別**（這是 RTL 的哪一支），不是語意。與第 566 輪一致，
`Random`／`Randomize`／`SOUND`／`NOSOUND`／`DELAY` 若出現一律保留 `待解讀`。

## 順帶的重要發現：PC-98 的 `Random` 位置與同一性

比對把 PC-98 的 `@Random$q4Word` 定位在 **`1B7A5h`**，而且它與 DOS 的
`1B645h` **逐位元組相同**。這代表兩平台的亂數演算法一致——日後要重現原版
隨機序列時，只需解一次 LCG，不必分平台各做一次。

依既有規則它**不標不阻塞**，維持 `待解讀`（見 spec 566）。

**種子來源也已讀出**（第 572 輪）：DOS `@Randomize`（`START.EXE:1B6CCh`）以
`INT 21h AH=2Ch` 取系統時間，把 `CX:DX` 存進 `dword_20998`——**那就是 Turbo
Pascal 的 `RandSeed`**。要重現原版隨機序列，需要的是這個種子加上 `Random` 的
LCG，兩邊都已定位。

## 為什麼這是硬證據

同一段 bytes 在 8086 上就是同一段行為。它不依賴名稱相似、位址推算或
entry index 對應——那些都是 `strong inference`，這一條是 `exact` 層級的
識別依據。

反過來說，**沒有比對成功不代表不是 RTL**：PC-98 版的 RTL 可能因不同編譯選項
或不同版本而有差異。本輪只認完全相同的那 44 個。

## 重生

```sh
tools/ida.sh py workplace/re-sweep/pc98/PC98-GAME.EXE.i64 \
  export_small_functions.py /work/small/PC98-GAME.EXE.json 4096
tools/ida.sh py workplace/re-sweep/dos/START.EXE.i64 \
  export_small_functions.py /work/small/START.EXE.big.json 4096
python3 scripts/cross_platform_rtl_match.py --write
```

（`export_small_functions.py` 的長度上限調到 4096 以涵蓋整個 resident。）

## 這份規格明確不宣稱

- **未比對成功的 289 個 PC-98 resident 函式的性質**。它們維持 `待解讀`。
- **overlay 之間的跨平台比對**。overlay 的 code offset 兩平台不同，且內容有
  實質差異（PC-98 多了日文與音源），不適用這條判準。
