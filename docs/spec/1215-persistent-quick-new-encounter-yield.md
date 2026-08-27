# 1215 — 持久 QUICK 的新遭遇首幀交接

狀態：`READY`（限 `quick_fight` 跨戰保留與新遭遇的玩家中斷點）

## 原版證據與問題

spec 421 已由 PC-98 `overlay 8 local 1375h..140Fh` 證實，QUICK 寫在
`Player+199h`；Space 只清除 `ControlMorale < 80h` 的玩家角色旗標。因此 remake
不能在戰後一律清掉 QUICK，也不能把 NPC／臨時盟友改回手動。

remake 原本在 `StartCombat` 結尾同步呼叫 `advanceCombatToParty`。若上一戰留下的
可控制角色全是 QUICK，該函式找不到手動玩家回合，會在第一張戰鬥畫面交給玩家
之前一路跑到勝敗。一般強度按鍵路徑的皇家衛兵戰後緊接八名火刀，實測第二戰在
玩家沒有按鍵機會時直接全滅；這是前端交接缺口，不是敵人數量或敗北規則錯誤。

## 修正契約

- `StartCombat` 保留所有角色的 QUICK 位元。
- 新遭遇若至少一名存活、`ControlMorale < 80h` 的玩家角色帶著 QUICK，先在
  `ModeCombat` yield，不同步跑完整場。
- 玩家可在這個邊界按 Space；`CombatManualControl` 仍只清可控制玩家角色並同步
  回持久 party 投影。
- 沒有持久 QUICK 的一般開戰仍照既有流程推進到第一個玩家回合。

## 回歸與觀測

- `TestStartCombatYieldsBeforePersistentQuickCanResolveNewEncounter`：一名足以一擊
  殺敵的 QUICK 玩家進入新遭遇後，戰鬥仍為 `StatusActive`，且 Space 可收回控制。
- 既有手動控制與 ALT+M 每戰重設回歸繼續通過。
- 相同 1000 幀、相同 `route-clean-716.json`、未 boost 的路徑，第二場從首幀前
  同步結算改成可逐幀操作，擊倒數由一名增至三名；隊伍仍敗北。因此本輪不宣稱
  正常隊伍已通過連戰。
