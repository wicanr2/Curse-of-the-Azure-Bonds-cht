# 第六百六十五輪：用 EMS 放 overlay，以及重疊安全的搬移

狀態：`READY`。日期：2026-08-14
位置：PC-98 `PC98-GAME.EXE` 的 `1A49Bh`、`1A544h`、`1A34Fh`、`1A397h`、`1A3E1h`。

## `int 67h` ＝ EMS

```text
1A49Bh:
    int 67h (AH = 41h)                      ← 取 page frame 段位址
    word_28108h := BX
    ax := 3FFFh ; dx := 0
    重複:                                    ← 走 overlay 描述子鏈
        bx := word_23AD4h + word_23B04h + 10h
        es := bx
        ax := ax + es:[8] ; dx := dx + 進位   ← 32-bit 累加所有大小
        bx := es:[0Eh]
    until bx = 0
    bx := 4000h
    ax := (dx:ax) div bx                     ← 換算成分頁數
    int 67h (AH = 43h, BX = 分頁數)          ← 配置 EMS
    if 成功 then word_23AE6h := DX            ← 保存 handle
```

`4000h` 段落 ＝ **16 KB**，正是 EMS 的分頁大小。累加時先塞 `3FFFh` 再除——這是
**無條件進位**的寫法（`(總和 + 分頁大小 − 1) div 分頁大小`）。

`int 67h AH=43h` 失敗的判斷是 `shl ah, 1` ＋ `jb`：EMS 的狀態碼放在 `AH`，
**左移一位把 bit 7 送進進位旗標**——`AH >= 80h` 就是錯誤。

`1A544h` 用 `AX = 4400h`（對映實體分頁 0）＋ `DX = word_23AE6h`（handle）把
overlay 換進 page frame，同樣用 `shl ah, 1` ＋ `jb` 判斷失敗。

所以 **overlay 不只放在常規記憶體，有 EMS 時會放進去**。remake 不需要這一層，但
要知道原版的載入路徑有兩條。

## `1A34Fh`：搬移 overlay 並修正 stub

```text
ax := word_23ADEh                            ← 新位置
dx := es:[10h]                                ← 舊位置
es:[10h] := ax
cx := (es:[8] + 1) shr 1                      ← 以 word 為單位的長度
if ax >= dx then                              ← 目的在來源之後
    si := (cx − 1) × 2 ; std                  ← 由後往前搬
else
    si := 0 ; cld                             ← 由前往後搬
di := si
DS := dx ; ES := ax
rep movsw

if es:[20h] <> 0CDh then                      ← stub 已改成 far jmp
    <sub_1A3C3h>()
    cx := es:[0Ch]
    di := 23h                                  ← 每筆 stub 的第 4 個 byte
    重複 cx 次:
        stosw                                  ← 只改 segment 那個 word
        di := di + 3                           ← 5 bytes 一筆
```

**方向依「目的在來源之前還是之後」選**——重疊時不會覆蓋掉還沒搬的部分。這是
`movsw` 搬重疊區的標準做法。

修正 stub 時只動 `di = 23h` 起、每 5 bytes 一筆的那個 word。對照
[spec 664](664-overlay-manager-stub-patching.md) 的 `EA off seg`：`+20h` 是 `EA`、
`+21h..22h` 是 offset、**`+23h..24h` 是 segment**——所以這裡只更新 segment，
offset 不變（overlay 內部位移不因搬移而改變）。

`es:[20h]` 是 `0CDh`（還是 stub）時整段跳過，因為那時沒有 segment 要修。

## `1A397h`：接到鏈尾

```text
<sub_1A3FCh>()
word_23ADEh := word_23ADEh + 回傳值
bx := 7852h                                   ← 鏈頭的位址（絕對）
重複:
    ax := [bx]
    if ax = 0 then break
    DS := ax ; bx := 14h                      ← 換到那個節點的 +14h 繼續
[bx] := ES
es:[14h] := 0
```

走訪時**把 `DS` 換成節點的 segment、`BX` 固定成 `14h`**，所以同一段程式碼既能讀
鏈頭（絕對位址 `7852h`）也能讀節點欄位。

## `1A3E1h`：算剩餘空間

```text
if word_23AE2h <> 0 then
    ax := word_23AE2h^[10h] − word_23ADEh
    if 沒有借位 then return ax
ax := word_23AE0h − word_23ADEh
```

先試「第一個 overlay 的位置減去目前配置點」，借位（會變負）才退回用
`word_23AE0h`。

## 明確不宣稱

- `sub_1A3C3h`／`sub_1A3FCh` 的行為。
- `word_23AD4h`／`word_23B04h`／`word_23AE0h` 的角色。
- EMS 分頁與常規記憶體之間如何選擇。
