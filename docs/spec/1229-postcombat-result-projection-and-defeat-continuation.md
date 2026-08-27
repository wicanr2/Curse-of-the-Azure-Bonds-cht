# 1229：POSTCOM 結果投影與非全滅敗戰續跑

狀態：`READY`（2026-08-27）

## 問題

remake 的 `finishCombat()` 只有隊伍獲勝時才回到 `COMBAT` 後方的 ECL 指令。
敵方獲勝但隊伍未全滅時，畫面回到地城，腳本卻被截斷；因此所有戰後敗戰分支
都走不到，包括散提爾堡守衛戰敗後押送法庭的 `ECL4/0x23`。

此外，remake 沒有把戰鬥結果寫入 `7EC7h`。spec 1156 已證明 POSTCOM 會寫
`0`、`80h` 或 `81h`，但當時刻意未把三值對應到 remake 狀態。

## 新的分支證據與勘誤

- `ECL4/0x20:07F0h` 與 `0966h` 都在 `COMBAT` 後比較 `7EC7h` 與 `80h`。
- 兩處只有 `7EC7h == 80h` 才跳到 `0423h`；該處顯示昏迷前文字、寫
  `4CE4=FFh`，再 `NEWECL 23h` 押送法庭。
- `7EC7h != 80h` 會離開該分支；實際讓隊伍走出戰場時，remake 狀態是
  `StatusPartyFled`，應投影為 `81h`，不會被守衛逮住。
- 這補足 spec 1156 當時缺少的區分：`81h` 的確屬於「沒有打贏」，但一般
  非全滅敗戰是 `80h`；不能把所有未勝結果都合併成 `81h`。

目前採用的結果契約：勝利 `0`、敵方獲勝 `80h`、隊伍逃跑 `81h`。全滅也在
POSTCOM 寫 `80h`，但依 spec 1204 走全滅畫面並終止，不再續跑 ECL。

## 實作

- `finishCombat()` 在收尾時把上述結果投影到目前 `BlockSession` 的 `7EC7h`。
- 隊伍獲勝、非全滅敗戰與逃跑都呼叫 `continueECLAfterEngineBoundary()`；只有
  全滅維持終止。
- 寶物處理仍只屬於勝利，不因共用續跑點而讓敗戰取得戰利品。
- `TestCombatDefeatWithUnconsciousMemberIsNotAWipe` 新增 `7EC7h == 80h` 回歸。

## 驗證與限制

Docker、無網路環境：

- `TestCombatDefeatWithUnconsciousMemberIsNotAWipe`：通過。
- `TestRealNewGameRunsToTheEnding`：完整主線與既有正常支線通過，312 句落回
  原文 0 句。

法庭／競技場仍未納入正常戰役覆蓋。原因不是 ECL 續跑：診斷覆蓋已看見
`07F0h`、`0966h` 的 `COMBAT` 邊界；而正常戰役測試角色有 9,161 HP，該
DUNGCOM 場景又把敵人放在玩家的逃跑邊界之外。守備會無限僵持，往敵人推進則
先成為 `StatusPartyFled`。這是戰場部署／測試角色強度的獨立缺口；在它修正前，
不得以直接改 HP、座標、`4CE4h` 或直接進 `ECL4/0x23` 冒充正常玩家路徑。
