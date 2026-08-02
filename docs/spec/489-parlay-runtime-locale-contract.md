# 第 489 輪：遭遇談判執行期文字資料契約

狀態：`READY`

## 範圍

本輪只移除 PARLAY 五種策略、策略提示與 generic 結果的 Go 中文 fallback；
不改 ECL encounter menu、劇情分支、怪物反應、Journal、戰鬥或 continuation。

## 身分與顯示契約

- 控制 identity 固定為 `PARLAY_HAUGHTY`、`PARLAY_SLY`、`PARLAY_MEEK`、
  `PARLAY_NICE`、`PARLAY_ABUSIVE`。
- 顯示由 `parlay_haughty`、`parlay_sly`、`parlay_meek`、`parlay_nice`、
  `parlay_abusive` 與 `parlay_menu_prompt` 解析；翻譯文字不決定分支。
- 無 ECL continuation 的 generic encounter 才使用 `encounter_parlay_done`；
  真實劇情仍由原 ECL tactic branch 產生 game-pack 訊息。

## 驗證

- 正式 catalog coverage 覆蓋七個 ID。
- generic encounter 將入口顯示 choice 改成 `DISPLAY`，仍由 `PARLAY` 進入五
  種策略；選 `PARLAY_MEEK` 後的期望全文在執行時由正式 catalog 組成。
- 法師塔長路徑同時驗證五個原始 tactic identity 與正式 catalog 顯示後，選
  `PARLAY_SLY` 繼續原劇情。
- Myth Drannor 羅剎妖居所正常 GEO 路徑同時驗證 identity／顯示，選
  `PARLAY_HAUGHTY` 後仍解鎖手札 57 並返回地城。
- 聚焦 `./internal/game` 與 Docker／Xvfb／`--network none` 正式全套 gate
  通過；漢字稽核 `164→157`。這不表示 encounter reaction 模型或完整遊戲
  已完成。
