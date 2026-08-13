# 第五百九十輪：ECL 的條件分支結構

狀態：`READY`。等級：`exact`。日期：2026-08-14
模組：`INTERPET`（overlay-02）。位址為 PC-98 overlay-local。

## `14h` 是比較，`16h`～`1Bh` 是分支

兩者靠 `DS:A88Ah..A88Fh` 這六個連續的旗標位元組串起來。

```text
opcode 14h（0B49h）：
    FillChar(DS:A88Ah, 6, 0)              ← 先把六個旗標全部清零
    READVAR(4)
    for i := 1 to 4 do v[i] := ADDRESSVALUE(i)
    if v[1] = v[2] and v[3] = v[4] then DS:A88Ah := 1
    else                                   DS:A88Bh := 1

opcode 16h..1Bh（0BAEh，[spec 589](589-ecl-text-and-flag-handlers.md)）：
    if DS:[A88Ah + (op - 16h)] = 0 then <far 0062:00B1>()
```

於是 ECL 的條件式長這樣：**`14h` 比較兩對值 → `16h`／`17h` 依結果決定要不要
呼叫 `0062:00B1`**。`14h` 進來就把六個旗標全清零，所以每次比較都是乾淨的
起點，不會被上一次的結果污染。

`14h` 一次比較**兩對** word（`v1:v2` 與 `v3:v4`），兩對都相等才算相等。

`18h`～`1Bh` 對應的 `A88Ch`..`A88Fh` 這四個旗標**不是 `14h` 設的**，而是
`03h`（`COMPARE`）設的（[spec 593](593-ecl-comparison-flags.md)）。所以大小
比較的分支接在 `14h` 之後永遠不成立——`14h` 只設得出「相等／不等」。

## `10h`（`09DDh`）：輸入字串

```text
READVAR(2)
addr := ADDFNC(high[2], low[2])
<far 0418:1150>(0, 28h, 0, 0Ah, @len, @buf)     ← 最長 28h ＝ 40 字元
<0A65:0649>(@s, FFh)
if s = '' then s := ' '                          ← 空輸入填一個空格
<far 0062:0075>(s, addr)                         ← 寫回 ECL 位址
<far 0064:003E>()
```

`unk_9DBh` 是長度 1 的字串，內容就是一個**空格**。所以玩家直接按 Enter 時
存進去的不是空字串——remake 存空字串會讓後續的長度判斷走不同分支。

## `20h`（`0C26h`）

```text
READVAR(1)
bank0^[1E4h] := DS:BDF0h                  ← 先把舊值收起來
v := ADDRESSVALUE(1)
DS:BDF0h := v
<far 0062:0034>(v) ; <far 0062:002F>()
DS:7896h := 1 ; DS:7897h := 1
FillChar(DS:BDDAh, 2, 0)
```

舊值存進 bank 0 的 `+1E4h`、新值放 `DS:BDF0h`，形狀是「切換到另一個東西並
記住原本在哪」。

## 明確不宣稱

- `0062:00B1` 做什麼（跳過下一段？執行下一段？）——沒讀它的本體之前，
  不能說 `16h` 是「相等時執行」還是「相等時跳過」。
- `A88Ch`..`A88Fh` 由誰設定。
- `DS:BDF0h`／`DS:7896h`／`DS:7897h`／bank0 `+1E4h` 的語意。
- `0418:1150` 六個參數的意義。
