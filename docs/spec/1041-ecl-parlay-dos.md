# 1041 — ECL `29h`（ENCOUNTER MENU）DOS 側：五個結果碼是遭遇選單的五個選項，不是交涉態度

- 證據等級：`exact`（DOS 側 649 條逐條讀完）
- 作法見 spec 783；PC-98 對側見 spec 611

★ 指令名稱由 spec 1083 定案：`29h` ＝ **`ENCOUNTER MENU`**（`PARLAY` 是另一個指令 `2Ch`）。
本規格原本把 `29h` 叫做 PARLAY，那是錯的；下面的行為描述不受影響。

## `dos overlay-02:02086h`（`retf`）

原本待解讀，PC-98 對側（`overlay-02:0222Ch`）已由 spec 611 讀完。
DOS 這一支 `2086h`..`2784h`。

## ★★★ 更正 spec 611：五個結果碼對到哪五個東西

spec 611 寫「五個結果碼（operand 5..9）正對應
`~HAUGHTY ~SLY ~NICE ~MEEK ~ABUSIVE`」。**DOS 側的一手證據推翻這一點。**

實際餵給選單的選項行是這兩個之一：

```asm
2334  cmp  word ptr es:[di+582h], 0      ; 怪物數
233A  jz   short loc_2348
233C  les  di, ds:4F99h
2340  cmp  word ptr es:[di+1CCh], 0
2346  jnz  short loc_235E
2348  mov  di, offset unk_2029           ; '~COMBAT ~WAIT ~FLEE ~PARLAY'   (27)
      …
235E  mov  di, offset unk_2045           ; '~COMBAT ~WAIT ~FLEE ~ADVANCE'  (28)
```

選單回傳的 `choice` 直接拿去索引 `結果碼[choice]`。所以：

| `choice` | 選項 | 取用的 operand |
|---|---|---|
| 0 | `COMBAT` | operand 5 |
| 1 | `WAIT` | operand 6 |
| 2 | `FLEE` | operand 7 |
| 3 | **`ADVANCE`** | operand 8 |
| 4 | **`PARLAY`**（由 3 改派而來） | operand 9 |

```asm
2397  …同一組條件…
23AF  cmp  [bp+var_1], 3
23B3  jnz  short loc_23B9
23B5  mov  [bp+var_1], 4        ; ★ 顯示 PARLAY 時，選項 3 改派成 4
```

> ★★★ **第四個選項顯示 `ADVANCE` 還是 `PARLAY`，取決於
> `bank1^[582h] <> 0` 且 `bank0^[1CCh] <> 0`**；
> 顯示 `PARLAY` 那一側再把 `choice = 3` 改派成 `4`，
> 所以兩個選項各自對到不同的 operand。
> ⇒ spec 611 說的「選項 3 在條件不成立時被改寫成 4，不是隱藏而是改派」是對的，
> **但改派的理由是「第四格換了一個選項」，不是「同一個選項不可用」。**
>
> ⚠ `'~HAUGHTY ~SLY ~NICE ~MEEK ~ABUSIVE'` 在 `CS:2785h`
> ——**落在本函式結束（`2784h`）之後**，屬於下一支函式。
> spec 611 的第一項證據是相鄰字串的誤認。

★★ 四個選項的 `~` 就是 spec 1039 的**波浪號熱鍵標記**：
`C`OMBAT／`W`AIT／`F`LEE／`P`ARLAY／`A`DVANCE。

## ★★★ 關掉 spec 611 的未定項：`bank1^[582h]` 是什麼

```pascal
bank1^[582h] := <overlay-07 entry#6>(DS:720Fh, DS:7210h, DS:7211h);
if bank1^[580h] < bank1^[582h] then bank1^[582h] := bank1^[580h];
```

`DS:720Fh`／`7210h`／`7211h` 是**隊伍座標與朝向**（spec 1022／1026 同一組）。

> ★★★ **`bank1^[582h]` 是「這一格遭遇到的數量」，由地圖座標算出，
> 再被 operand 2（`bank1^[580h]`）當上限夾住。**
> 結果碼 `3`／`4` 的分支會 `dec bank1^[582h]` 再重畫
> ——**每走一次少一個**，減到 0 才寫回結果。
> ⇒ spec 611 猜的「可以再試一次的迴圈」形狀對，但計數的來源是地圖，不是重試次數。

## ★★★ 關掉 spec 611 的未定項：`t1`／`t2` 跟什麼比

