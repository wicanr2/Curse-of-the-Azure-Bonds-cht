# 第七百二十四輪：恐懼，與 `PUTEFFECT` 的參數位置

狀態：`READY`。等級：`exact`。日期：2026-08-14
位置：DOS `overlay-22:4A8Fh`。

## `4A8Fh`：恐懼

```text
a, b := entry#15/#16(DS:6506h)
<sub_175Bh>(a, b, DS:7559h, DS:755Ah, 3, 6)

for k := 1 to DS:7434h do
    p := far [7431h + k×4]                       ← ⚠ 沒有 NIL 檢查（spec 716）
    s := <overlay-23 entry#8>(p, 4, 0)
    if s = 0 then                                 ← 豁免失敗
        d := <sub_E11h>(54h)                      ← 持續時間（spec 712）
        <overlay-23 entry#21（PUTEFFECT）>(p, 8Eh, d, 0, 1, 1, s, '<名字> runs in terror'（CS:4A72h）)
        p^[18Dh]^[10h] := 1
        p^[198h]       := 1
        if p^[0F7h] <= 7Fh then p^[0F7h] := 0B3h
        p^[18Dh]^[0Ah] := NIL                     ← 清掉它鎖定的目標
    else
        <overlay-24 entry#20>(p, '<名字> is unaffected'（CS:4A81h）, 0Ah, 1)
```

`+0F7h` 就是 `KILLTHEDUDE` 用來判斷「要不要把英文名換成片假名」的那個欄位
（PC-98 側 spec 記過）。這裡是「小於等於 `7Fh` 就設成 `0B3h`」——只動一次，
已經大於 `7Fh` 的不動。

`p^[18Dh]^[0Ah]` 被清成 `NIL`，正是 spec 722 那個「選目標的結果欄位」。所以
被嚇跑的單位會失去它原本鎖定的目標。

## `PUTEFFECT` 的參數位置

`overlay-23 entry#21` 是 `retf 14h` ＝ 10 個 word（spec 711）。三個呼叫點把每個
位置對上了：

```text
PUTEFFECT(目標 far, 效果編號, 持續時間, ?, ?, ?, ?, 訊息 far)
           2 word     1        1       1  1  1  1     2
```

| 呼叫端 | 目標 | 效果編號 | 持續時間 | 中間四個 |
|---|---|---|---|---|
| `4432h`（spec 703） | `DS:6506h` | `4` | `sub_E11h(49h)` | `0, 0, 0, 0` |
| `4822h`（spec 711） | 目標陣列第 k 格 | `23h` | `sub_E11h(52h)` | `0, 0, 1, s` |
| `4A8Fh`（本輪） | 目標陣列第 k 格 | `8Eh` | `sub_E11h(54h)` | `0, 1, 1, s` |

**效果編號和持續時間的 id 是兩個不同的數字**（`23h`／`52h`、`8Eh`／`54h`）。
先前只看一個呼叫點時很容易把它們當成同一個。

第 10 個 word 一律是 `0A54:0634h` 留在堆疊上的字串結果（spec 690），
第 7 個位置在兩支裡是 `entry#8` 的回傳值 `s`——也就是**豁免結果會一起傳進去**。

## 明確不宣稱

- `PUTEFFECT` 中間四個參數的意義。
- `+198h`、`+18Dh^[10h]`、`+0F7h := 0B3h` 的語意。
- `sub_175Bh` 的行為與它最後兩個常數 `3`／`6`。
