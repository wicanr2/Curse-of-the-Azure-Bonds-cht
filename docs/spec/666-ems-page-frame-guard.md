# 第六百六十六輪：EMS page frame 的偵測與保存還原

狀態：`READY`。日期：2026-08-14
位置：PC-98 `PC98-GAME.EXE` 的 `1A5F1h`、`1A617h`、`1A634h`、`1A9E1h`、`1ABE0h`。

## 三支一組：偵測、關閉、還原

```text
1A5F1h（偵測）:
    int 67h (AX = 4100h)                     ← 取 page frame 段位址
    if AH <> 0 then return                    ← EMS 不在
    if BX <> 0B000h then return               ← ★ 只認 B000h
    int 67h (AX = 7000h)
    if AH <> 0 then return
    cs:byte_1A64Bh := 1                       ← 記在程式碼段

1A617h（關閉）:
    if cs:byte_1A64Bh <> 1 then return
    int 67h (AX = 7000h) ; cs:byte_1A64Ch := AL    ← 先存目前狀態
    int 67h (AX = 7001h, BL = 0)                   ← 設成 0

1A634h（還原）:
    if cs:byte_1A64Bh <> 1 then return
    int 67h (AX = 7001h, BL = cs:byte_1A64Ch)      ← 寫回存下來的值
```

`7000h`／`7001h` 不是標準 EMS 功能碼（標準到 `5Xh` 為止），是**驅動程式的擴充**。
`1A617h` 與 `1A634h` 是明確的「先讀存起來、事後寫回」配對。

### 為什麼只認 `B000h`

PC-98 的 `B000h` 是**圖形 VRAM 的一個平面**。EMS page frame 落在那裡就會與繪圖
撞位址——所以這一組的存在本身，說明它是**在用到那塊記憶體之前先把 page frame 讓
開、用完再還原**。

判斷寫死成「等於 `B000h`」而不是「與 VRAM 範圍重疊」，所以 page frame 在別的位址
時整組功能不啟用（`byte_1A64Bh` 保持 0），三支都直接返回。

旗標與存下來的值**都放在程式碼段**（`cs:byte_1A64Bh`／`cs:byte_1A64Ch`），與
[spec 659](659-console-raw-mode-and-textrec.md) 的 `CON` handle 同一種做法。

## `1A9E1h`：以 16 為底的兩段式加減

```text
<sub_1AC19h>()
ax := si − word_23AECh
dx := di − word_23AEEh
if 借位 then { ax := ax + 10h ; dx := dx − 1 }
es:di := dword_23AF0h
while di <> 0 do
    dx := dx + es:[di+6] − es:[di+2]
    ax := ax + es:[di+4]
    if ax >= 10h then { ax := ax − 10h ; dx := dx + 1 }
    ax := ax − es:[di]
    if 借位 then { ax := ax + 10h ; dx := dx − 1 }
    di := di + 8
<sub_1AC70h>()
```

`ax` 被維持在 `0..0Fh`、`dx` 每滿 16 進一——是「段落 ＋ 段內偏移」的兩段式表示
（一個段落 16 bytes）。**每次加減都手動處理進借位**，沒有用 32-bit 運算。

走訪的是 `dword_23AF0h` 指的陣列，每筆 **8 bytes**，用到 `+0`／`+2`／`+4`／`+6`
四個 word。

## `1ABE0h`：從陣列尾端退一筆

```text
di := low(dword_23AF0h) − 8
if di = 0 then 設進位返回                     ← 已經空了
si := (di shr 4) + high(dword_23AF0h)
if si <= word_23AEEh then 設進位返回          ← 退過頭
low(dword_23AF0h) := di
```

`di shr 4` 把 offset 換算成段落數再加上 segment——與 `1A9E1h` 的兩段式表示一致。

**進位旗標表示失敗**（兩個 `stc` 出口），成功則是落到 `retn` 時的旗標狀態。

## 明確不宣稱

- EMS 功能碼 `7000h`／`7001h` 的確切語意（只知道是一讀一寫的一對）。
- `sub_1AC19h`／`sub_1AC70h` 的行為。
- `dword_23AF0h` 陣列每筆 8 bytes 的欄位意義。