函式一進門就先算兩個量：

```asm
209C  lea  di, [bp+var_40A] ; push
20A2  lea  di, [bp+var_40B] ; push
20A8  call <overlay-07 entry#30>          ; 兩個輸出參數
```

後面的兩處比較：

| 門檻 | 來源 | 比誰 | 用在 |
|---|---|---|---|
| `t1` | `ADDRESSVALUE(0Dh)` ＝ operand 13 | **`var_40B`** | 結果碼 0 ＋ 選 `FLEE` 時 |
| `t2` | `ADDRESSVALUE(0Eh)` ＝ operand 14 | **`var_40A`** | 結果碼 2 ＋ 選 `COMBAT` 時 |

```pascal
{ 結果碼 0、choice = 2(FLEE) }
if var_40B < t1 then STOREVALUE(dest, 1) else STOREVALUE(dest, 2);

{ 結果碼 2、choice = 0(COMBAT) }
if t2 > var_40A then STOREVALUE(dest, 0) …;
```

⇒ **兩個門檻各自對到 `entry#30` 的兩個輸出**，而不是同一個量。

## ★★ 結果碼與選項是二維分派

spec 611 只看到外層。DOS 側清楚是 `結果碼[choice]` 決定外層、`choice` 決定內層：

| 結果碼 | `choice` | 行為 |
|---|---|---|
| 0 | ≠ 2 | `STOREVALUE(dest, 1)` |
| 0 | 2（FLEE） | `var_40B < t1` ? `1` : `2` |
| 1 | 0 | `STOREVALUE(dest, 1)` |
| 1 | 1 | 顯示 `'Both sides wait.'`，重來 |
| 1 | 2 | `STOREVALUE(dest, 2)` |
| 1 | 3 | 數量 ≠ 0 → 減一並重畫；＝ 0 → 顯示 `'Both sides wait.'`；都設重來 |
| 1 | 4 | 數量 ＝ 0 → `STOREVALUE(dest, 3)`；否則減一並重畫、重來 |
| 2 | 0 | `t2 > var_40A` → `STOREVALUE(dest, 0)` ＋ 設 `DS:65A0h := 1`、`DS:65A1h := 11h` |
| … | … | （其餘分支同一形狀） |

## 其他對照

| | DOS | PC-98（spec 611） |
|---|---|---|
| 進門三個旗標 | `DS:8B6Bh := 1`、`DS:8B66h := 0`、`DS:4FC9h := 1` | `DS:BDFDh`／`BDF8h`／`7F36h` |
| operand 1／3 | `DS:7601h`／`DS:7602h` | — |
| 文字槽 | `DS:7648h ＋ i × 100h` | `DS:A8DAh ＋ i × 100h` |
| 樣式旗標 | `DS:8B62h`、`DS:8B63h`、`bank0^[1CCh]`、`DS:728Ah = 50h` 四者 | — |

★ 文字槽與 spec 1039（`2Bh`）**共用同一組緩衝區**，兩平台都是 stride `100h`。
★ `DS:728Ah = 50h` 是 spec 1029 的**圖片子索引**，這裡拿來當樣式條件之一。

## 中文化

| DOS | 長度 | 備註 |
|---|---|---|
| `'~COMBAT ~WAIT ~FLEE ~PARLAY'` | 27 | ★ 波浪號熱鍵，中文要跟 PC-98 的熱鍵表 |
| `'~COMBAT ~WAIT ~FLEE ~ADVANCE'` | 28 | 同上 |
| `'Both sides wait.'` | 16 | PC-98 「お互いに様子を見ている」 |

⚠ 兩行選項行的上限是 `28h` ＝ 40 bytes，中文全形**整行最多 20 字**（含分隔）。

## 明確不宣稱

- 沒有宣稱 `overlay-07 entry#30` 算出來的兩個量是什麼。
- 沒有宣稱 `overlay-07 entry#6(座標, 朝向)` 怎麼算出遭遇數量。
- 沒有宣稱結果碼 `3` 是否存在（本支只看到 `0`／`1`／`2` 三種外層）。
- 沒有宣稱 `DS:65A0h`／`DS:65A1h` 被設成 `1`／`11h` 代表什麼。
- 沒有宣稱 `DS:8B6Bh`／`8B66h`／`4FC9h` 三個旗標的語意。
- 沒有宣稱 `'~HAUGHTY ~SLY ~NICE ~MEEK ~ABUSIVE'` 由哪一支用
  （只確定不是本支）。
