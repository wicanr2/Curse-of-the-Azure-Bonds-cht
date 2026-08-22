# 1180 攻擊次數以「半次」為單位，而整數次數靠**回合奇偶**換算

狀態：`READY`

## 結論

原作的攻擊次數不是整數，是**半次**：`2` ＝ 一回合一次、`3` ＝ 一次半、
`8` ＝ 四次。換成「這一回合實際打幾下」只有一個函式：

```
ADJUSTBLOWS(半次值) = (半次值 + (ROUND and 1)) div 2
```

★ 加的是**回合數的最低位**，所以奇數的半次值會在回合之間交替：`3` 在第 1、3、5
回合給 **2** 次，第 0、2、4 回合給 **1** 次，平均剛好 1.5。這就是 AD&D 的
「每兩回合三次攻擊」。**寫成 `半次值 / 2` 會每一回合都少半次，而且不會有任何徵兆。**

## 名字都是原作自己的

PC-98 版帶完整 Borland 除錯符號，這一串全部是原名：

| 位址／欄位 | 符號 | 內容 |
|---|---|---|
| `overlay-13:0F12h`（DOS）| **`ADJUSTBLOWS`** | 上面那個換算 |
| `overlay-13:0E3Eh` 附近 | `FIGBLOWS` | 算出基準值 |
| 角色 `+11Ch`（2 bytes）| **`BASEATTBLOWS[0..1]`** | 兩個武器槽的基準值（半次）|
| 角色 `+19Ch`（DOS，2 bytes）| **`ATTBLOWS[0..1]`** | 這一回合的整數次數 |
| 角色 `+0DDh` | `ZEROLEVELBLOWS` | |
| `DS:758Dh` | **`ROUND`** | 本場戰鬥的回合數，開場 0 |
| `DS:758Eh` | **`LASTROUND`** | ＝ `ROUND + 0Fh` |
| `DS:75E6h` | `SHOWDUDEMOVE` | spec 808 的「這個角色要不要畫」，名字確認 |

`ROUND`／`LASTROUND` 的認定不靠猜：spec 1019 早就量出 PC-98 資料段 ＝ DOS ＋ `3292h`，
而 `758Dh + 3292h = A81Fh` 正是符號表裡 `ROUND` 的位址，`758Eh` 對到 `LASTROUND`。
兩者相鄰，也對得上 `LASTROUND := ROUND + 0Fh` 那一行。

`ADJUSTBLOWS` 在兩個平台上**位元組完全相同**，只有資料位址不同
（DOS `758Dh` ↔ PC-98 `A81Fh`）。

## 取捨：遠程會取代基準值

`overlay-13:0DD9h`（spec 808）決定送進 `ADJUSTBLOWS` 的是哪一個值：

```pascal
if 有遠程武器 and 有彈藥 then
    v := byte[5CFBh + 武器類別 × 10h]     { 類別表 +5 射速，下限夾成 2 }
else
    v := 角色^[11Ch]                      { BASEATTBLOWS[槽] }
n := ADJUSTBLOWS(v);
if 遠程 and (彈藥^[39h] > 0) and (彈藥數 < n) then n := 彈藥數;
```

CoAB 的 `ITEMS` 裡射速**全部是偶數**（2、4、6、12、28、40），所以遠程那一路的
奇偶項永遠不改變結果——半次的意義只在近戰的 `BASEATTBLOWS` 上顯現。

## 槽的選擇

`overlay-13`（spec 1010）動手前先挑槽：

```pascal
槽 := 1;
if 打手^[11Ch] = 0 then 槽 := 2;          { BASEATTBLOWS[0] 是 0 就用第二個槽 }
for j := 1 to 2 do
    if 打手^[19Bh + j] > 0 then 槽 := j;   { 19Bh+j ＝ ATTBLOWS[j−1] }
```

原版資料印證：`GIANT SPIDER` 與 `PHASE SPIDER` 的 `BASEATTBLOWS` 是 `{0, 2}`
——牠們槽 0 沒有攻擊，用槽 1 咬。

## remake 先前錯在哪

`monster.Record` 把 **`+0A1h`** 當成每回合攻擊次數。那一格：

- 落在 `SPELLSKNOWN`（`+079h`，100 bytes）**裡面**；
- 六章 68 種怪物**全部是 0**；
- 整包 overlay 反組譯裡**沒有任何一處讀它**（同樣的掃描對 `+0DEh`／`+11Ch`／
  `+19Ch`／`+0DDh`／`+151h` 分別找到 10／11／10／9／14 處，所以不是掃描方法的問題）。

後果是每一隻怪物都只打一下。改讀 `BASEATTBLOWS[0]` 之後：

| 半次值 | 每回合次數 | 種數 |
|---:|---:|---:|
| 0 | 用槽 1 | 2 |
| 2 | 1 | 51 |
| 4 | 2 | 13 |
| 8 | 4 | 2 |

**15 種怪物**（`TROLL`、`RAKSHASA`、`DRACOLICH`、`BLACK DRAGON`、`THRI-KREEN`…）
先前每回合少打一到三下。

## remake 這一側

- `monster.Record.AttackBlows [2]uint8` ← `+11Ch`／`+11Dh`，`AttacksPerTurn` 移除。
- `combat.Fighter.AttackBlows [2]int`，`combat.AdjustBlows(blows, round)`，
  `(*Battle).AttacksThisRound(f)` 走原作的取捨：架著遠程武器就用已投影的整數
  次數，否則 `AdjustBlows(BASEATTBLOWS[槽], ROUND)`。
- 加速（`27h`）／緩速（`2Ah`）改成乘除**半次值**再換算。先換算再加倍會把 1.5 次
  的那半次先丟掉。

## 回歸

- `TestAdjustBlowsAlternatesOnOddHalfBlows`：0..8 的真值表，外加「連續兩回合
  加起來 ＝ 半次值本身」——這一條直接擋住 `/2` 那種寫法。
- `TestMonsterRecordsCarryBaseAttackBlows`：拿原版六章資料驗分布，並且**釘死
  `+0A1h` 必須是 0**。突變驗過：把讀取端改回 `+0A1h`，三個半次值同時消失。
- `TestAttacksThisRoundFallsBackToTheSecondSlot`、
  `TestAttackSequenceUsesTheRecordBlows`、
  `TestAttacksThisRoundPrefersTheProjectedMissileRate`。

## 不宣稱

- **彈藥數上限還沒接**：spec 808 的 `彈藥^[39h]` 那一段 remake 這一側沒有實作，
  遠程攻擊次數目前不受箭矢數量限制。
- 隊伍側的 `BASEATTBLOWS` 怎麼由職業與等級算出來（`FIGBLOWS`、spec 771 的
  `overlay-13:1CE2h` ＋2／＋3 累加）沒有接：remake 的隊員仍只從武器射速拿次數，
  高等級戰士的一次半還沒有。
- `overlay-13:0DD9h` 裡 `15BCh+1`／`15C6h+1` 兩個判斷式的內部沒有讀。
- `LASTROUND`（`ROUND + 0Fh`）用來限制什麼沒有讀完。
