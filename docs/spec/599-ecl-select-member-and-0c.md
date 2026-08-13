# 第五百九十九輪：`0Ah`（選第 N 名隊員）與 `0Ch`

狀態：`READY`。等級：`exact`。日期：2026-08-14
模組：`INTERPET`（overlay-02）。位址為 PC-98 overlay-local。

## `0Ah`（`0306h`）：把目前目標換成第 N 名隊員

```text
DS:7898h := 1
READVAR(1)
v := ADDRESSVALUE(1)
n    := v and 7Fh                         ← 低 7 位是索引（0-based）
extra := v and 80h                        ← 最高位是額外動作的開關
node := DS:9598h ; i := 0
while n > 0 and i < n and node <> nil do
    node := node^[18Ah] ; i := i + 1
if node <> nil then DS:9594h := node ; DS:BDFFh := 0
else                                       DS:BDFFh := 1     ← 索引超出範圍
if extra <> 0 and DS:BDE4h <> 0 and DS:BDE5h <> 0 then
    if DS:789Dh = DS:9594h then DS:7898h := 0
    <far 00E9:002F>(1, 0) ; <far 014A:002A>(DS:9594h)
    DS:BDE4h := 0 ; DS:BDE5h := 0
```

- **索引是 0-based**：`n = 0` 時迴圈不執行，直接取鏈頭。
- **operand 的最高位是旗標不是數值**——只有低 7 位參與索引。`36h` 寫
  `+0F7h` 時也是把最高位當標記（[spec 598](598-ecl-add-npc.md)），
  這是這個引擎反覆出現的手法。
- 索引超出範圍時**不改 `DS:9594h`**，只把 `DS:BDFFh` 設為 1。目前目標維持
  原樣，不是變成 nil。

## `0Ch`（`03E0h`）

```text
saved := DS:A2A8h
READVAR(3)
DS:A893h    := ADDRESSVALUE(1)
bank1^[580h] := ADDRESSVALUE(2)
DS:A894h    := ADDRESSVALUE(3)
bank1^[582h] := <far 0062:003E>(DS:A2ABh, DS:A2AAh, DS:A2A9h)
if bank1^[580h] < bank1^[582h] then bank1^[582h] := bank1^[580h]   ← 取較小者
<far 0062:0048>(DS:BDDAh, bank1^[582h], DS:A894h, DS:A893h)
<far 0064:003E>()
DS:A2A8h := saved
```

與 `0Dh`（[spec 588](588-ecl-return-and-gosub-stack.md)）成對：兩者都存下
`DS:A2A8h` 再還原，都以 `(DS:BDDAh, bank1^[582h], DS:A894h, DS:A893h)` 呼叫
`0062:0048`。差別是 `0Ch` **算出**這個計數（並以 operand 2 為上限），`0Dh`
則是把它**遞減一格**。

`0062:003E` 的三個參數 `DS:A2A9h`／`A2AAh`／`A2ABh` 就是虛擬地圖暫存器
`C04Bh`／`C04Ch`／`C04Dh` 的實體位址（[spec 563](563-ecl-memory-model-and-operand-resolution.md)），
所以那個值是從**當前地圖座標**算出來的。

## `32h`（`29FBh`）：全隊有沒有某物品

```text
READVAR(1)
id := ADDRESSVALUE(1)
FillChar(DS:A88Ah, 6, 0)
DS:A88Bh := 1                             ← 預設「沒有」
node := DS:9598h ; found := false
while node <> nil and not found do
    item := node^[14Eh]
    while item <> nil and not found do
        if id = item^[56h] then
            DS:A88Ah := 1 ; DS:A88Bh := 0
            found := true
        item := item^[52h]
    node := node^[18Ah]
```

與 `3Fh`（查效果，[spec 595](595-ecl-target-selection-and-effect-query.md)）
一樣**借用比較旗標**，所以後面接 `16h` 是「有」、`17h` 是「沒有」。

兩層迴圈的條件都帶著 `found`，所以**找到第一個就整個停下來**——與 `40h`
（移除物品，[spec 596](596-ecl-party-item-sweep.md)）走遍全部不同。

物品鏈的欄位與 `40h` 一致：`+52h` next、`+56h` 類型 id，掛在角色 `+14Eh`。

## `2Ch`（`2940h`）：四選一的選單

```text
READVAR(6)
for i := 0 to 3 do v[i] := ADDRESSVALUE(i + 1)    ← 前四個 operand 是候選值
Move(DS:00ACh, DS:A33Fh, 46h)                     ← 兩段從 DS 複製過去的資料
Move(DS:00F2h, DS:A335h, 0Ah)
choice := <far 0062:0084>(1, 0, 0, 0Fh, 0Ah, 0Dh, @buf, @buf)
dest := ADDFNC(high[5], low[5])
STOREVALUE(dest, v[choice])
<far 0064:003E>()
```

**前四個 operand 是候選值、第五個是目的位址**；宣稱 arity 6，但**第六個
operand 在 body 裡沒有被用到**（與 `2Ah` 一樣，[spec 594](594-ecl-random-and-indexed-store.md)）。

`choice` 直接當索引取 `v[choice]`，**沒有範圍檢查**——選單常式回傳 0..3 之外
的值就會讀到 frame 上的其他位元組。

## 明確不宣稱

- `bank1^[580h]`／`[582h]`、`DS:BDDAh`、`DS:A893h`／`A894h` 的語意。
- `DS:BDE4h`／`BDE5h`／`DS:7898h`／`DS:BDFFh` 之後被誰讀。
- `0062:003E` 從地圖座標算出的是什麼。
