# 454：神殿服務玩家文字與治療項目資料分離

狀態：`READY`

## 範圍

剛德神殿的十種治療不再於 `temple.go` 保存中文名稱。`templeCure` 只保存
stable locale key 與原規則價格；主選單、治療選單、確認、查看、集中／分配
金幣、估價、金幣不足及施術結果也全部由正式 locale JSON 解析。

效果 ID、骰數、死亡／石化狀態、詛咒裝備與 typed coin 扣款仍是規則層資料，
本輪沒有把機制搬進翻譯 JSON，也沒有把中文換成英文 Go fallback 來規避稽核。

## 測試與玩家路徑

- coverage 測試逐一驗證 19 個神殿 UI stable IDs 與十個 cure keys。
- Cure Light Wounds 確認選項由 `temple_confirm` 解析，不複製目前譯文。
- Remove Curse 仍驗證 effect `24h` 與 cursed equipment 的 typed 副作用。
- 真實 ECL／GEO 整合路徑從 terrain `92h`、PICTURE 6 經 engine service
  boundary 進入神殿，驗證 stable ID 顯示、治療、扣款、續跑及返回 `(0,7)`。

## 可量測結果

exact baseline 從第 453 輪的 1,251 降為 1,223 occurrences；28 個移除項目
全部位於 `internal/game/temple.go`，沒有新增漢字 literal：

| 分類 | 第 453 輪 | 本輪 | 差異 |
|---|---:|---:|---:|
| 本地化／劇情 fallback | 409 | 409 | 0 |
| Ebiten 前端 UI | 164 | 164 | 0 |
| runtime／工具 UI | 678 | 650 | -28 |
| 合計 | 1,251 | 1,223 | -28 |

這只證明目前神殿服務切片的玩家文字資料分離；神殿完整原版功能、所有角色選擇、
原版施術演出與全遊戲其餘 1,223 次債務仍未完成。
