# 第六百五十七輪：DOS 配置、兩 byte 裝置命令，與一段字串轉換碎片

狀態：`READY`。等級：`exact`。日期：2026-08-14
位置：PC-98 `PC98-GAME.EXE` 的 `188E3h`、`18A44h`、`18A8Eh`、`18D3Ah`。

## `188E3h`：向 DOS 要記憶體

```text
es:[di]   := 0
bx := 要的段落數
int 21h (AH = 48h)
if 進位 then ax := 0                    ← 失敗就記 0，不記錯誤碼
es:[di+2] := ax
or al, ah                                ← 設旗標供呼叫端判斷
```

失敗時 **DOS 回傳的錯誤碼被丟掉**（`int 21h` 失敗時 `AX` 是錯誤碼，這裡直接覆寫成
0）。呼叫端只知道成功與否。

回傳約定是靠 `or al, ah` 設 `ZF`——`AX = 0` 時 `ZF` 置起。

## `18A44h` 與 `18A8Eh`：兩 byte 命令緩衝

```text
18A44h(v):
    if byte_24E73h = 1 then return                  ← 總開關
    v := v − 1
    if byte_24E71h = v then return                  ← 值沒變就不做
    byte_24E71h := v
    <18A8Eh>()                                       ← 先送「1」
    <sub_19259>(320h)
    byte_241C9h := 0
    byte_241C8h := v
    <sub_18BDB>(DS:@byte_241C8h, 7Eh)                ← 再送「0 ＋ v」

18A8Eh():
    byte_241C9h := 1
    <sub_18BDB>(DS:@byte_241C8h, 7Eh)
```

`byte_241C8h`／`byte_241C9h` 是**相鄰的兩個 byte**，整組當一個 2-byte 命令送給
`sub_18BDB`，第二個參數固定 `7Eh`。

`18A8Eh` 只設 `byte_241C9h := 1` 就送出——**不動 `byte_241C8h`**，所以送出去的是
「上一次的值 ＋ 1」。`18A44h` 則是先呼叫它（等於送出「舊值 ＋ 1」），中間插一個
`sub_19259(320h)`，最後才送「新值 ＋ 0」。

這個「先送 1 再送 0」的成對寫法，加上中間的延遲呼叫，是**對硬體送命令再等它生效**
的典型形狀。`7Eh` 多半是命令或埠編號，本輪不宣稱。

`byte_24E73h = 1` 是總開關：整支直接跳過。

## `18D3Ah`：不是函式，是碎片

```asm
18D3A  jb   short loc_18D59        ← 以條件跳躍開頭，沒有 prologue
```

前一支函式在 `18C94h` 就以 `retf 10h` 結束，中間 `18C97h..18D39h` 沒有任何被認出的
函式。所以 `18D3Ah` 是某支未被 IDA 認出的函式的**中段**，標為邊界碎片。

它的內容是 **ASCIIZ 轉 Pascal 短字串**（就地右移一格再補長度）：

```asm
add   di, 1Eh
xor   al, al
mov   cx, 100h
repne scasb                  ; 找結尾 0
not   cl                     ; cl = 長度
mov   al, cl
dec   di
mov   si, di
dec   si
std
rep   movsb                  ; 由後往前搬一格
stosb                        ; 補上長度 byte
```

`cx` 上限 `100h`——**字串超過 256 bytes 沒有結尾 0 時，`not cl` 會得到 0**，長度變
成 0。判讀記在這裡，但這段的完整入口與呼叫端不在本輪範圍。

## 明確不宣稱

- `sub_18BDB`／`sub_19259` 的行為與 `7Eh`、`320h` 的意義。
- `byte_24E71h`／`byte_24E73h` 的用途。
- `18D3Ah` 所屬函式的起點。
