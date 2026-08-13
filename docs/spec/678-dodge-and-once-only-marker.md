# 第六百七十八輪：閃避判定，與用 bit 4 當「只做一次」的標記

狀態：`READY`。等級：`exact`。日期：2026-08-14
位置：PC-98 `overlay-12` 的 `0F5Bh`、`0C0Ch`。

## `0F5Bh`：三道條件的閃避

```text
if DS:9594h^[152h]:[154h] = nil then return          ← 沒有那個東西就不能閃
if <far 014Ah:0565h>(arg_2, arg_4, DS:9594h) <= 1 then return
if ROLLDICE(1, 64h) > arg_0 then return               ← 百分比判定
備妥 'は避けた。'，<far 014Ah:0522h>(arg_2, arg_4, 訊息, 1, 0Ah)
DS:0A02Eh := 0                                        ← 傷害歸零
DS:0A039h := 0FFh
DS:0A03Ah := DS:0A03Ah − 1
```

`arg_0` 直接就是**成功百分比**：`1d100 <= arg_0` 才算閃過。`arg_0 = 0` 時永遠失敗
（`1d100` 最小是 1），`arg_0 >= 100` 時永遠成功。

閃過的後果有三個：**傷害歸零**（`DS:0A02Eh`，[spec 640](640-save-for-half-and-damage-global.md)）、
`DS:0A039h` 設成 `0FFh`、`DS:0A03Ah` 減一。

`DS:0A039h`（＝ DOS 的 `DS:6F9Fh`，[spec 677](677-pc98-overlay12-twins.md)）在別處是
被**加減**的（`+= 種族加成`、`-= 4`、`-= 2`），這裡卻是**整個覆寫成 `0FFh`**。同一個
變數兩種用法，remake 不能把它當成單純的累加器。

第二道條件的回傳值要**大於 1**，不是非零——等於 1 也擋掉。

## `0C0Ch`：bit 4 只讓它發生一次

```text
if (arg_2^[3] and 10h) = 0 then
    arg_2^[3] := arg_2^[3] + 10h                      ← 設 bit 4
    備妥 'は歳をとった。'，<far 014Ah:0522h>(arg_6, 訊息, 1, 0Ah)
    arg_6^[76h] := arg_6^[76h] + 1                    ← word，加一
DS:0A030h := byte(DS:0A030h × 2)                      ← 無條件
```

`+3` 的 **bit 4 是「已經套用過」的標記**：第一次呼叫才印訊息並讓 `+76h` 加一，
之後再呼叫就只做最後那一行。

設 bit 用的是 `add 10h` 而不是 `or 10h`——因為前面已經確認該 bit 為 0，兩者等價。

`DS:0A030h × 2` **在條件外面**，每次呼叫都會做。乘完只存回 byte，所以第 8 次呼叫
之後就固定是 0。

`+3` 這個 byte 到目前為止有三種用途：bit 4 是本輪的「已套用」標記、
[spec 638](638-overlay12-effect-batch2.md) 的 `1222h` 也取 bit 4、
[spec 677](677-pc98-overlay12-twins.md) 的 `0773h` 取高 4 bit 當骰子面數。
**bit 4 同時屬於「高 4 bit」**，所以骰子面數會被這個標記改變——`0C0Ch` 執行過的
節點，`0773h` 算出的面數會多 1。

## 明確不宣稱

- `DS:0A030h`／`DS:0A03Ah` 的用途。
- `+152h`／`+154h`（far pointer）與 `+76h` 的語意。
- `014Ah:0565h` 回傳值的意義。
