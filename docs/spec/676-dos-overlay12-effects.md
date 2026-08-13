# 第六百七十六輪：DOS `overlay-12` 三支 —— 英文原文與種族加成

狀態：`READY`。等級：`exact`。日期：2026-08-14
位置：DOS `overlay-12` 的 `00E8h`、`0188h`、`03E5h`。

## 欄位偏移對照再驗證一次

`00E8h` 讀 `arg_6^[19Ch]` 與 `arg_6^[19Dh]`；PC-98 的對應那支
（[spec 638](638-overlay12-effect-batch2.md) 的 `01E6h`）讀的是 `+19Bh` 與 `+19Ch`。
**差 1，與 [spec 641](641-dos-field-offset-shift.md) 的對照表一致**（`+14Ch` 以後
DOS ＝ PC-98 − 1）。

同樣地 `03E5h` 的 `+1A4h` ＝ PC-98 的 `+1A5h`（目前 HP），`0188h` 的 `+11Ah`
（`RACETYPE`）兩平台相同。三支都對得上，對照表在 `overlay-12` 這邊再獲一次驗證。

## `00E8h`：扣量不足就改走別的路

```text
total := arg_6^[19Dh] + arg_6^[19Ch]
if arg_2^[3] > total then
    arg_2^[3] := arg_2^[3] − total            ← 夠扣就直接扣
else
    <far 013Eh:043Fh>(arg_6, 3, 0)             ← 不夠扣改呼叫別的
備妥 'is fighting with snakes'
<far 014Ah:0634h> → <sub_1572> → <far 0158h:0559h>
<far 0159h:059Ah>(arg_6)
```

兩個欄位**相加**當成一個總量。判斷是嚴格大於（`jbe` 走 else），所以**剛好相等時
走的是 else**，不是扣到 0。

字串是英文原文 `'is fighting with snakes'`——DOS 版的訊息就是中文化的來源文本，
PC-98 版對應的是日文（同一支的日文對照見
[spec 633](633-alignment-encoding-and-effect-callbacks.md) 那批的寫法）。

## `0188h`：依種族給 0..3 的加成

```text
DS:6FA3h := arg_6^[18Dh]^[0Ah]
case DS:6FA3h^[11Ah] of                        ← RACETYPE
    0Ah:        v := 1
    9, 0Ch:     v := 2
    4:          v := 3
    否則:        v := 0
end
DS:6F9Fh := DS:6F9Fh + v
DS:6F94h := DS:6F94h + v                       ← 傷害值（spec 641）
DS:6F95h := 9                                  ← 旗標欄位整個覆寫
```

**同一個加成同時加到兩個全域**。`DS:6F94h` 是傷害值、`DS:6F95h` 是旗標欄位
（[spec 641](641-dos-field-offset-shift.md) 由與 PC-98 的動作對應定出）。

`case` 有 `else` 分支（`v := 0`），所以**沒有
[spec 629](629-spell-pack-idiom-and-uninit.md) `2690h` 那種未初始化問題**。

種族 `9` 與 `0Ch` 共用同一個結果——兩個不同的種族值給一樣的加成。

## `03E5h`：兩道條件才動手

```text
if <sub_3C>(0Ah, arg_2^[3], 0Fh, arg_6) = 0 then return
if arg_6^[1A4h] <= 1 then return               ← 目前 HP 不足
DS:6F95h := 0
<far 014Ah:0494h>(arg_6, 0, 0, 1)
if DS:4FBAh = 5 then return
<far 014Ah:04FAh>(DS:6506h)
```

第二道是 `jbe`——**HP 等於 1 也不動手**，要嚴格大於 1。

最後那道 `DS:4FBAh = 5` 的檢查在**動作之後**：前面的 `014Ah:0494h` 一定會執行，
只有最後一個呼叫被跳過。

## 明確不宣稱

- `sub_3C`／`sub_1572` 與各 far routine 的行為。
- `DS:6F9Fh`／`DS:4FBAh`／`DS:6506h` 的語意。
- `+19Ch`／`+19Dh` 兩個欄位分別代表什麼（只知道相加當一個總量用）。
