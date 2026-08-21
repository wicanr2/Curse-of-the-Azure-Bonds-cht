# 1165 — 三份小記錄的欄位名，以及六處讀法待對

- 證據等級：`exact`（PC-98 型別／成員表逐筆讀出）
- 前置：spec 1164（record 版面怎麼展開）
- 產物：[`docs/audit/pc98-record-layouts.md`](../audit/pc98-record-layouts.md)

spec 1164 打開的那張表不只有角色記錄。這一篇把另外三份逐欄對完，並把**名字與
既有讀法不一致**的六處集中列出來——名字不等於語意，符號給的是 Pascal 端的欄位
名，怎麼用仍以逐條讀出來的程式碼為準。

## 22 bytes 戰鬥狀態 ＝ `COMBATVARREC`

逐欄對照在 spec 806。三件事值得單獨提：

- **`+0Bh`／`+0Dh` 不是欄位。** `TARGET` 是 4 bytes 的 `CHARRECPTR`（`+0Ah`），
  那兩格是指標中間的位元組。spec 806 原本把它們列在「還沒看到用途」。
- **`+11h` ＝ `TRIEDTOTURN`。** 「還沒看到用途」的最後一格。
- **`+05h` ＝ `ZEROLEVBLOWS`，而角色記錄 `+0DDh` ＝ `ZEROLEVELBLOWS`。**
  兩份記錄各自獨立地印證 spec 806 的「`c^[5] := 角色^[0DDh]`」。

## `.SWG` 物品記錄 ＝ `CHARITEMFILREC`

十四欄一一對上：`NAME`／`TITLE`／`NEXT`／`ITEMPTR`／`NAMENUM`／`PLUS`／
`PLUSSAVE`／`READY`／`IDENTIFIED`／`CURSED`／`ENCUMBERANCE`／`NUMITEMS`／
`VALUE`／`SPECIAL`。

`+35h` 叫 `IDENTIFIED` ——與 spec 1036 讀到的名字一致（它已經寫著
「`identified := 4`」）。spec 1036 留的問題「寫成 `6`／`4` 各代表什麼」在這裡
有了形狀：remake 把它當成蓋住 `NAMENUM` 三格的遮罩（`6` ＝ 藏後兩格、
`4` ＝ 藏第一格），與「已鑑定的位元」是同一個位元組的兩種極性讀法。

## `.FX` 效果記錄 ＝ `EFFECTREC`

五欄：`EFFECTNUM`／`DURATION`／`SPECIAL`／`SPECIALOFF`／`NEXT`。

原版樣本 `CHRDATA1.FX` 三筆都是 `EFFECTNUM=61h`／`1Ah`／`2Fh`，
`DURATION=0`，`+03h=FFh`，`+04h=0`。

## 六處讀法待對

| 記錄 | 位移 | 既有讀法 | PC-98 欄位 |
|---|---|---|---|
| 戰鬥狀態 | `+0Fh` | 動作計數（spec 786）| `NUMATTACKERS` |
| 戰鬥狀態 | `+10h` | 「先前已經逃走」（`internal/combat`）| `TURNED`（被牧師逼退）|
| 戰鬥狀態 | `+12h` | 累計轉向量 mod 8（spec 786）| `ATTACKSPACING` |
| 戰鬥狀態 | `+13h` | 超編旗標（spec 801／805）| `CHARTYPE` |
| `.SWG` | `+00h` | 42 bytes 右側補白的字串 | `NAME: STR40`（41）＋ `TITLE`（1）|
| `.FX` | `+03h` | 強度，`0FFh` ＝ 永久（spec 441）| `SPECIAL` |

`+10h` 兩者的可觀察行為相同（都是掉頭就跑），差別在來源：轉化不是士氣崩潰，
士氣那條是 `+14h`（`ROUTING`）。其餘五處要回去讀 DOS 側的用法才能定案。

**這一輪只並列不覆蓋**：既有的 `decoded` 判讀是從程式碼讀出來的，光憑一個名字
不足以推翻它。台帳與規格裡兩種說法都留著，並標出待對。

## 明確不宣稱

- 沒有改任何解析器。`.SWG` 的名字仍然讀 42 bytes，`.FX` 的 `+03h` 仍然當強度。
- 沒有宣稱 DOS 與 PC-98 的欄位語意一定相同——只確認版面對得上。
