# 第六百五十四輪：overlay 緩衝的最佳適配配置器

狀態：`READY`。等級：`exact`。日期：2026-08-14
位置：PC-98 `PC98-GAME.EXE` 的 `185EAh`、`186FFh`、`18767h`。

## 串列節點

三支共用同一條串列，頭在 `DS:word_2418Eh`，節點以 **segment 值**串接：

| 位移 | 內容 |
|---|---|
| `+0` | 這一塊的大小（word） |
| `+4` | 下一塊的 segment（`0` 表示結尾） |

節點位址就是 segment 本身（`mov es, si` ＋ `es:[di]`，`di` 恆為 0）。

## `18767h`：最佳適配

```text
輸入：AX ＝ 需要的大小
si := word_2418Eh ; cx := 0FFFFh ; bx := 0FFFFh
while si <> 0 do
    surplus := es:[si:0] − 需求
    if surplus < 0 then 跳過（太小）
    else if surplus = 0 then bx := si ; break        ← 剛好就直接用
    else if surplus < cx then bx := si ; cx := surplus
    si := es:[si:4]
return bx + 1
```

**剛好夠**（`surplus = 0`）就立刻採用、不再往下找；否則挑**剩最少**的那一塊。
找不到時 `bx` 仍是 `0FFFFh`，`inc bx` 之後回傳 **0**——所以呼叫端用 0 判斷失敗。

回傳的是**節點 segment ＋ 1**，也就是跳過節點自己佔的那一段。

比較用 `jb`／`jbe`（**無號**）。

## `186FFh`：最大可用塊

```text
bx := word_24188h
if bx <> 0 then bx := bx − 1                 ← 起始值先減一
si := word_2418Eh
while si <> 0 do
    if es:[si:0] > bx then bx := es:[si:0]
    si := es:[si:4]
return AX ＝ bx，DX ＝ word_24186h
```

回傳串列裡最大的一塊，但**下限是 `word_24188h − 1`**——如果串列全部比它小，回傳的
是那個減一後的值而不是實際最大值。

## `185EAh`：釋放 DOS 記憶體塊

```text
if word_24182h = 0 then return 0FFFFh
ES := word_24182h
int 21h (AH = 49h)                            ← DOS 釋放記憶體
word_24182h := 0                              ← 先清掉再看結果
if 進位（失敗）then return 0FFFFh
return 0
```

**先把 `word_24182h` 清成 0，再檢查 `int 21h` 的進位旗標**。所以釋放失敗時那個
全域已經被清掉了——**同一塊記憶體再也不會被嘗試釋放**，會漏掉。

回傳值慣例與其他兩支相反：這支 `0` 表示成功、`0FFFFh` 表示失敗。

## 明確不宣稱

- `word_24182h`／`word_24186h`／`word_24188h` 各自的角色。
- 節點 `+2` 那個 word 的用途（三支都沒有讀它）。
- 這條串列裝的是 overlay 緩衝、一般堆積、還是兩者。
