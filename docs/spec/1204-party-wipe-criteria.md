# 1204：全滅的兩個判準——POSTCOM 的 PARTYDEAD 與 ECL DAMAGE 收尾（READY）

原作判「隊伍全滅」的地方有兩處，**判準不相同**，遊戲行為也因此不同：
戰鬥打輸但全員只是昏迷，隊伍活下去；同一批人被劇情傷害放倒，遊戲結束。
本輪把兩支一手讀完、接進 remake，並解掉一個會誤導後續解讀的 offset 陷阱。

## 一、戰後：POSTCOM 的 `PARTYDEAD`（狀態集合判準）

PC-98 `overlay-05` `sub_1775`（POSTCOM 收尾），`17C9h..1843h`：

```
17C9  cmp  [7F34h], 0          ; PARTYDEAD 已被預設就不重算
17CE  jnz  183E
17D0  mov  [7F34h], 1          ; 初值 1
      …走角色名冊（頭 9598h、next +18Ah）…
17F3  cmp  es:[di+196h], 1     ; ANIMATED  ┐
17FE  cmp  es:[di+196h], 6     ; DEAD      │ 命中任何一個 → 這一位算「已滅」
1809  cmp  es:[di+196h], 8     ; GONE      │
1814  cmp  es:[di+196h], 7     ; STONED    │
181F  cmp  es:[di+196h], 2     ; TEMPGONE  ┘
1827  mov  [7F34h], 0          ; 都不是（含 0 OK、3 RUNNING、4 UNCONC、5 DYING）→ 不算全滅
183E  cmp  [7F34h], 0
1843  jz   184F                ; 非全滅 → 一般戰後清理（18F6h 直接跳 1963h）
1845  cmp  [0BDE8h], 0         ; ADUEL：決鬥輸掉不算全滅
184C  jmp  18F8                ; 全滅那一條
```

- **判準**：每一位的 `CHARSTATUS` 都在 `{1 ANIMATED, 2 TEMPGONE, 6 DEAD,
  7 STONED, 8 GONE}` 才算全滅（ordinal 見 spec 427）。
  **還有人昏迷（UNCONC）／瀕死（DYING）就不算**——隊伍帶著倒地的人回到地圖。
- 全滅那一條（`18F8h`）：戰後結果格 `+58Eh := 80h`、印全滅訊息、等鍵、
  `MUSICNO := 2` 換全滅曲（spec 1192）、`TTY := 1` 收尾。
  ⇒ **spec 1156 空著的那一格有了部分答案：`80h` 是 POSTCOM 全滅路寫的**。
- 非全滅那一條跳過訊息與換曲點，走一般清理（釋放兩條暫存鏈等）。

## 二、劇情傷害：ECL `2Eh DAMAGE` 收尾（能行動旗標判準）

DOS `overlay-02` `2BEEh..2C1Dh`（handler 收尾，spec 1152 的引文本輪一手重驗）：

```
2BEE  mov  [4FC7h], 1          ; 全滅 := true
      …走角色名冊（頭 650Ah、next +189h）…
2BFE  cmp  es:[di+196h], 0     ; DOS +196h ＝「站著且能行動」旗標（spec 1010）
2C04  jz   2C0B
2C06  mov  [4FC7h], 0          ; 有人能行動 → 不算全滅
2C1D  cmp  [4FC7h], 0
2C24  …印 'The entire party is killed!'（CS:2903h）、延遲…
```

- **判準**：每一位的「站著且能行動」旗標都是 0。這面旗標被 `KILLDUDE`
  （`overlay-23:0016h`）在**任何形式的放倒**時清 0（昏迷、瀕死、死亡都是），
  `STANDUP`（`251Ah`）站起來才設回 1。⇒ **昏迷也算倒下**，比 POSTCOM 嚴。
- `4FC7h` 是兩個主迴圈的收尾旗標（spec 1045／1095），`PROGRAM 3` 也設它
  （spec 1154）——三個來源、同一個全滅畫面。

