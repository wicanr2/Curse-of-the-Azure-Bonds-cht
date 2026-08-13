# 第五百八十二輪：豁免檢定與 `LOSEDUDE`

狀態：`READY`。等級：`exact`。日期：2026-08-14
模組：`EFFECTS`（overlay-23）。

## `MAKESAVE`：AD&D 豁免檢定

PC-98 `12D8h`／DOS `12EBh`，8 bytes 參數，3 個呼叫者。

```text
MAKESAVE(modifier, save_type, char) -> boolean
    DS:A042h := 1                                ← 結果放全域
    DS:A02Ch := ROLLDICE(1, 20)                  ← 骰值也放全域
    if DS:A02Ch = 1  then DS:A042h := 0 ; exit   ← 自然 1 必失敗
    if DS:A02Ch = 20 then DS:A042h := 1 ; exit   ← 自然 20 必成功
    DS:A02Ch := DS:A02Ch + char^[187h] + modifier
    DS:A041h := save_type
    CHECKFX(0Ch, char)                           ← 可再改寫 DS:A02Ch
    DS:A042h := (char^[0DFh + save_type] <= DS:A02Ch)
    return DS:A042h
```

- **角色記錄 `+0DFh` 起是豁免表**，以 `save_type` 直接當索引。
- `+187h` 是豁免的全域修正，與參數 `modifier` 相加。
- **自然 1／20 在加修正之前就判定**，所以任何加值都救不回自然 1。
- 骰值 `DS:A02Ch`、類型 `DS:A041h`、結果 `DS:A042h` 都是全域，`CHECKFX(0Ch)`
  就是靠這點介入。與 `ATTEMPTTOHIT` 用 `DS:A039h` 是同一套設計
  （[spec 577](577-attempttohit-and-effect-chain-walk.md)）。

## `LOSEDUDE`

PC-98 `1486h..1551h`（IDA 只認到 `150Ch`）。

```text
LOSEDUDE(msg, ?, new_state, char)
    <複製 msg 40 bytes 到區域>
    if char^[197h] = 0 then exit
    idx := <取索引>(char)
    <?>(0, 3, char)
    <顯示>(1, 0Ah, msg, char)
    char^[197h] := 0
    char^[196h] := new_state
    if char^[196h] <> 3 then char^[1A5h] := 0     ← 狀態 3 保留 HP
    <?>(0, 0, idx) ; <far 02A8:0B7D>()
    DS:[9740h + idx*4] := 0                       ← 清掉某張 4-byte stride 的表項
    <far 0189:0048>()
    if <far 014A:00CA>(char) then REMOVEFX(char)
```

狀態 `3` 是唯一不清 HP 的例外。

## IDA 的函式邊界在這裡切了五刀

`LOSEDUDE` 這一支被 IDA 切成 `1486h`／`1510h`／`1529h`／`1542h`／`154Ch` 五塊，
其中 `154Ch`（`mov sp,bp / pop bp / retf 0Ah`）是它的**共用出口**——函式開頭
那個 `jmp 154Ch` 就是「條件不成立就直接返回」。台帳把這四塊標為`邊界碎片`
是對的。

**更要緊的是 body 裡有缺口**：`1525h..1528h`（`d1 e7 d1 e7`＝`shl di,1` ×2）
與 `1548h..154Bh`（`0E E8 3E 00`＝`push cs / call REMOVEFX`）**IDA 完全沒認成
指令**，所以不在任何匯出裡。

這對兩件事有影響，都已在工具裡處理：

1. **讀函式一律走位址範圍**（`scripts/show.py --range`），它會把缺口明確印
   出來，再從原始位元組補。靠函式匯出讀會**安靜地漏掉指令**。
2. **跨平台助憶碼配對**（`cross_platform_pair.py`）比的是匯出的指令序列，
   缺口會讓序列變短。缺口位置兩平台不一定一樣，所以這只會造成**配不上**
   （少配），不會造成配錯——判準要求序列完全相同且各自唯一。

## 明確不宣稱

- `save_type` 的取值範圍與各值對應哪一類豁免（`+0DFh` 表的長度未定）。
- `DS:9740h` 那張 4-byte stride 表是什麼。
- `LOSEDUDE` 裡四個 far call 的本體。
