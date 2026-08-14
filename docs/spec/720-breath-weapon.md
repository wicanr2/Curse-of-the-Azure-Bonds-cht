# 第七百二十輪：龍息，以及 `274Ch` 算的是距離

狀態：`READY`。等級：`exact`（流程）／`strong inference`（`274Ch` ＝ 距離）。
日期：2026-08-14
位置：DOS `overlay-22:5E11h`（`retf 0Ah`）。

## 流程

```text
<間接 far 呼叫 [DS:72A0h]>(@DS:7746h, 1, 41h)
DS:6FA3h := arg_6^[18Dh]^[0Ah]                      ← 目標

if ROLLDICE(1, 100) > 50 then 返回                   ← 五成機率直接不發動
if <overlay-24 entry#33（274Ch）>(arg_6, DS:6FA3h) >= 2 then 返回
if DS:6FA3h = NIL then 返回                          ← ⚠ 檢查在用過之後

DS:6F95h := 1                                        ← 傷害旗標
DS:7746h := <overlay-24 entry#34（2828h）>(arg_6)     ← 回傳常數 1（spec 691）
<overlay-24 entry#20>(arg_6, 'Breathes Fire'（CS:5E03h）, 0Ah, 1)
<overlay-24 entry#24>(17h)

a := <overlay-32 entry#15>(arg_6)   b := <overlay-32 entry#16>(arg_6)
c := <overlay-32 entry#15>(DS:6FA3h) d := <overlay-32 entry#16>(DS:6FA3h)
<overlay-24 entry#25>(a, b, c, d, 1, 1Eh)            ← 兩組座標，像是動畫

r := <overlay-23 entry#8>(DS:6FA3h, 3, 0)
<overlay-23 entry#20>(DS:6FA3h, 7, 2, r)
```

## `274Ch` 回傳的是距離

spec 699 只讀出 `274Ch` 的機械行為：借共用緩衝查一次，回傳
`byte [6E95h + k×3] div 2`。這裡的用法把語意補上了：

```text
if <274Ch>(施法者, 目標) >= 2 then 不發動
```

龍息要相鄰才打得到，`>= 2` 就放棄——所以那個回傳值是**距離**，而且因為原始欄位
除以 2，單位是「半格」之類的東西。`0`／`1` 算相鄰。

這同時解釋了 spec 699 那個「找不到就回傳最後一筆」的行為為什麼沒被當成問題：
查不到的時候拿到的多半是個大值，自然落在「太遠」那一邊。

## 三件要注意的

**`NIL` 檢查在用過之後。** `DS:6FA3h` 先被拿去呼叫 `274Ch`，才檢查是不是
`NIL`。順序反了，但因為 `274Ch` 只讀不寫、而且 `NIL` 解出來是中斷向量表的內容
（spec 716），實務上不會當掉——只是那次距離計算的結果是垃圾。

**間接 far 呼叫在呼叫圖裡是看不到的。** `call dword ptr ds:72A0h` 沒有目標位址
可以解析，`scripts/overlay_call_graph.py` 抓不到（它只認 `9A`）。
`overlay-22` 裡這種寫法有多少支需要另外掃。

**五成機率寫死在程式裡。** `ROLLDICE(1, 100) > 50` 不查任何表格，也不受等級或
屬性影響。

## 明確不宣稱

- `[DS:72A0h]` 這個函式指標指向誰、三個參數的意義。
- `overlay-24` entry#25、`overlay-23` entry#8／#20 的行為。
- `DS:7746h`、`DS:6F95h := 1` 的語意。
- 傷害數值——本輪沒看到任何骰子，傷害應該由 `entry#20` 那邊決定。
