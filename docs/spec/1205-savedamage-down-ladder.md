# 1205：SAVEDAMAGE——傷害放倒的狀態階梯，戰鬥與 ECL 共用同一支（READY）

spec 1204 留下的問題：「戰鬥擊倒依什麼門檻選 昏迷／瀕死／死亡」。答案是
**只有一支**：`SAVEDAMAGE`（PC-98 `overlay-24:2658h`，Borland 符號直接叫
這個名字），戰鬥的 `PUTDAMAGE`（`overlay-23:2195h` 的 far call）與 ECL
`2Eh DAMAGE` 的傷害套用走的是同一個階梯——remake 的
`party.ApplyDamageWithHealthStatus` 早已逐條照抄它，本輪把**戰鬥路徑**也
接進來。

## 階梯（`overlay-24:2658h..2753h`，一手逐條）

```
SAVEDAMAGE(char, damage)          ; retf 6
剩餘 := 0; 溢出 := 0
if HP(+1A5h) >= damage: 剩餘 := HP − damage
else:                   溢出 := damage − HP
if 溢出 > 9 or (剩餘 = 0 and 狀態(+196h) = 1 ANIMATED):
    狀態 := 6 DEAD
elif 溢出 > 0:
    狀態 := 5 DYING
    if DS:7F27h = 5:  角色^[18Dh]^[0Eh] := 溢出      ; 出血量，戰鬥中才記
elif 剩餘 = 0:
    狀態 := 4 UNCONC
                                   ; 其餘：狀態不變
if 狀態 ∉ {0 OK, 1 ANIMATED}:      ; set 常數在 CS:2638h，bytes `03 00…`
    能行動旗標(+197h) := 0
    HP := 0
    if DS:7F27h = 5:
        dec byte [0A02Ah + 陣營(+198h)]              ; 該邊存活數
        戰鬥記錄(+18Eh)^[3] := 0
else:
    HP := 剩餘
```

- `>9` ⇔ HP ≤ −10 死亡、`1..9` 瀕死、剛好歸零昏迷；ANIMATED 倒下直接死
  （屍體不會再瀕死）。**與 `party.ApplyDamageWithHealthStatus` 逐條相同**
  ——該實作當初由 ECL 側解出，本檔補上「它就是共用常式」的證明。
- 出血量記在戰鬥記錄 `+0Eh`，包紮清掉（spec 696）；`DS:7F27h = 5` 是
  戰鬥模式旗標（PUTDAMAGE 同一格門）。
- 倒下收尾同時遞減 `0A02Ah + 陣營` 的**每邊存活數**——戰鬥勝負判定的
  資料源之一。

## ⚠ 解析 far call 目標時的 header 陷阱

`PUTDAMAGE` 呼叫的是 `14Ah:0ACh`。`14Ah` 是 TPOV **stub 段**的
program-relative paragraph；把它換算回檔案位置必須用**真正的 MZ header
大小**（`PC98-GAME.EXE` 是 **2,464 bytes**，不是常見的 128）：

```
stub 段檔案位置 = header(2464) + 14Ah×16 = 7,744 ⇒ overlay-24 的 stub 段
stub 0ACh → entry #28 → code offset 2658h ＝ SAVEDAMAGE
```

用 128 去算會把 `14Ah` 錯配到 overlay-13，落在一支 `retf 0Ah` 的函式上
——與呼叫端只推 3 個 word 對不上。⇒ 規則：**stub 段換算錯了，第一個
訊號是呼叫端與被呼叫端的引數字數不合**；對不上就回頭驗 header 大小，
不要硬解。（entry 對照的正解可用 stub 段內 `CD 3F <offset>` 的位元組
直接讀出，並拿 Borland 符號表同段的名字交叉核對。）

## remake 接線（本輪）

- `combat.Fighter` 新增 `DownOverkill`：`applyPositiveDamage`（battle 層
  唯一的扣血入口）在放倒那一擊記下 `傷害 − 當時 HP`；追擊不覆寫。
- `internal/game.syncCombatDownStatuses()` 把戰鬥擊倒投影成名冊狀態
  （同階梯；瀕死帶出血量），呼叫點：`finishCombat` 的全滅閘之前、
  `CombatCanBandage` 入口——戰鬥擊倒的瀕死者從此包紮得到。
- 測試：`TestApplyPositiveDamageRecordsDownOverkill`（溢出記錄、追擊不
  覆寫、歸零記 0）、`TestCombatDownsProjectSaveDamageLadder`（三種倒法
  各一、POSTCOM 閘不誤判、包紮可用、投影冪等）。

## 明確不宣稱

- `KILLDUDE` 的**非傷害**呼叫端（死亡魔法、石化等直接指定 new_state 的
  路徑）未盤點；本檔只涵蓋傷害放倒。
- 每邊存活數 `0A02Ah` 的所有讀取端未盤點；remake 的勝負判定走自己的
  fighter 掃描，未逐位對照。
- `postCombatPartyDead` 對「狀態 OK 而 HP 歸零」仍退回 HP 判定——階梯
  接上後正常路徑不會再出現這種成員，留著是為了不對沒鋪狀態的舊存檔／
  測試名冊誤判方向翻轉；純化要等存檔相容性盤點過。
