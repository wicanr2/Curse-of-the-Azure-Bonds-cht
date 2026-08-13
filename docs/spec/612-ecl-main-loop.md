# 第六百一十二輪：ECL 的主執行迴圈

狀態：`READY`。等級：`exact`。日期：2026-08-14
位置：PC-98 `overlay-02:39A5h`（72 bytes，3 個呼叫者）。

```text
RUN_ECL(pc):
    ECL_PC   := pc
    DS:7896h := 0
    while DS:7896h = 0 and DS:7F34h = 0 do
        DS:A890h := DS:A891h                  ← 上一個 opcode 保留一份
        DS:A891h := script[ECL_PC]            ← 從 bank 3 取新 opcode
        <dispatcher>()                        ← sub_373Eh
    DS:7896h := 0
```

這是整個 ECL 直譯器的**執行入口**。`sub_373Eh` 正是
[spec 560](560-ecl-opcode-dispatch-table.md) 解出的 PC-98 dispatcher。

## 三件先前不知道的事

1. **迴圈有兩個終止條件**：`DS:7896h`（停止旗標，`00h` 設 1）與
   **`DS:7F34h`**。後者就是「全隊都不能行動」那個旗標
   （[spec 596](596-ecl-party-item-sweep.md)、
   [609](609-ecl-area-effect-and-wipeout.md)）——**全滅會直接中止 script 的
   執行迴圈**，不必等 `00h`。
2. **`DS:A890h` 保存上一個 opcode。** dispatcher 讀的是 `DS:A891h`，每次取新
   opcode 前先把舊的抄到 `A890h`。共用 handler 重讀 `A891h` 判斷自己身分
   （[spec 587](587-ecl-handler-21-37-shared.md)）時，`A890h` 提供的是「上一
   條指令是什麼」。
3. **每輪都重新從 `script[ECL_PC]` 取 opcode**，PC 由各 handler 自己推進——
   迴圈本身不加 PC。所以 handler 忘記推進 PC 就是無窮迴圈。

## 誰呼叫它

`3237h`（2 個呼叫者，其中一個是 `38h(9)`，[spec 608](608-ecl-terminate-modes.md)）：

```text
<far 00D0:002F>()
flag := 0
RUN_ECL(DS:7F1Bh)                             ← 第一個進入點
<far 00D0:0025>(@flag)
if flag <> 0 then
    <far 014A:00D9>()
    RUN_ECL(DS:7F1Dh)                         ← 第二個進入點
DS:A66Ch := 1 ; <far 0172:0025>() ; DS:8CF7h := 0
```

## 進入點是一張五格的表

`327Eh`（場景主迴圈）用到另外三個：

```text
repeat
    <初始化畫面>(DS:A2C6h)
    DS:A31Ch := 0 ; DS:A326h := FFh ; DS:7897h := 0
    DS:A2ADh := <由地圖座標算>(DS:A2AAh, DS:A2A9h)
    bank1^[5AAh] := 0
    DS:789Dh := DS:9594h                       ← 記住目前目標
    RUN_ECL(DS:7F1Fh)
    if DS:7897h = 0 then
        bank0^[1E4h] := DS:BDF0h
        <條件成立時 far 0172:0025>()
        DS:7897h := 0
        RUN_ECL(DS:7F17h)
        if DS:7897h = 0 then
            RUN_ECL(DS:7F19h)
            if DS:7897h = 0 then
                DS:9594h := DS:789Dh           ← 還原目標
                <far 014A:002A>(DS:9594h)
until DS:7897h = 0
DS:7F28h := DS:7F27h
```

把兩支加起來，**ECL 的進入點是 `DS:7F17h` 起的五格 word 表**：

| 位址 | 由誰執行 |
|---|---|
| `DS:7F17h` | `327Eh` 第二段 |
| `DS:7F19h` | `327Eh` 第三段 |
| `DS:7F1Bh` | `3237h` 第一段 |
| `DS:7F1Dh` | `3237h` 第二段 |
| `DS:7F1Fh` | `327Eh` 第一段 |

**`DS:7897h` 是「重跑這一輪」的旗標**：非 0 就跳回迴圈開頭重新初始化畫面並
從 `7F1Fh` 再來一次；每段 `RUN_ECL` 之後都檢查它，非 0 就跳過後續各段。

## `3CE4h` 是空函式

原始位元組是 `55 89 e5 89 ec 5d cb`＝
`push bp / mov bp,sp / mov sp,bp / pop bp / retf`——**7 bytes 的空函式**，
沒有呼叫者。

⚠ IDA 把它反組譯成 `… / in al, dx / pop bp / retf`：`89 ec` 的第二個位元組
被單獨當成 `ec`（`in al, dx`）。這種**看起來像 I/O 存取的假指令**，是位元組
邊界錯位的典型症狀。同形的 7-byte 空函式在 `overlay-16:5581h`、
`overlay-18:1756h` 也出現過。

## 明確不宣稱

- `DS:7F1Bh`／`7F1Dh` 兩個進入點分別對應什麼。
- `DS:8CF7h`／`DS:A66Ch` 的語意。
