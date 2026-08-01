# 第三十一輪：AD&D combat core

狀態：`PARTIALLY SUPERSEDED`（攻擊核心仍有效；先攻已由 READY spec 419 取代）

> 本輪原先把先攻寫成 d20＋bonus 並以 ID 解 tie，那是早期 remake
> approximation，不是原版規則。第 419 輪已由 PC-98 primary bytes 證明
> `1d6 + DEX reaction adjustment`、Action.delay 與逐 TeamList 節點 d100
> selector；相關舊斷言不得再引用。

`internal/combat` 建立平台無關的戰鬥模型：

- `Fighter` 保存 party／enemy、HP、AC、攻擊加值與傷害；先攻欄位現況見
  spec 419。
- `StartRound` 的現行原版排程契約見 spec 419；本檔不再定義先攻公式。
- `ResolveAttack` 實作公開 reference 對齊的攻擊判定：天然 1 必 miss、天然 20 必 hit，其餘 d20＋attack bonus `>=` target AC 才命中。
- 命中後套用注入的 weapon damage total，更新 HP 與 PartyWon／EnemyWon／Draw 狀態。
- seed 與注入骰點使 regression 不依賴隨機環境。

本輪只完成規則核心，尚未從 ECL `LOAD MONSTER`／`SETUP MONSTER` 產生敵人，也尚未完成 combat map、移動、法術、物品、逃跑、PARLAY 或 Ebiten 戰鬥 UI。

驗證：`go test -vet=off ./internal/combat` 與全 repo Go packages。

- [x] 建立 party／enemy fighter model。
- [x] 建立 deterministic scheduler 基礎；原版公式與 tie order 由 spec 419
  supersede。
- [x] 建立 d20／AC／damage resolution。
- [ ] 接入 ECL encounter data。
- [ ] 接入戰鬥畫面與完整 AD&D action loop。
