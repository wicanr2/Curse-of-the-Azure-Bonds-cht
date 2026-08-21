# 1169 — 法術主表那 13 筆無名列不是占位，是**物品效果列**

- 證據等級：`exact`（九支 handler 逐條讀完；分派表 101 筆全解；訊息字串逐筆取自
  overlay-22 的常數區）
- 上游：spec 921（USE 的分派）、spec 1009（目標模式）、spec 1111（屬性表版面）、
  spec 1124（傷害骰子）、spec 815（把這 13 筆稱作占位的出處）
- 產物：`gamepack/rules/spell-damage.json`、
  [`docs/audit/spell-damage-table.md`](../audit/spell-damage-table.md)

## 原版自己說了

火球的 handler（`overlay-22 entry#64`，`3703h`）第二條指令就是：

```asm
3714: mov  BYTE PTR ds:0x6f9d,0x1     ; 這是範圍效果
3719: cmp  BYTE PTR ds:0x6f97,0x40    ; ← 目前的法術編號是 40h 嗎
371e: jne  0x3735
3720:   push 1 / push 3 / call <entry#9>   ; 1d3
372b:   shl ax,1 / inc ax                  ; 骰數 := 1d3 × 2 + 1
3733:   jmp  0x3741
3735: else 骰數 := <overlay-24 entry#44>(法術編號)   ; ＝ 施法者等級
```

`40h` 是法術主表裡**沒有名字**的一列。原作為它在火球裡分了一條路，還給了
`1d3 × 2 + 1` 顆 d6——**如果那一列玩家取不到，這段程式沒有存在的理由**。

取得的管道就是 USE：充能物品的效果編號 ＝ `物品^[3Dh] and 7Fh`（spec 921），
而那個編號就是主表的列號。`40h` 的持有者是**項鍊**（第 2 章第 3 區塊，7 次充能）
——AD&D 的飛彈項鍊，一發威力隨機的火球。

⇒ spec 815 說這 13 筆「玩家取不到」是錯的。**它們是物品專用的效果列**，
沒有名字是因為玩家看到的是物品名，不是法術名。

## 九列的完整讀法

`+6` 目標模式、`+8` 豁免種類、`+0Ah` 效果碼全部取自主表（spec 1111）；
骰子與訊息取自各自的 handler。`—` 代表該欄是 0。

| 效果 | 持有物品（章:區塊，充能） | handler | 目標 | 骰子／效果 | 訊息 |
|---|---|---|---|---|---|
| `39h` | Potion of Speed（2:2、4:32，1）| `entry#72` | 自己 | 效果 `27h` 加速 | `is Speedy` |
| `3Bh` | Potion（2:2，1）| `entry#74` | 自己 | 力量改成 `18/(1d4×10＋40)` | `is stronger` |
| `3Dh` | Wand（4:35，10）| `entry#76` | 1 個 | 效果 `34h` 定身，豁免無效 | `is paralyzed` |
| `3Fh` | Dust（2:2，1）| `entry#78` | 4 個 | 效果 `47h` 隱形 | `is invisible` |
| `40h` | Necklace（2:3，7）| `entry#64`（＝火球）| 半徑 3 | `(1d3×2＋1)d6` 火，豁免減半 | （無）|
| `41h` | Wand of Magic Missiles（4:32，88）| `entry#79` | 1 個 | `2d4＋2`，旗標 `08h` | （無）|
| `61h` | Potion of Invisibility（2:2、4:32，1）| `entry#106` | 自己 | 效果 `19h` 隱形 | （無）|
| `62h` | Wand of Beaker（3:17，89）| `entry#107` | 半徑 3 | `6d6`，**只有植物吃得到** | （無）|
| `63h` | Potion／Potion Extra Healing（2:2、4:32，3）| `entry#108` | 自己 | 治療 `2d4＋2` | `is Healed` |

四個訊息字串（`is Speedy`／`is stronger`／`is paralyzed`／`is invisible`）
就是把上表釘死的東西：無名列的語意不必用物品名去猜，**原作自己在 handler 裡寫了**。

## ★ `62h` 只打植物

