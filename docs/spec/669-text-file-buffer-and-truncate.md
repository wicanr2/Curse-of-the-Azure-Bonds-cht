# 第六百六十九輪：文字檔的緩衝寫出、模式檢查，與依 `^Z` 截斷

狀態：`READY`。等級：`exact`。日期：2026-08-14
位置：PC-98 `PC98-GAME.EXE` 的 `1BCAAh`、`1BC46h`、`1BC86h`、`1BB5Eh`；
`1BCB9h` 改判為邊界碎片。

## `1BCAAh`：放一個字元進緩衝，滿了就沖出去

完整範圍是 `1BCAAh..1BCE9h`（IDA 在 `1BCB9h` 切成兩段，那裡不是函式起點）。

```text
es:[bx+di] := al                       ← 寫進緩衝
bx := bx + 1
if bx <> dx then return                 ← 還沒滿（dx ＝ BufSize）
（滿了）
di := sp ; es:di := ss:[di+2]           ← 從堆疊取回 TextRec 指標
es:[di+8] := bx                          ← 回寫 BufPos
call dword ptr es:[di+14h]               ← InOutFunc
if AX <> 0 then word_23B08h := AX        ← 記錄 InOutRes
重新載入 ax := BufEnd, bx := BufPos, dx := BufSize, es:di := BufPtr
```

沖出去之後**四個暫存器全部重新載入**——`InOutFunc` 可能改動 `BufPos`／`BufPtr`，
所以不能沿用舊值。

錯誤只記進 `word_23B08h`（`InOutRes`），**不中止**：緩衝照樣繼續用。

## `1BC46h` 與 `1BC86h`：模式檢查

```text
1BC46h（輸入）:
    if InOutRes = 0 且 Mode <> fmInput then word_23B08h := 68h    ← 錯誤 104
    bx := BufPos(+8) ; dx := BufEnd(+0Ah) ; es:di := BufPtr(+0Ch)

1BC86h（輸出）:
    if InOutRes = 0 且 Mode <> fmOutput then word_23B08h := 69h   ← 錯誤 105
    bx := BufPos(+8) ; dx := BufSize(+4) ; es:di := BufPtr(+0Ch)
```

`68h` ＝ 104、`69h` ＝ 105，是 Turbo Pascal 的「檔案未開啟供輸入／輸出」。

兩處的錯誤路徑都是 `jmp` 回**同一段載入程式碼**——**設完錯誤碼還是照樣載入指標並
返回**，呼叫端若不查 `InOutRes` 就會拿著沒開啟的檔案繼續寫。

兩支的 `dx` 取的欄位不同：輸入取 `BufEnd`（讀到哪裡）、輸出取 `BufSize`（可以寫到
哪裡）。同一個暫存器在兩條路徑上意義不同。

## `1BB5Eh`：讀最後 128 bytes 找 `^Z` 再截斷

```text
(dx:ax) := 檔案大小                       ← int 21h AX=4202h，位移 0
(dx:ax) := (dx:ax) − 80h
if 借位 then (dx:ax) := 0                  ← 檔案比 128 短就從頭
seek 到該處                                ← AX=4200h
讀 128 bytes 進 [di+80h]                   ← AH=3Fh，緩衝就是 TextRec 的 Buffer
if 讀取失敗 then ax := 0
bx := 0
while bx <> ax do                          ← ax ＝ 實際讀到的位元組數
    if [bx+di+80h] = 1Ah then 跳出
    bx := bx + 1
若沒找到就直接返回                          ← 不截斷
dx := bx − ax ; cx := 0FFFFh               ← 相對檔尾的負位移
seek                                       ← AX=4202h
寫 0 bytes                                 ← AH=40h，CX=0 ⇒ 截斷
```

`1Ah` 是 DOS 的文字檔結束標記（Ctrl-Z）。這支把它之後的內容切掉——**用「寫入 0
bytes」達成截斷**，是 DOS 的標準手法。

只看**最後 128 bytes**。`^Z` 若在更前面（檔案後面還有超過 128 bytes 的垃圾），
這支找不到，也就不會截斷。

## `1BCB9h`：邊界碎片

`1BCAAh` 那一支的中段。IDA 把 `mov di, sp` 之後切成新函式，但它沒有 prologue、也
不是任何呼叫的目標——是 `bx = dx` 時的落點。

## 明確不宣稱

- `InOutFunc` 實作在哪、回傳值的完整語意。
- `1BB5Eh` 由誰呼叫（`Close` 或 `Flush`）。
