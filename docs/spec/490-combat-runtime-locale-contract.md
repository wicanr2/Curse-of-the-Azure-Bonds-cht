# 第 490 輪：戰鬥執行期文字資料契約

狀態：`READY`

## 範圍

本輪移除 `combat_state.go` 的玩家可見繁中副本，涵蓋戰鬥檢視、移動／防守／
反擊、近戰結果、法術結果、快速戰鬥、速度、延後／包紮、怪物施法、PICTURE
fallback 與勝敗訊息。沒有改動命中、傷害、豁免、AI、PRNG、visual timeline、
聲音、地形、ECL continuation 或 renderer geometry。

## 資料契約

- Go 只選 stable message ID 並傳入 fighter、HP、座標、傷害、次數等 runtime
  arguments；繁中格式只存在正式 locale JSON。
- `combat_view_*`、`combat_miss／hit／multi_attack`、`combat_victory／defeat／draw`、
  `combat_monster_held`、`combat_monster_magic_missile`、`combat_cloudkill` 與
  `event_picture` 補成正式 catalog key。
- 快速戰鬥未知法術名稱改走既有 `campSpellLabel`，不在戰鬥檔另造
  「法術 0xNN」副本。
- Lightning Bolt／Fireball／monster lightning 仍由 typed protected count
  決定 message ID；翻譯字串不參與規則分支。

## 驗證

- 正式 catalog coverage 覆蓋本輪所有戰鬥 stable ID。
- 歷史測試中的 held、敵方多段攻擊、怪物 Magic Missile 與角色檢視改載正式
  catalog，期望全文於執行時由同一 ID 與 typed arguments 組成。
- 斷網 Docker 抽樣 gate：`go test -count=1 ./internal/combat ./internal/game`；
  包含既有近戰、移動、法術、快速戰鬥、勝敗與多條正常 ECL 戰鬥玩家路徑。
- marker：`ROUND490_SAMPLED_EXIT=0`；漢字稽核 `157→94`，移除 63 筆
  `runtime_ui_debt`。

這只證明戰鬥文字資料邊界改善，不證明完整戰鬥規則、動畫、音效或全遊戲完成。
