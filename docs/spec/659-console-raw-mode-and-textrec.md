# 第六百五十九輪：把 `CON` 切成 raw 模式、鍵盤讀取，與 TextRec 初始化

狀態：`READY`。日期：2026-08-14
位置：PC-98 `PC98-GAME.EXE` 的 `19386h`、`192A0h`、`192D2h`、`19259h`。

## `19386h`：開 `CON` 並關掉 ASCII 模式

```text
DS := CS
DX := 4E9h                       ← 檔名字串位址
int 21h (AX = 3D00h)             ← 開檔（唯讀）
cs:word_193CDh := AX             ← 存 handle（存在程式碼段裡）
BX := AX
int 21h (AX = 4400h)             ← IOCTL 取裝置資訊
if (DL and 20h) = 0 then
    DH := 0 ; DL := DL or 20h
    int 21h (AX = 4401h)         ← IOCTL 設回去
```

`CS:4E9h` 的內容是 **`"CON"`**（ASCIIZ）——`seg050` 的基底是 `18EE0h`，
`18EE0h + 4EDh = 193CDh` 正好等於 IDA 標的 `word_193CD`，兩邊互相印證。

`int 21h` IOCTL 的 `DL` bit 5 是**二進位（raw）模式**。這支檢查它，沒設就設上——
之後從 `CON` 讀進來的資料不再被 DOS 做 ASCII 加工（不攔 Ctrl-C、不做行緩衝）。

handle 存在**程式碼段**的 `word_193CDh`，不是資料段。

## `192A0h`：鍵盤讀取，特殊鍵分兩次回傳

```text
if byte_280E4h <> 0 then
    al := byte_280E4h ; byte_280E4h := 0 ; return al     ← 上一次留下的掃描碼
重來:
    al := 0 ; <sub_19085>()
    if al <> 0 then return                                ← 有別的來源
    int 18h (AH = 1)                                      ← PC-98 鍵盤 BIOS：有沒有按鍵
    if BH = 0 then goto 重來
    int 18h (AH = 0)                                      ← 取按鍵
    if AL <> 0 then return AL                             ← 一般字元
    byte_280E4h := AH                                     ← AL = 0 表示特殊鍵
    return
```

`AL = 0` 代表特殊鍵（方向鍵、功能鍵），此時把 `AH`（掃描碼）**存進 `byte_280E4h`
留給下一次呼叫**。所以特殊鍵要讀兩次才拿得到掃描碼——第一次回 0、第二次回掃描碼。

這是 DOS／PC-98 的標準做法，但 remake 若把輸入改成事件式，**這個「兩次」的協定會
影響呼叫端**，不能只換底層。

等待迴圈中間夾 `sub_19085`，沒有睡眠或讓出。

## `192D2h`：TextRec 初始化

```text
[di+2]  := 0D7B0h                ← Turbo Pascal 的 fmClosed
[di+4]  := 80h                   ← 緩衝大小 128
[di+0Ch] := @[di+80h]            ← 緩衝指標（offset）
[di+0Eh] := DS                   ← 緩衝指標（segment）
[di+10h] := offset loc_19329     ← 函式指標（offset）
[di+12h] := CS                   ← 函式指標（segment）
[di+30h] := 0
```

`0D7B0h` 是 Turbo Pascal `TextRec.Mode` 的 `fmClosed`，`+4` 是 `BufSize`、
`+0Ch` 是 `BufPtr`、`+10h` 是 `OpenFunc`——欄位位置與 Turbo Pascal 的 `TextRec`
一致。等級 `strong inference`（沒有型別表可對，是由常數與位置推的）。

後半是**一次性初始化**（`byte_28102h` 當旗標）：

```text
dx := word_280F9h
word_280D8h := ((high(dx) × byte_280DAh) + low(dx)) × 2
if byte_280DAh <> 50h then word_280D8h := word_280D8h × 2
```

`50h` ＝ 80，是文字模式的欄數（[spec 645](645-pc98-text-layer-primitives.md)）。
所以 `byte_280DAh` 是每列欄數，**不是 80 欄時再乘一次 2**——多半是 40 欄模式下每格
佔兩倍寬。

## `19259h`

```text
ax := 參數
if ax = 0 then return
ax := ax div word_280D2h          ← 無號除法
<sub_19271>()
```

參數為 0 直接返回，不會除。

## 明確不宣稱

- `sub_19085`／`sub_19271`／`loc_19329` 的行為。
- `word_280D2h`／`word_280F9h`／`byte_280DAh` 的確切語意。
- `[di+30h]` 對應 `TextRec` 的哪個欄位。
