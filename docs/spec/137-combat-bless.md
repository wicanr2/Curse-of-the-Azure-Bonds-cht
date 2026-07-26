# 第一百三十七輪：combat Bless

## RuleBook 證據

本輪以本地 RuleBook 規則摘錄為依據：Bless 提升友方角色 THAC0 1；施法時鄰近怪物的角色不受影響。spell table 將 Bless 記為 `Both`、`6`、`5dia`、`6r`，本輪只接入已核對的一級戰鬥效果。

## 實作

- `BlessSpellID` 為 `1`。
- 牧師回合按 `B` 進入無目標施法確認；Enter 才消耗一個 memorized slot，Esc 不修改 slot。
- combat core 對未與存活怪物相鄰的存活隊友增加 `AttackBonus +1`，以 `Blessed` 狀態避免同一戰鬥重複疊加，並保存 `BlessRounds=6`。
- 每次開始新戰鬥回合遞減 Bless；第六個回合結束時撤回 `AttackBonus +1`。
- 訊息與控制列已加入繁中 `B：祝福`。

## 明確 boundary

當 party／enemy 都有 CombatMap position 時，以八方向相鄰（Chebyshev distance <= 1）排除鄰近隊友；缺少位置資料的低階 direct API 保留「無法判定則不排除」的相容 fallback。這不是完整的原版面積／地形規則，但已不再把所有隊友宣稱為固定合法目標。法術解除、重新施放與更完整 round boundary 仍屬後續 rules layer。

## 驗證

`internal/combat` 測試確認 Bless 只提升一次、排除相鄰隊友並在六次 `StartRound` 後解除；`internal/game` 測試確認 Begin／Cancel 不扣 slot、Enter confirmation 扣 slot 並同步隊伍攻擊加值。
