# 第 446 輪：戰鬥動畫中段存檔與同幀續跑

狀態：`READY`

## 問題

第 443 輪的 remake save v7 已能保存 active combat，但 visual transaction
開始播放後仍會失敗即關閉，因為 elapsed 只存在 Ebiten frontend 的 wall clock。
這會讓玩家在箭矢、法術命中、死亡演出或 Sleep `TWINKLE` 期間無法存檔。

## 實作契約

- `State.combatVisualElapsed` 是目前 visual transaction 的權威時間位置；frontend
  只提供依 combat speed 換算後的單調遞增時間差。
- save v7 的 `CombatSnapshot.visual_elapsed_nanos` 保存精確 `time.Duration`；既有
  v7 存檔缺少欄位時視為零，與第 443 輪只允許第一幀前存檔的資料相容。
- 載入後 frontend 以「保存 elapsed＋新 wall-clock delta」續跑，Draw 直接使用
  State 已提交的位置，所以第一張載入畫面與存檔幀相同。
- `visual_travel_sent`、`visual_impact_sent`、`visual_death_sent` 一併 round-trip；
  同一幀重入及後續 handoff 不重播已發出的離散音效。
- elapsed 必須位於 `0..event.Duration()`；impact／death marker 必須位於
  `-1..ImpactCount-1`。沒有 visual event 卻帶非零 elapsed、時間倒退及越界資料
  一律失敗即關閉。

## 驗證

- 正常手動 Sleep 在 `700ms` 的 `TWINKLE` 中段存檔；全新 State 載入後
  `VisualFrame`、elapsed 與 Battle 相同，同幀重入及完成 handoff 都不重播
  `CAST／SPELLHIT`，其後自然到期與正傷害喚醒仍維持第 443 輪結果。
- 一擊致死弓箭在 death frame 存檔；載入後維持同一 death frame，已發出的
  `ARROW／HIT／DEAD` 不重播，完成 handoff 後仍是 party victory。
- frontend 純函式回歸證明 speed 0／4／9 都保留 saved base，再疊加各自縮放後
  的新 clock delta。
- 損壞 snapshot 的超時 elapsed 與越界 impact marker 均被拒絕。
- 正式 Docker／Xvfb／`--network none` gate：
  `go test ./cmd/... ./gamepack ./internal/...`，marker
  `ROUND446_FORMAL_EXIT=0`。

## 邊界

這只保存 renderer-neutral timeline 與「哪些離散 cue 已送出」。音效播放器內
正在播放的 PCM sample offset、PC-98 BGM 合成器／driver 狀態與原版 SAVGAM
戰鬥格式尚未保存，因此不能宣稱音訊取樣點無縫續播或原版存檔相容。
