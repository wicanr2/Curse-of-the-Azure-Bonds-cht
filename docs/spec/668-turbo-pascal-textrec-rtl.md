# 第六百六十八輪：Turbo Pascal 的 text file RTL —— `TextRec` 佈局確認

狀態：`READY`。等級：`exact`。日期：2026-08-14
位置：PC-98 `PC98-GAME.EXE` 的 `1B9C7h`、`1BA41h`、`1BA90h`。

## `1B9C7h`：`Assign`

逐欄位建立記錄，順序就是佈局：

```asm
xor ax, ax ; stosw          ; +00h  Handle   := 0
mov ax, 0D7B0h ; stosw      ; +02h  Mode     := fmClosed
mov ax, 80h ; stosw         ; +04h  BufSize  := 128
xor ax, ax ; stosw          ; +06h  Private  := 0
stosw                       ; +08h  BufPos   := 0
stosw                       ; +0Ah  BufEnd   := 0
lea ax, [di+74h] ; stosw    ; +0Ch  BufPtr（offset）
mov ax, es ; stosw          ; +0Eh  BufPtr（segment）
mov ax, 148Ch ; stosw       ; +10h  OpenFunc（offset）
mov ax, cs ; stosw          ; +12h  OpenFunc（segment）
xor ax, ax ; mov cx, 0Eh ; rep stosw   ; +14h..+2Fh 清 0
lodsb ; cmp al, 4Fh ; ja → al := 4Fh   ; 名字截到 79 字
mov cl, al ; rep movsb      ; +30h  Name
xor al, al ; stosb          ;       NUL 結尾
```

寫完六個 word 時 `di` 在 `+0Ch`，所以 `lea ax, [di+74h]` ＝ 基底 `+80h`——
**緩衝區在 `+80h`**。名字上限 `4Fh` ＝ 79 字 ＋ NUL ＝ 80 bytes，剛好填滿
`+30h..+7Fh`。

這與 Turbo Pascal 的 `TextRec` **逐欄位吻合**：

| 偏移 | 欄位 |
|---|---|
| `+00h` | `Handle` |
| `+02h` | `Mode` |
| `+04h` | `BufSize` |
| `+06h` | `Private` |
| `+08h` | `BufPos` |
| `+0Ah` | `BufEnd` |
| `+0Ch` | `BufPtr` |
| `+10h` | `OpenFunc` |
| `+14h` | `InOutFunc` |
| `+18h` | `FlushFunc` |
| `+1Ch` | `CloseFunc` |
| `+20h` | `UserData`（16 bytes） |
| `+30h` | `Name`（80 bytes） |
| `+80h` | `Buffer` |

[spec 659](659-console-raw-mode-and-textrec.md) 由常數位置推的 `TextRec` 判斷，
**這裡由完整的建構順序確認**，等級由 `strong inference` 升為 `exact`。

`Mode` 的四個值也全部出現：`0D7B0h` ＝ `fmClosed`、`0D7B1h` ＝ `fmInput`、
`0D7B2h` ＝ `fmOutput`、`0D7B3h` ＝ `fmInOut`。

## `1BA41h`：`Reset`／`Rewrite` 的共同部分

```text
dx := 0D7B3h                              ← 目標模式 fmInOut
mode := es:[di+2]
if mode = fmInput 或 fmOutput then
    <1BA90h>()                             ← 先關掉
else if mode <> fmClosed then
    word_23B08h := 66h ; return            ← 錯誤 102
es:[di+2] := dx                            ← Mode := fmInOut
es:[di+8] := 0                             ← BufPos
es:[di+0Ah] := 0                           ← BufEnd
<sub_1BACBh>(bx = 10h)                     ← 呼叫 OpenFunc
if 失敗 then es:[di+2] := fmClosed
```

`bx` 是**欄位偏移**——`10h` 就是 `OpenFunc`。`sub_1BACBh` 是「照 `bx` 指的偏移
呼叫那個函式指標」的通用跳板。

## `1BA90h`：`Close`

```text
al := 1
if mode = fmInput then 跳過沖洗
else if mode = fmOutput then
    <sub_1BACBh>(bx = 14h)                 ← InOutFunc，把緩衝寫出去
else
    word_23B08h := 67h ; return            ← 錯誤 103
<sub_1BACBh>(bx = 1Ch)                     ← CloseFunc
es:[di+2] := fmClosed
```

兩個錯誤碼對得上 Turbo Pascal 的 I/O 錯誤：**`66h` ＝ 102、`67h` ＝ 103**，
分別是「檔案未指派」與「檔案未開啟」。`word_23B08h` 就是 `InOutRes`。

## 明確不宣稱

- `sub_1BACBh` 怎麼由 `bx` 取出並呼叫函式指標（形狀已明，內部未讀）。
- `OpenFunc` 預設值 `CS:148Ch` 指的那支做什麼。
