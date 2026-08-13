# 第六百七十輪：檔案關閉、定位與大小，以及標準 handle 的保護

狀態：`READY`。等級：`exact`。日期：2026-08-14
位置：PC-98 `PC98-GAME.EXE` 的 `1BF69h`、`1C03Bh`、`1C0B2h`、`1BD32h`。

## `1BF69h`：`Close`，但標準 handle 不關

```text
<sub_1BF8Eh>()                            ← 檢查（回傳旗標）
if 不通過 then return
bx := es:[di]                              ← Handle
if bx > 4 then
    int 21h (AH = 3Eh)                     ← 關閉
    if 失敗 then word_23B08h := AX
es:[di+2] := 0D7B0h                        ← Mode := fmClosed
```

**`Handle ≤ 4` 直接跳過關閉動作**——DOS 的 `0`..`4` 是標準 handle（stdin、stdout、
stderr、aux、prn）。關掉它們會影響整個行程，所以這裡擋住。

但**不論有沒有真的關，`Mode` 都會被設成 `fmClosed`**。

## `1C03Bh`：`Seek`

```text
cx := high(記錄編號) × es:[di+4]           ← RecSize
(dx:ax) := low(記錄編號) × es:[di+4]
cx := cx + dx
dx := ax
int 21h (AX = 4200h)                       ← 由檔首定位
if 失敗 then word_23B08h := AX
```

**兩次 16×16 乘法拼成 32-bit**：低位那次取整個 `dx:ax`，高位那次只取 `ax` 並加到
`cx`。高位乘法的 `dx`（超過 32-bit 的部分）**被丟掉**——記錄編號大到讓
`編號 × RecSize` 超過 4 GB 時會靜默回捲。

`es:[di+4]` 是 `FileRec.RecSize`（`TextRec` 的同一個位置是 `BufSize`，
[spec 668](668-turbo-pascal-textrec-rtl.md)）——兩種記錄共用前面的欄位。

## `1C0B2h`：一次取得位置與大小

```text
if Mode <> 0D7B3h（fmInOut）then
    word_23B08h := 67h ; ax := dx := 0 ; 設進位 ; return
int 21h (AX = 4201h, 位移 0)               ← 目前位置 → dx:ax
暫存 dx, ax
int 21h (AX = 4202h, 位移 0)               ← 檔案大小 → dx:ax
取回剛才的位置到 bx（低）、cx（高）
暫存 dx, ax（大小）
int 21h (AX = 4200h, cx:dx = 原位置)       ← 定位回去
取回大小到 cx（低）、bx（高）
```

「移到檔尾量大小、再移回原處」是 DOS 沒有直接取檔案大小時的標準做法。

**回傳分兩組**：`dx:ax` 是最後一次 `4200h` 的結果（＝還原後的位置），
`bx:cx` 是檔案大小。同一支函式用兩組暫存器帶回兩個值。

錯誤時**明確設進位**並把 `dx:ax` 清 0——與 `1BF69h`／`1C03Bh` 只記 `InOutRes`
不同，這支有進位旗標可查。

## `1BD32h`：`Flush`

```text
if es:[di+1Ah] = 0 then return             ← FlushFunc 的 segment 為 0 ⇒ 指標是 nil
if word_23B08h <> 0 then return             ← 已經有錯就不做
call dword ptr es:[di+18h]                  ← FlushFunc
if AX <> 0 then word_23B08h := AX
```

判斷「有沒有裝 `FlushFunc`」是**檢查指標的 segment 部分**（`+1Ah`）而不是整個
dword——offset 為 0 但 segment 非 0 仍算有效。

## 明確不宣稱

- `sub_1BF8Eh` 的檢查內容。
- `1C0B2h` 的呼叫端怎麼區分它回傳的兩組值。
