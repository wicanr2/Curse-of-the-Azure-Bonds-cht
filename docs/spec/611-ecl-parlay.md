# 第六百一十一輪：`29h`（ENCOUNTER MENU／遭遇選單）

> ★ **指令名稱更正**：`29h` ＝ **`ENCOUNTER MENU`**，不是 `PARLAY`
> （`PARLAY` 是另一個指令 `2Ch`）。原作自己的助憶碼表見 spec 1083；
> 五個結果碼的對應由 spec 1041 更正。

狀態：`READY`。等級：`exact`。日期：2026-08-14
位置：PC-98 `overlay-02:222Ch`（1,812 bytes，669 條指令）。
**這是 `INTERPET` 的最後一支未讀 handler。**

```text
saved := DS:A2A8h
DS:BDFDh := 1 ; DS:BDF8h := 0 ; DS:7F36h := 1
READVAR(0Eh)                                  ← 14 個 operand，全模組最多
DS:A893h     := ADDRESSVALUE(1)
bank1^[580h] := ADDRESSVALUE(2)
DS:A894h     := ADDRESSVALUE(3)
dest := ADDFNC(high[4], low[4])
for i := 0 to 4 do outcome[i] := ADDRESSVALUE(i + 5)     ← operand 5..9：五個結果碼
for i := 1 to 3 do <解 packed text> → DS:[A8DAh + i*100h] ← operand 10..12：三行文字
t1 := ADDRESSVALUE(0Dh) ; t2 := ADDRESSVALUE(0Eh)        ← operand 13、14：兩個門檻
bank1^[582h] := min(<由地圖座標算出>, bank1^[580h])
repeat
    <把三行文字循環捲動顯示>
    choice := <選單>()
    if bank1^[582h] = 0 or bank0^[1CCh] = 0 then
        if choice = 3 then choice := 4        ← 該選項不可用時改走 4
    case outcome[choice] of
        0: STOREVALUE(dest, 1) 或 (dest, 2)   ← 依 t1／t2 與計數比較
        1: <顯示訊息>
        2: STOREVALUE(dest, 2)
        4: if bank1^[582h] = 0 then STOREVALUE(dest, 3)
           else bank1^[582h] := bank1^[582h] − 1 ; 重新顯示
until <不再重來>
DS:7F36h := 0 ; DS:BDFDh := 0 ; DS:A2A8h := saved
```

## 這是 Gold Box 的交涉指令

兩個關鍵證據：

1. **五個結果碼**（operand 5..9）對應遭遇選單的五個選項
   `~COMBAT ~WAIT ~FLEE ~PARLAY` ／ `~COMBAT ~WAIT ~FLEE ~ADVANCE`
   （DOS `overlay-02:2029h`／`2045h`）。
   ⚠ **這一條在 spec 1041 被更正**：先前寫的
   `~HAUGHTY ~SLY ~NICE ~MEEK ~ABUSIVE`（`overlay-02:2785h`）
   落在本函式結束之後，屬於下一支函式，是相鄰字串的誤認。
2. 分支裡的訊息是「**相手は逃げた**」（對方逃走了，`unk_221Fh`）與
   「**お互いに様子を見ている**」（雙方互相觀望，`unk_2208h`）。

所以 ECL script 用一條 `29h` 就描述完整場交涉：三行提示文字、五種態度各自的
結果碼、兩個門檻值、結果寫回哪裡。

## 幾個實作細節

- **14 個 operand 是全模組最多的**（次多是 `2Bh` 的變長）。
- **選項 3 在條件不成立時被改寫成 4**（`bank1^[582h] = 0` 或
  `bank0^[1CCh] = 0`），不是隱藏而是改派——玩家仍看得到那個選項。
- 結果碼 `4` 會**遞減 `bank1^[582h]` 並重新顯示**，形成「可以再試一次」的迴圈；
  次數用完（`= 0`）才寫回結果 `3`。
- 三行文字用 `DS:A8DAh` 起、stride `100h` 的槽位，與 `2Bh`
  （[spec 605](605-ecl-horizontal-menu.md)）**共用同一組緩衝區**。
- `DS:A2A8h` 進來存、離開還原——與 `0Ch`／`0Dh`
  （[spec 599](599-ecl-select-member-and-0c.md)）同一個習慣。

## 明確不宣稱

- 五個結果碼各自的完整語意（只讀出 `0`／`1`／`2`／`4` 四種分支）。
  ★ spec 1041 補上：外層是**結果碼**、內層是**選項**的二維分派。
- ~~`t1`／`t2` 兩個門檻與什麼比較~~ —— **spec 1041 關掉**：
  `t1` 比 `overlay-07 entry#30` 的第二個輸出、`t2` 比第一個輸出。
- ~~`bank1^[582h]` 在這裡代表的次數是什麼~~ —— **spec 1041 關掉**：
  是 `<overlay-07 entry#6>(隊伍座標, 朝向)` 算出的**遭遇數量**，
  再被 operand 2 當上限夾住；結果碼 `3`／`4` 每次減一。