## ⚠ offset 陷阱：`+196h` 在兩平台是不同欄位

DOS 角色記錄比 PC-98 少一格（`+14Ch` MONSTERTYPE，spec 1166），其後全部
offset 位移 1：

| 欄位 | DOS | PC-98 |
|---|---|---|
| `CHARSTATUS` | `+195h` | `+196h` |
| 站著且能行動旗標 | `+196h` | `+197h` |

spec 427（PC-98 的 `+196h`＝CHARSTATUS）與 spec 1010／台帳（DOS 的
`+196h`＝能行動旗標）**都對**；把其中一邊的 offset 直接搬到另一邊，會把
上面兩段迴圈讀成「全員 OK 才算全滅」這種自洽但荒謬的結論。
⇒ 規則（re-retro 62/63 同族）：**跨平台引用角色記錄 offset 前，先過
spec 1166 的位移表**；語意衝突時第一個檢查就是「是不是同一個 offset 指到
不同欄位」。

## 三、狀態 API（overlay-23，台帳既有筆記，本輪引用）

- `KILLDUDE(msg, new_state, char)`：`+196h`（PC-98）已在 `{6,7,8}` 不覆寫；
  否則設 new_state、能行動旗標 := 0、目前 HP := 0，REMOVEFX＋CHECKFX(0Dh)。
- `HEALDUDE(char, amount, only_if_hurt)`：狀態不在 `{0,1,4,5}` 回 false；
  治療瀕死（5）者會把狀態改成昏迷（4）。
- `STANDUP(hp, char)`：狀態 := 0、能行動旗標 := 1、HP := hp。

## remake 接線（本輪）

- `internal/game/state.go`：`postCombatPartyDead()`（POSTCOM 判準，remake 狀態
  集合取交集 `{Animated, Dead, Stoned}`——TEMPGONE／GONE 無對應機制）與
  `partyCannotAct()`（DAMAGE 判準，狀態在 `{OK, Animated}` 算能行動）。
- 戰後閘（`combat_state.go` finishCombat 之後）與**全滅那一首的換曲點**改用
  `postCombatPartyDead`——原作非全滅的敗戰跳過換曲點，打輸但有人只是昏迷
  不再進全滅畫面、不放全滅曲。
- ECL DAMAGE 的全滅先前**沒有接**：`resolvePendingECLDamage` 收尾補上
  `partyCannotAct` → `enterPartyKilled("DAMAGE")`；事件流在判定全滅後停止
  套用本次執行剩餘的訊號。
- 測試：`TestPartyWipeCriteriaFollowOriginalAsm`（逐格釘兩判準的差異）、
  `TestCombatDefeatWithUnconsciousMemberIsNotAWipe`、既有
  `TestResolvePendingECLDamageFinishesActiveCombatWhenPartyFalls` 改釘
  事件名 `DAMAGE`（昏迷的唯一成員：POSTCOM 不算、DAMAGE 算）。

## 明確不宣稱

- 傷害放倒的狀態階梯已由 [spec 1205](1205-savedamage-down-ladder.md) 收掉
  （`14Ah:0ACh` ＝ overlay-24 的 `SAVEDAMAGE`，戰鬥與 ECL 共用；remake 的
  戰鬥路徑同輪接上）。`postCombatPartyDead` 對狀態 OK 而 HP 歸零的成員仍
  退回 HP 判定——理由與純化條件見 1205 的「明確不宣稱」。
- 非全滅敗戰的清理細節（POSTCOM `184Fh..18F6h` 的兩條暫存鏈、
  `sub_1506`／`sub_A5C`／`sub_1072`）未逐支解讀；remake 的敗戰收尾不對照
  這幾支的內部行為。
- `ADUEL`（決鬥）在 remake 走不到（spec 1150），閘裡沒有對應項。
- DAMAGE 收尾判定全滅後，remake 停止套用該次執行剩餘的訊號；原作是跑完
  整段才由主迴圈收尾。玩家可見差異未量測。
