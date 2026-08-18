# 1135 — 回合開始的生命週期：先攻算式的最後一格、`CHECKFX(07h)` 的呼叫時機

- 證據等級：`exact`（先攻的三支來源逐條讀完並用 `cmd/far-call-map` 把 far call
  解回 `overlay-NN entry#K`；remake 這一側逐項對回程式碼，缺口由測試釘住）
- 上游 spec 806（先攻值怎麼算出來）、spec 804（回合開始的重設）、spec 791（選誰動）、
  spec 1123（`CHECKFX` 兩張表）、spec 1113（突襲遮罩）

## 一、★ 先攻的基底是敏捷的遠程命中調整

spec 806 把先攻寫成 `<far 1524h+3>(角色) ＋ <far 145Ah+3>(6, 1)`，
其中 `<far 1524h+3>` 是 IDA 對 overlay far call 的假位址（spec 1112）。
`cmd/far-call-map` 解回真正的目標：

| spec 806 的寫法 | 真正的目標 | 是什麼 |
|---|---|---|
| `<far 1442h+2>` | `overlay-23 entry#4` | — |
| **`<far 1524h+3>`** | **`overlay-24 entry#11` ＝ `sub_120A`** | **敏捷的遠程命中調整**（讀角色 `+17h`）|
| `<far 145Ah+3>(6, 1)` | `overlay-23 entry#9` | 擲 1d6 |

`sub_120A` 是 AD&D 1e 的敏捷遠程命中調整表（spec 694 已讀）：
`0..2 → −4`、`3 → −3`、`4 → −2`、`5 → −1`、`6..15 → 0`、`16 → +1`、`17 → +2`、
`18 → +3`，另外 `19..20 → +3`、`21..23 → +4`、`24..25 → +5` 是本作自己補的。

完整算式因此是：

```
先攻 := 敏捷遠程命中調整(敏捷) + 1d6
若 先攻 < 1 則 先攻 := 1
若 (陣營 + 1) and 突襲遮罩 <> 0 則 先攻 := 先攻 − 6
若 先攻 < 0 或 先攻 > 20 則 先攻 := 0        { 0 ＝ 這回合不動作 }
```

⚠ **20 從擲骰擲不出來**：`1d6`（1..6）加上調整（−4..+5）上限是 11。
所以 `20` 是一個哨兵值，只有別的地方明寫才會出現——原作在回合開始把它壓成
`19`（spec 804），效果是「排在所有正常值前面，但仍是一個正常值」。

## 二、remake 這一側的逐項對照

| 項目 | 原作 | remake | 結果 |
|---|---|---|---|
| 先攻算式 | 上面那一段 | `viewport`／`combat/initiative` 的 `RollActionDelay` | **逐項相同**，含敏捷表的每一段 |
| 選誰動 | 取 `+3` 最大，同值擲 1..100（spec 791）| `SelectNext`：每一筆都擲 d100，相等時取後面那筆 | 相同（**delay 0 的也要擲**，少擲會換掉共用 PRNG 的續流）|
| `20 → 19` | 回合開始無條件壓（spec 804）| `BeginQuickFightAction` 只在 `QuickFight` 時壓 | 等價：20 只由快速戰鬥的交接旗標寫入 |
| 延遲（DELAY）| 同一回合重新插入 | `BeginScheduledRound` ＋ `NextScheduledTurn` ＋ `SetDelay`，`State.CombatDelay()` | 已接 |
| 防禦（GUARD）| 回合開始清掉，撐到自己下一次被選中 | `NextScheduledTurn` 清 `Guarding`，`State.CombatGuard()` | 已接 |
| 快速戰鬥（QUICK）| 交給 AI | `State.CombatQuick()`／`CombatQuickAll()`／`CombatToggleQuickMagic()` | 已接 |
| 定身（held）| `CHECKFX(7)` → `CLEARACTION` → 先攻歸 0 → 不動作 | `MonsterIsHeld()` 短路 ＋ `ClearAction` | 已接（實作位置不同，結果相同）|
| 突襲 −6 | `DS:4F9Dh^[596h]` 的位元遮罩 | CoAB 一律傳 `0` | **未接**——沒有人設那個遮罩（spec 1113 量到的正是這一點）|

`InitiativeBonus` 這個欄位只有 `cmd/azure-bonds-game` 的示範與 checkpoint 夾具
會設；正式隊伍投影不設，所以正式路徑跑的就是上面那條算式。

## 三、★ `CHECKFX(07h)` 要在「分派給 AI 或玩家選單」之前

原作 `overlay-08 entry#4`（`sub_26B`，COMBAT 單元的回合開始重設）的順序是：

```pascal
p^[18Dh]^[0Fh] := 0;                 { 動作計數 }
p^[18Dh]^[12h] := 0;                 { 累計轉向 }
p^[18Dh]^[07h] := 0;                 { 保留的機會攻擊 }
<CHECKFX>(p, 7);                     { ← 在這裡 }
if 有號(p^[18Dh]^[3]) <= 0 then 離開;
if p^[18Dh]^[3] = 14h then p^[18Dh]^[3] := 13h;
…
if 有號(p^[18Dh]^[3]) > 0 then
    if p^[198h] <> 0 then <AI>(p) else <玩家選單>(p);   { ← 分派在最後 }
```

`CHECKFX(7)` 在那個二選一**之前**，所以玩家操作的隊員一樣要套。remake 先前把
它擺在「隊伍且非快速戰鬥就交還 UI」之後，於是**玩家控制的角色永遠不會套到這個
時機的記錄寫入**——纏繞術那一族（效果 `88h`：把戰鬥狀態的移動率設成 0）對整支
隊伍無效，只有敵人與開了快速戰鬥的隊員會被減速。

⚠ 這種缺口**自洽的測試看不到**：AI 那一側走的是同一支函式，拿敵人來測會過。
`TestCanActRecordWritesApplyToPlayerControlledTurn` 因此明確測**玩家控制**的隊員，
並先確認它在修正前會紅（移動率 12 ≠ 0）。

⚠ 移到前面之後，這一段在玩家還在選的期間會被重新走到——移動與選目標都會回到
`advanceCombatToParty`。目前 `07h` 這個時機唯一有數字的修正是 `88h` 的
`set 0`，重複套用沒有差別；`TestCanActTimingModifiersAreIdempotent` 擋住之後
有人在這個時機加上加減型修正而沒有改成「每回合只套一次」。

## 明確不宣稱

- 沒有宣稱 `overlay-23 entry#4`（`<far 1442h+2>`）是什麼。
- 沒有宣稱 `+18Dh^[3]` 的量綱，也沒有宣稱為什麼 20 要壓成 19。
- 沒有宣稱 remake 的 `MonsterIsHeld()` 短路與原作「`CHECKFX(7)` → `CLEARACTION`
  → 先攻歸 0」在**所有**效果碼上等價；驗過的是結果（那一回合不動作）。
- 沒有宣稱回合開始的三個歸零（動作計數、累計轉向、保留的機會攻擊）在 remake
  有對應欄位——`+0Fh`／`+12h`／`+07h` 目前沒有投影進 `Fighter`。
- 沒有宣稱突襲遮罩由誰寫入；CoAB 仍傳 0。
