# 1213 — QUICK 法術資料失敗的控制交接

狀態：`READY`（缺少 `combat_ai_spells` metadata 時的玩家控制交接；不宣稱一般強度已通過八名火刀連戰）

## 問題

五級角色可合法記憶非戰鬥或尚未接入 QUICK selector 的法術。`Alt+M` 開啟後，
selector 會掃描完整 `SpellSlots`；只要其中一格沒有 `combat_ai_spells` metadata，
remake 便顯示「已收回玩家角色控制」。舊實作只清除 `QuickFight`，卻保留
`combatQuickMagic=true`。玩家再次按 `Q` 時，同一角色會在同一格法術上得到同一個
錯誤，戰鬥不消耗回合、HP 與敵人數都不變，形成可重現的死環。

正常按鍵長跑把五級法師的四支已知法術全準備後，於第一場皇家衛兵戰從第 343 幀
開始重複此訊息；到第 1000 幀仍停在同一場戰鬥，證明這不是單純提示文字。

## 修正

`tryQuickSpell` 的玩家角色 metadata 失敗路徑現在同時：

1. 以既有 `SetPlayerCharactersManual` 收回可控制角色；
2. 同步 roster 的 QUICK 狀態；
3. 關閉本場 `Alt+M` 快速施法 gate；
4. 保留原本的記憶法術槽與明確錯誤訊息。

因此玩家可改用手動施法，或再次按 `Q` 只使用一般 QUICK 近戰；不會因同一筆缺少
metadata 的法術永久停住。這不替偵測魔法等法術偽造戰鬥行為，也不刪除合法記憶槽。

## 驗證

- `TestCombatAltMMissingMetadataReturnsManualAndDisablesQuickMagic`：使用真實 spell ID
  `11`（法師偵測魔法），鎖定手動控制、`Alt+M` 關閉與 slot 完整保留。
- `TestCombatAltMEnablesQuickMagicMissileFromGlobalSpellSlot`：正對照證明具 metadata
  的 Magic Missile 仍由 QUICK 施放並消耗一格。
- `go test ./internal/game ./cmd/azure-bonds-game -count=1`：通過。

## 一般強度現況

正常路徑現以正式 MEMORIZE 準備五支已具戰鬥 metadata 的牧師法術，法師準備
燃燒之手，並在每場戰鬥以正式 `Alt+M` 後交給 `Q`。1000 幀控制組仍在八名火刀
連戰後六人 HP 歸零；這個戰力／戰術 gate 尚未完成，不可由 ECL 0x04 或財寶選單
外觀宣稱勝利。
