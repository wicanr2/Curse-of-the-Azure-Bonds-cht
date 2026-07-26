# 第三十一輪：AD&D combat core

狀態：`READY`（限可注入骰點的核心規則）

`internal/combat` 建立平台無關的戰鬥模型：

- `Fighter` 保存 party／enemy、HP、AC、攻擊加值、傷害與 initiative 加值。
- `StartRound` 對存活角色擲 d20＋initiative bonus，按結果與 ID 決定順序。
- `ResolveAttack` 實作公開 reference 對齊的攻擊判定：天然 1 必 miss、天然 20 必 hit，其餘 d20＋attack bonus `>=` target AC 才命中。
- 命中後套用注入的 weapon damage total，更新 HP 與 PartyWon／EnemyWon／Draw 狀態。
- seed 與注入骰點使 regression 不依賴隨機環境。

本輪只完成規則核心，尚未從 ECL `LOAD MONSTER`／`SETUP MONSTER` 產生敵人，也尚未完成 combat map、移動、法術、物品、逃跑、PARLAY 或 Ebiten 戰鬥 UI。

驗證：`go test -vet=off ./internal/combat` 與全 repo Go packages。

- [x] 建立 party／enemy fighter model。
- [x] 建立 deterministic initiative。
- [x] 建立 d20／AC／damage resolution。
- [ ] 接入 ECL encounter data。
- [ ] 接入戰鬥畫面與完整 AD&D action loop。
