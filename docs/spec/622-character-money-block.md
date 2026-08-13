# 第六百二十二輪：角色紀錄的錢欄位 —— `+FBh` 硬幣陣列與 `ROBDOUGH`

狀態：`READY`。日期：2026-08-14
位置：PC-98 `overlay-07:1DD1h`（`ROBDOUGH`，303 bytes）、
`overlay-21:0148h`（85 bytes）。

## 欄位配置

角色紀錄（PC-98 `1A7h` bytes、DOS `1A6h`，見 [spec 626](626-goduel-and-charrec-size.md)）
的 `+FBh` 起是**七個 word**：

| offset | 內容 | 證據等級 |
|---|---|---|
| `+FBh` `+FDh` `+FFh` `+101h` `+103h` | **五種硬幣**（陣列） | `exact` |
| `+105h` `+107h` | 另外兩項（寶石／珠寶） | `strong inference` |

前五個是**陣列不是五個獨立欄位**——`overlay-21:0148h` 用
`i := 0..4` 迴圈存取 `[di + 0FBh + 2i]` 歸零，索引算式是
`cbw` ＋ `shl ax,1`。後兩個不在迴圈內。

## `+101h` 與 `+103h` 的身分：5 gp = 1 pp

```asm
mov  ax, [bp+arg_0]      ; 金額
xor  dx, dx
mov  cx, 5
div  cx
mov  es:[di+103h], ax    ; 商 → +103h
...                      ; 重算一次
xchg ax, dx
mov  es:[di+101h], ax    ; 餘 → +101h
```

金額除以 5，**商進 `+103h`、餘留 `+101h`**。AD&D 的兌換率是
**1 platinum = 5 gold**，所以 `+103h` 是白金、`+101h` 是金幣，陣列由低價值排到
高價值：`+FBh` 銅、`+FDh` 銀、`+FFh` 琥珀金、`+101h` 金、`+103h` 白金。

這條推論的強度來自「除數正好是 5、而且商放在餘數的**後一格**」——反過來排
（白金在前）會得到 `+103h` 是銀、`+101h` 是琥珀金，兌換率對不上。

## `ROBDOUGH(char, factor)`

七個欄位**各自**做同一件事：

```asm
mov  ax, es:[di+<欄位>]
xor  dx, dx                  ; 無號 word → 32-bit
call far 0A65h:0CBEh         ; longint → real
mov  cx, [bp+arg_4]          ; 6-byte Turbo Pascal real
mov  si, [bp+arg_6]
mov  di, [bp+arg_8]
call far 0A65h:0CAAh         ; real 乘法
call far 0A65h:0CC2h         ; real → 整數
mov  es:[di+<欄位>], ax
```

`factor` 是一個 **6-byte Turbo Pascal real**，用三個 word 參數
（`arg_4`/`arg_6`/`arg_8`）傳入。整支函式沒有分支——**七個欄位無條件全乘**，
寶石珠寶也一起。

`xor dx, dx` 表示轉換前把 word 當**無號**看待。

## 明確不宣稱

- `0A65h:0CC2h` 是四捨五入還是無條件捨去（兩者都是 Turbo Pascal RTL 的標準
  routine，這裡沒有區分的證據）。
- `factor` 由呼叫端怎麼算、是「被偷走的比例」還是「剩下的比例」。
- `+105h`／`+107h` 哪個是寶石哪個是珠寶。
- `overlay-21:0148h` 的 `DS:9594h` 指向哪一個角色紀錄。