```pascal
取範圍(中心 ← 施法者座標, 半徑 3);
for i := 1 to 目標數 do begin
    目標 := 目標表[i];
    if 目標 = NIL then continue;
    豁免成功 := (目標^[11Ah] <> 12h);              { RACETYPE }
    <overlay-23 entry#20>(目標, entry#10(6, 6), byte[3E02h], 豁免成功);
end;
```

`byte[3E02h]` 是**這一列自己的 `+8`**（`37DAh ＋ 98 × 16 ＋ 8`，98 ＝ `62h`），
值是 1 ＝ 豁免完全無效。`entry#20` 的第四個參數為非 0 時就照 `+8` 減傷，
所以：

- `RACETYPE = 12h` → 豁免視為失敗 → **吃滿 6d6**；
- 其餘 → 豁免視為成功 ＋ `+8 = 1` → **傷害歸 0**。

`RACETYPE = 12h` 在六章的怪物資料裡是四種：`BIT O' MOANDER`、
`LG VEGEPYGMY`、`SM VEGEPYGMY`、`SHAMBLING MOUND`——全部是 Moander 的植物眷屬。
89 次充能的魔杖只對這一類生效。

## `entry#20` 的四個參數

`overlay-23:1FD6h`（`retf 0Ah`）：

```pascal
procedure 結算(目標: 遠指標;   { bp+0Ch }
               傷害: byte;     { bp+0Ah }
               豁免種類: byte; { bp+08h }
               豁免成功: byte) { bp+06h }
begin
    DS:6F94h := 傷害;
    <sub_3FE>(6, 目標);
    if 豁免成功 <> 0 then begin
        if 豁免種類 = 1 then DS:6F94h := 0
        else if 豁免種類 = 2 then DS:6F94h := DS:6F94h div 2;
    end else
        <sub_3FE>(14h, 目標);
    if DS:6F94h = 0 then 離開;
    ...
end;
```

⚠ **`豁免種類 = 3` 沒有分支**——與 spec 1111 對 `+8` 的觀察一致，不要為了對稱補。

## ⚠ 這個缺口是自己的過濾器造成的

`scripts/spell_damage_table.py` 原本寫

```python
if spell["effect_id"] != 0 or spell["placeholder"]:
    continue
```

那個 `or spell["placeholder"]` 讓 `40h`／`41h`／`62h`／`63h` 四列從來沒被讀過，
於是「這幾個效果沒有骰子」看起來像是資料裡沒有——**其實是查詢自己有洞**。
排除條件的來源是 spec 815 的「玩家取不到」，而那句話本身沒有被驗證過。

規則：**排除條件也是斷言**，要跟結論一樣需要證據。

## 中文化

四句訊息接在名字後面，句型與 spec 921 的 `uses an item` 相同：

| 原文 | 建議 |
|---|---|
| `is Speedy` | `動作加快了` |
| `is stronger` | `力氣變大了` |
| `is paralyzed` | `被定住了` |
| `is invisible` | `隱形了` |
| `is Healed` | `傷勢復原了` |

## 明確不宣稱

- 沒有宣稱 `3Bh` 那一支的 `entry#19`／`entry#11`／`entry#24` 各自做什麼；
  只讀出「查一次 `15h`、把 `92h` 那一格設成 `1d4×10＋40`、印 `is stronger`」的形狀。
- 沒有宣稱 `sub_3FE(6, …)`／`sub_3FE(14h, …)` 是什麼（形狀上像 `CHECKFX` 的時機）。
- `5Fh` 與 `60h` 的持有者是三件**名字叫 `Scroll` 但裝備槽是 `0Ah`** 的物品
  （spec 1171）：原作按槽判別，它們走的是充能物品那條路。兩支 handler 都是
  「只套 `+0Ah`、沒有訊息」的形狀（效果碼 `49h`／`6Dh`，兩者的修正表都還是
  `unread`）。剩下 `3Ch`、`3Eh` 兩列在六章的 `ITEM*.DAX` 裡沒有持有者。
- 沒有宣稱 `RACETYPE = 12h` 的官方名稱（本文件只列出資料裡屬於它的四種怪物）。
