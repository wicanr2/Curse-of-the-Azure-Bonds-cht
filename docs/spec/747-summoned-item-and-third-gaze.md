# 第七百四十七輪：召喚物品的完整生命週期，與第三處麻痺凝視

狀態：`READY`。等級：`exact`。日期：2026-08-14
位置：DOS `overlay-12` 的 `05B6h`、`289Ch`（兩支都沒有位元組缺口）。

## `05B6h`：套用與解除共用一支

```text
; 先在物品鏈上找有沒有「類別 14h 且 +31h = 0F3h」的那一件
item := 角色^[14Dh]；found := 0
while item <> NIL 且 found = 0 do
    if item^[2Eh] = 14h 且 item^[31h] = 0F3h then found := 1
    else item := item^[2Ah]

if arg_0 <> 0 then                                    ← 解除
    if item <> NIL then <overlay-24 entry#17>(角色, item)   ← spec 695 的移除並釋放
    goto 收尾

if found <> 0 then goto 收尾                          ← 已經有一件就不再給
if 角色^[14Ch] >= 10h then goto 收尾                  ← ⚠ 物品數量上限 16

GetMem(node, 3Fh)
FillChar(node^, 3Fh, 0)
node^[2Eh] := 14h        ← 類別
node^[30h] := 14h        ← 次數（spec 703 的 +30h）
node^[31h] := 0F3h       ← 識別用
node^[32h] := 1          ← ⚠ 讓 spec 737 的「打不到」規則放行
node^[3Dh] := 17h        ← 第二個效果槽（spec 703）
node^[3Eh] := 0A0h       ← 第三個效果槽
<overlay-24 entry#18>(角色, node)                     ← spec 695：掛的是**複本**
<overlay-19 entry#7>(node)
FreeMem(node, 3Fh)                                    ← 本地那份用完就釋放
<overlay-24 entry#20>(角色, '<名字> Gains an item'（CS:05A8h）, 0Ah, 1)

收尾：<overlay-24 entry#7>(角色)
```

### 這支示範了 spec 695 那條「掛的是複本」為什麼重要

節點在**堆疊之外的暫存**建好、複製進鏈、然後**立刻 `FreeMem`**。如果
`entry#18` 掛的是來源本身而不是複本，這裡的 `FreeMem` 就會把剛掛上去的節點
釋放掉。spec 695 當時只記下「掛的是複本」這個事實，這裡看到它被依賴的地方。

### 三個新確定的欄位

- **`角色^[14Ch]` 是物品數量，上限 `10h` ＝ 16。** 超過就不給新物品，
  而且是**靜默失敗**——連訊息都沒有。
- `node^[32h] := 1`：spec 737 的 `25BAh` 用 `w^[32h] = 0` 當「打不到」的條件之一，
  所以這件召喚出來的武器**天生就能打到需要魔法武器的目標**。
- `node^[31h] := 0F3h` 與 `[2Eh] = 14h` 這一對是它的識別碼——搜尋時兩個都要比，
  只比類別會誤判別的 `14h` 物品。

## `289Ch`：第三處麻痺凝視

```text
p := 角色^[18Dh]^[0Ah]                                ← ⚠ 直接讀，沒有 NIL 檢查
DS:6FA3h := p
<overlay-24 entry#20>(角色, '<名字> gazes...'（CS:2886h）, 0Ah, 0)
<overlay-24 entry#24>(12h)
<overlay-24 entry#25>(自己座標, p 座標, 4, 2Dh)
if <overlay-23 entry#8>(p, 1, 0) = 0 then
    <overlay-23 entry#11>(p, 34h, 0, 0FFh, 0)
    <overlay-24 entry#20>(p, '<名字> is paralyzed'（CS:288Fh）, 0Ah, 0)
```

麻痺（效果 `34h`）現在有三個入口，`entry#11` 的第三個參數各不相同：

| 位置 | 第三個參數 | 選目標 |
|---|---|---|
| `overlay-22:6022h`（spec 722） | `3Ch` | 先清 `NIL` 再用分派表選 |
| `overlay-12:2782h`（spec 741） | `0` | 直接讀 `+18Dh^[0Ah]` |
| `overlay-12:289Ch`（本輪） | `0` | 直接讀，**沒有 `NIL` 檢查** |

`gazes...` 與 `is paralyzed` 兩句在 `overlay-22` 與 `overlay-12` **各有一份**
（`docs/audit/duplicate-strings.md` 的跨模組重複）。中文化兩邊都要改。

`entry#20` 最後一個參數在這支是 `0`，spec 741 的 `2782h` 是 `1`。同樣是麻痺訊息，
兩支傳的旗標不同。

## 明確不宣稱

- `+14Ch`（數量）與 `+14Dh`（鏈頭）之外，物品清單還有沒有別的欄位。
- `overlay-19 entry#7`、`overlay-24 entry#7`／`entry#18` 的行為。
- 效果編號 `17h`／`0A0h` 代表什麼。
- `entry#20` 最後那個 `0`／`1` 的差別。
