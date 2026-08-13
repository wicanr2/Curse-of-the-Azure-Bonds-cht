# 第六百零七輪：`28h`（依百分比縮放）

狀態：`READY`。等級：`exact`。日期：2026-08-14
位置：PC-98 `overlay-02:2125h`（227 bytes）。

```text
READVAR(3)
all := ADDRESSVALUE(1)                    ← 0 ＝ 只對目前目標，非 0 ＝ 全隊
pct := ADDRESSVALUE(2)
r   := Real(100 − pct) / 100.0            ← 6-byte Real 除法
v   := ADDRESSVALUE(3)
if all = 0 then
    <far 0062:00A7>(DS:9594h, r) ; <far 0062:00AC>(DS:9594h, v)
else
    for each node in DS:9598h 鏈 do
        <far 0062:00A7>(node, r) ; <far 0062:00AC>(node, v)
```

## 那個 Real 常數是 100.0

程式裡沒有 `100.0` 這個立即值，除數是用三個暫存器組起來的：

```text
mov cx, 87h ; xor si, si ; mov di, 4800h
call far 0A65h:0CB0h                      ← RTL 的 Real 除法
```

Turbo Pascal 的 6-byte `Real`：`value = 2^(exp−129) × 1.f`。這裡
`exp = 87h = 135` ⇒ `2^6 = 64`，尾數 `4800h` ⇒ `f = 4800h / 8000h = 0.5625`，
所以 `64 × 1.5625 = 100.0`。

**這是浮點運算，不是整數。** `(100 − pct) / 100` 用整數算的話，`pct = 30` 會
得到 0（`70/100`），行為完全不同。

## 又一個取出卻沒用到的變數

全隊分支裡有：

```text
item := node^[14Eh]                       ← 存進區域變數
```

之後**完全沒有被讀**。這是本專案在 `INTERPET` 裡找到的第二個死變數，
第一個是 `22h` 的第二個輸出（[spec 597](597-ecl-opcode-22-dead-output.md)）。
兩者都無害（不改變行為），但說明這份原始碼有未清理的痕跡。

## 明確不宣稱

- `0062:00A7`（吃 Real）與 `0062:00AC`（吃 byte）對角色做了什麼。
- `all` 非 0 時**沒有**排除任何角色——不看狀態、不看 NPC 標記，整條鏈都套用。
