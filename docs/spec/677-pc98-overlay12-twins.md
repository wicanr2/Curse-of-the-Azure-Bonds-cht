# 第六百七十七輪：PC-98 `overlay-12` 三支 —— 全域對照再擴一筆

狀態：`READY`。等級：`exact`。日期：2026-08-14
位置：PC-98 `overlay-12` 的 `0183h`、`054Ch`、`0773h`。

## `0183h` 是 DOS `0188h` 的對應

兩支的 `case` 完全相同（`RACETYPE` `0Ah`→1、`9`／`0Ch`→2、`4`→3、其餘→0），
差別只在寫哪三個全域：

| | DOS（`0188h`，[spec 676](676-dos-overlay12-effects.md)） | PC-98（`0183h`） |
|---|---|---|
| 加成目標一 | `DS:6F9Fh` | **`DS:0A039h`** |
| 加成目標二 | `DS:6F94h` | `DS:0A02Eh` |
| 覆寫成 9 | `DS:6F95h` | `DS:0A02Fh` |

後兩組 [spec 641](641-dos-field-offset-shift.md) 已經定過（傷害值與旗標欄位），
**第一組是本輪新增**：`DS:6F9Fh` ↔ `DS:0A039h`。

`DS:0A039h` 在 [spec 633](633-alignment-encoding-and-effect-callbacks.md)／
[spec 638](638-overlay12-effect-batch2.md)／[spec 639](639-overlay12-poison-and-race-gates.md)
被多支加減過，一直沒有跨平台對應；這裡由「兩支同構函式寫哪個位址」定出來。

## `054Ch`：毒死的訊息，以及一個臨時旗標

```text
if <far 014Ah:0547h>(arg_6, arg_8, 37h, @var_4) <> 0 then
    備妥 'は毒のために死んだ。'，<far 013Eh:0405h>(arg_6, arg_8, 6, 訊息)
DS:0A036h := 1
<sub_140Fh>(arg_6, arg_8, 0Fh, 0)
DS:0A036h := 0
```

效果 id `37h` 與 [spec 639](639-overlay12-poison-and-race-gates.md) 的 `147Ah`
（毒）是同一個。那支印「は毒を受けた。」＋「は死んだ。」，這支印
「は毒のために死んだ。」——**兩條不同的毒死路徑，訊息不一樣**。

`DS:0A036h` 在呼叫前後**設 1 再設回 0**，是個臨時旗標。
[spec 633](633-alignment-encoding-and-effect-callbacks.md) 的 `003Ch` 正是
「`DS:0A036h = 0` 就回傳 0 什麼都不做，否則回傳 1 並呼叫 `sub_1437`」——所以
**只有在這支的中間那一段，`003Ch` 才會真的動作**。

兩支合起來才看得出這個旗標的用途：一支設、一支查。

## `0773h`：骰子面數由欄位算出

```text
n := (arg_2^[3] div 16) + 1                   ← 取高 nibble 再加一
r := ROLLDICE(1, n)                            ← 1 顆 n 面骰
if r = 1 then …
```

`div 16` 用的是 `cwd` ＋ `idiv`（**有號**），但 `arg_2^[3]` 先 `xor ah, ah` 零延伸，
所以實際是無號。取的是效果節點 `+3` 的**高 4 bit**。

**面數不是常數而是由資料算出**——同一支程式碼可以擲 1d1 到 1d16。`+3` 的高 nibble
為 0 時得到 `1d1`，結果恆為 1。

`+3` 的低 nibble 另有用途（[spec 638](638-overlay12-effect-batch2.md) 的 `1222h`
取它的 bit 4）。

## 明確不宣稱

- `DS:0A039h`／`DS:6F9Fh` 代表什麼（只知道兩者對應，且被多支加減）。
- `sub_140Fh`／`014Ah:0547h` 的行為。
- `0773h` 在 `r = 1` 之後做什麼（本輪只讀到判斷處）。
