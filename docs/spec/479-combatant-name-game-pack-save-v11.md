# 第 479 輪：戰鬥者名稱資料化與存檔 v11

狀態：`READY`

## 問題與證據

- `internal/game/combat_state.go` 曾以 23 分支英文名稱 `switch` 直接回傳中文。
  原始身分來自 MON*CHA record，但翻譯同時散落在 Go fallback 與 UI locale。
- `BattleSnapshot` 保存 runtime `Fighter.Name`。該欄位已是翻譯後顯示值，因此
  繁中戰鬥中存檔、以英文讀檔時仍會顯示中文敵人名稱。
- 真實 ECL 長路徑已覆蓋 HIPPOGRIFF、BLACK DRAGON、DRACOLICH、三種
  ZHENTRIM、MOGION、CULTIST、SHAMBLING MOUND 與 BIT O' MOANDER；這些
  路徑可驗證資料搬移沒有改變 MON record、數量、sprite 或戰鬥 continuation。

## READY 契約

1. Engine game-pack 提供 `combatant_name_rules`：stable `id`、exact 原始
   `source`、跨 locale `message_id`。
2. Engine 驗證 rule ID／source 唯一與所有 locale 完整；partial source 不匹配，
   且不推論怪物類型、陣營、能力或 MON block。
3. CoAB 的 23 筆來源名稱與英／繁中顯示值只放 game-pack。State 找不到 rule
   時保留原始名稱，不得回退到中文 switch。
4. MON record 投影的 fighter 同時保存 `SourceName` 與目前 locale 的 `Name`。
   玩家角色沒有 MON source 時維持空 `SourceName`。
5. remake save v11 保存 active fighter `SourceName`；讀檔時依目前 game-pack／
   locale 重建 `Name`。舊 snapshot 沒有來源欄位時保留原值，不猜測反推。
6. 產品測試從正式 game-pack 解析期望名稱，不複製目前中文譯名。

## 驗證

- Engine `go test -count=1 ./...`：`ROUND479_ENGINE_FORMAL_EXIT=0`。
- CoAB game-pack 驗證 23 rules 在 en／zh-TW 均可 exact 解析，`HIPPO` 不會誤中
  `HIPPOGRIFF`。
- 真實 image 長路徑保留 8 鷹馬、3／14 黑龍、龍巫妖、1+8+2 散塔林部隊、
  摩貢儀式戰與三具摩安德殘軀，名稱期望均由 source 經 pack 解析。
- 合成 active-combat round-trip 以繁中建立 HIPPOGRIFF 戰鬥，英文 State 讀取
  save v11 後保留 `SourceName=HIPPOGRIFF`，顯示值重解為英文。
- Go 漢字稽核 368→345；`localization_debt` 23→0，frontend 133、runtime 212
  不變。

本輪不改 renderer 幾何與素材，因此不新增 README 截圖。原版忠實 UI 終驗與
README 過期圖汰換仍是完整功能收斂後的獨立門檻。
