# 第六百零九輪：`2Eh`（範圍效果與全滅判定）

狀態：`READY`。等級：`exact`。日期：2026-08-14
位置：PC-98 `overlay-02:2ACEh`（824 bytes，295 條指令）。

```text
saved := DS:9594h                             ← 進來存、離開還原
READVAR(5)
flags := ADDRESSVALUE(1)                      ← 位元旗標
a := ADDRESSVALUE(2) ; b := ADDRESSVALUE(3)
c := ADDRESSVALUE(4) ; d := ADDRESSVALUE(5)
amount := <loc_142D>(b, a) + c                ← 每次施加前重算
…依 flags 的位元對目標或全隊反覆做
    if <loc_1427>(…) then <far 0062:00DB>(node, amount)
…
DS:7F34h := 1
for each node in DS:9598h 鏈 do
    if node^[197h] <> 0 then DS:7F34h := 0    ← 只要有一個就清掉
if DS:7F34h <> 0 then                         ← 全隊都倒下
    <far 019E:0000>()
    DS:9636h := 2 ; DS:9637h := 2
    <顯示>「パーティーは全滅した！」（`unk_2AB7h`）
    <far 08EE:0379>(0BB8h)                    ← 0BB8h ＝ 3000
DS:9594h := saved
```

## `flags` 的位元

`ADDRESSVALUE(1)` 的結果被逐位元拆開：`10h`、`20h`、`40h` 各控制一段流程，
另有一個中間值取 `80h`／`1Fh`／`07h` 三種遮罩。**這不是一個數值而是一組開關。**

## `amount` 每次重算

`<loc_142D>(b, a) + c` 在函式裡出現**三次**，每次施加前都重新呼叫。若那支是
擲骰，代表**每個目標各擲一次**，不是全體共用一個結果。

## 全滅判定

- 用的是角色記錄 **`+197h`**——`KILLDUDE` 把它設 0、`STANDUP` 設 1
  （[spec 579](579-character-status-fields.md)）。所以「還活著」的判準是
  `+197h <> 0`。
- 迴圈**不提早結束**：找到活人只是把旗標清掉，仍然走完整條鏈。
- `DS:7F34h` 這個旗標與 `3Eh`／`38h(3)` 共用
  （[spec 596](596-ecl-party-item-sweep.md)、
  [608](608-ecl-terminate-modes.md)），語意一致：**1 代表全隊都不能行動**。
- 全滅時把 `DS:9636h` 與 `DS:9637h` 都設成 **2**（其他地方是設 1 或 11h），
  然後顯示訊息並延遲 `0BB8h`＝**3000**。

## 明確不宣稱

- `flags` 各位元的意義。
- `loc_142D`（算 `amount`）、`loc_1427`（條件）、`0062:00DB`（施加）的本體。
- `0BB8h` 的單位（毫秒？tick？）。
