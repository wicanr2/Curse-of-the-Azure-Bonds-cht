# 第六百一十三輪：ECL 場景的完整生命週期

狀態：`READY`。等級：`exact`。日期：2026-08-14
位置：PC-98 `overlay-02:3A21h`（693 bytes，216 條指令）。
**這是 `INTERPET` 的最後一支未讀函式。**

```text
DS:789Dh := DS:9594h                          ← 記住目前目標
DS:A66Ch := 1 ; DS:BDF6h := 0
DS:789Ah := 0 ; DS:789Bh := 0
DS:7898h := 0 ; DS:7899h := 0
DS:BE00h := 1 ; DS:7F27h := 4 ; DS:7897h := 0
<far 019E:014A>()
if bank0^[1E4h] = 0 then DS:BE00h := 0 ; DS:BDF0h := 1
else                     DS:BDF0h := bank0^[1E4h]
if bank0^[1CCh] = 0 then DS:7F27h := 3
…
RUN_ECL(DS:7F1Fh)
if DS:7F27h <> 3 then DS:A66Ch := 1 ; <far 0172:0025>()
if DS:7897h = 0 then bank0^[1E4h] := DS:BDF0h
else                 <場景主迴圈 327Eh>()

repeat                                        ← 主迴圈，從 3B1Bh 起
    DS:7F2Fh := 0
    if <far 081F:034Eh>(…) then               ← 檔案存在
        <顯示>「データの整理をします リターン・キーを押してください」
        …; DS:8BE9h := 1 ; goto done
    …
    if DS:7F34h = 0 then RUN_ECL(DS:7F17h)
    if DS:7897h <> 0 then <327Eh>() ; goto tail
    if DS:7F34h <> 0 then goto tail
    bank0^[1E0h] := DS:A2A9h                  ← 記下地圖座標
    bank0^[1E2h] := DS:A2AAh
    <far 00C9:0025>()
    if DS:8BE9h <> 0 then goto tail
    <far 0172:0025>()
    if bank0^[1E0h] <> DS:A2A9h then          ← 座標變了
        <far 0893:0000>(DS:484Eh)
        DS:BDF4h := 0 ; DS:BDF5h := 1
        RUN_ECL(DS:7F19h)
        if DS:7897h <> 0 then <327Eh>()
tail:
until DS:7F34h <> 0
DS:7F34h := 0
```

## 三段 `RUN_ECL` 的分工

| 進入點 | 何時執行 |
|---|---|
| `DS:7F1Fh` | 進場一次 |
| `DS:7F17h` | 主迴圈每輪，`DS:7F34h = 0` 才執行 |
| `DS:7F19h` | **只有地圖座標變動時** |

`7F19h` 的觸發條件是**比較 `bank0^[1E0h]` 與當前 `DS:A2A9h`**——進入迴圈前記下
座標，動作之後比對，不同才執行。這是「踏上新格子」的事件。

## 終止

主迴圈由 **`DS:7F34h`** 結束——就是「全隊都不能行動」那個旗標
（[spec 596](596-ecl-party-item-sweep.md)）。離開前把它清零。

所以同一個旗標同時是：執行迴圈的終止條件（[spec 612](612-ecl-main-loop.md)）、
場景生命週期的終止條件、以及 `3Eh`／`38h(3)`／`2Eh` 的讀寫對象。**全隊失去行動
能力會一路把三層迴圈都收掉。**

## `INTERPET` 至此讀完

`pc98/overlay-02` 的 86 個函式：**58 已解讀、28 邊界碎片、0 待解讀**。
52 個 opcode handler、dispatcher、執行迴圈、場景生命週期全部有據可查。
