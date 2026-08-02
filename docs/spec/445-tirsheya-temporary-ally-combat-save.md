# 第四百四十五輪：提爾雪雅臨時盟友戰鬥存檔

狀態：`READY`

## 目的

第 444 輪證明純隊伍對敵的 ECL combat handoff 可跨 save v7。這輪驗證更高風險
的混合陣營：ECL 將 monster slot 轉為 party-side `QuickFight` 臨時盟友；該
Fighter 必須留在 active Battle，卻絕不能寫入永久冒險 roster。

## 原始事件與玩家路徑

路徑是同一條 Standing Stone→Myth Drannor 長 regression，在紅網 save/load
後繼續：

1. 完成 Burial Glen，從東界選 `PATH` 進 ECL6／GEO6 block `42h`；
2. 由 exact spawn `(0,12,E)` 合法移動至 terrain `01h`；
3. 與提爾雪雅對話，`WAIT→YES`，解鎖 Journal 5；
4. 第一戰迎戰 HELL HOUND `44h`×5＋MARGOYLE `45h`×5；
5. 貝爾哈抵達後選 `BEYRHA`；
6. 原 ECL 以 `LOAD CHARACTER 8 → team 80h → ADD NPC 43h` 將第一隻
   RAKSHASA `43h` 投影為 party-side QuickFight 臨時盟友；
7. 第二戰對抗 RAKSHASA `43h`×1＋兩種隨從各六，共 12 名敵人。

上述 ECL／slot／team 語意沿用 READY spec 399 的 raw bytes 與 typed evidence。
本輪不重新命名未知欄位，也不在 frontend 依事件名稱特判。

## Save／Load transaction

第二戰進入正常 party-turn boundary 後：

- Battle party side 有英雄＋一名臨時羅剎妖，enemy side 有 12 名；
- 臨時盟友必須同時為 `TemporaryAlly=true`、`QuickFight=true`，名稱與
  `MON6 43h` 相符；
- `partyRoster` 仍只有英雄。

測試保存 `BattleSnapshot`、`SessionSnapshot`，寫入 version 7 temp save，再由
玩家自備 ECL image 建立全新 State，重掛真實 `MON6CHA／MON6SPC／ITEM6` 後
載入。restore 後逐欄驗證：

- Battle snapshot 完全相同；
- ECL session snapshot 完全相同；
- 臨時盟友完整 Fighter 完全相同；
- 永久 roster 仍只有英雄。

後續只使用 loaded state。戰勝後 ECL 寫 `4CD1=1`、返回 dungeon；
`PartyFighters()` 與 `partyRoster` 都只剩英雄，且英雄不帶 `TemporaryAlly`。
重踏 terrain 不重播事件，原長路徑繼續倉庫與後續區域。

## 證據等級

- encounter、monster IDs、slot 8、team `80h` 與 NPC add：`exact`，沿用 READY
  spec 399 的 ECL bytes／consumer evidence。
- save/load 前後 Battle、ECL、臨時盟友 equality：`proven`，本輪真實 campaign
  deterministic regression。
- 戰後 `4CD1=1`、runtime party／roster 無污染：`proven`。
- 高能力值英雄仍只用於長路徑加速，不支持原版 encounter balance 完成。
- 原版 SSI SAVGAM 是否保存此 mixed-team battle：`unknown`。

## 尚未完成

- 臨時盟友死亡、玩家戰敗、撤退與中途 cure／spell interruption 的 save corpus。
- final battle、persistent area、line spell、treasure service boundary 等其他
  active-combat 類型取樣。
- mid-animation／音訊 sample offset 與原版 SAVGAM combat layout。
