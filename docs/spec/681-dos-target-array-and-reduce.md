# 第六百八十一輪：DOS 側的目標陣列，與「已被削弱」的流程

狀態：`READY`。日期：2026-08-14
位置：DOS `overlay-22` 的 `20E1h`、`15FBh`。

## `DS:7434h`／`DS:7435h` 對應 PC-98 的目標陣列

`20E1h` 的開頭是：

```text
p := DS:7435h（far pointer）
if p = nil then return
if DS:7434h <= 0 then return               ← byte 計數
```

**「一個 byte 的計數 ＋ 緊接著的 far pointer」**——與
[spec 630](630-spell-target-array.md) 的 PC-98 目標陣列同一個形狀：

| | PC-98 | DOS |
|---|---|---|
| 陣列基底 | `0A51Dh` | `7431h`（推得） |
| 計數 | `0A520h` | **`7434h`** |
| 第 1 筆 | `0A521h` | **`7435h`** |

等級 `strong inference`：形狀相同、用法相同（都是「計數為 0 就不做」＋「取第 1 筆
當目標」），但沒有直接的同構函式對照。基底位址是由「第 1 筆 ＝ 基底 ＋ 4」推回去
的，本輪沒有看到直接引用基底的指令。

## `20E1h`：四道條件才削弱

```text
p := DS:7435h
if p = nil 或 DS:7434h = 0 then return
if <loc_1456h>(0, 4, p) <> 0 then return            ← 這一道要「等於 0」
if <loc_1573h>(p, 0Ch, @var_4) = 0 then return      ← 這一道要「不等於 0」
<loc_143Dh>(p, 0Ch, 0)                               ← 移除效果 0Ch
<far loc_14A4h>(p, 0)
備妥 'has been reduced'，顯示
```

**兩道相鄰的檢查方向相反**：第三道要求回傳 0、第四道要求回傳非 0。抄的時候很容易
把其中一個寫反。

效果 `0Ch` 先確認**存在**（第四道）才移除——移除與訊息是同一條路徑上的兩步，不會
出現「印了訊息卻沒移除」的情況。

字串 `'has been reduced'` 是英文原文，PC-98 側對應的是日文。

## `15FBh`：四個 byte 參數擴成一個四 word 的記錄

```text
ss:[arg_0 − 1Ch] := signext(arg_8)
ss:[arg_0 − 1Ah] := signext(arg_6)
ss:[arg_0 − 18h] := signext(arg_4)
ss:[arg_0 − 16h] := signext(arg_2)
<far loc_1898h>(ss:(arg_0 − 1Ch))
```

四個 byte 參數各自 `cbw`（**有號**擴展）成 word，依序放進呼叫端堆疊上的四個位置，
再把那塊的位址傳下去。

`arg_0` 是**呼叫端的 `BP` 值**（用 `ss:` 定址），所以這支直接寫進呼叫端的區域變數
區——不是自己的 frame。`add di, 0FFE4h` 就是 `− 1Ch`。

參數順序在傳遞時**反過來**：`arg_8` 放最前面、`arg_2` 放最後。

## 明確不宣稱

- `loc_1456h`／`loc_1573h`／`loc_143Dh`／`loc_14A4h`／`loc_1898h` 的行為。
- 效果 `0Ch` 的語意。
- `15FBh` 組出來的四 word 記錄代表什麼（座標、矩形、或別的）。
