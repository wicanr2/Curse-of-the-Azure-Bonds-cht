# 453：訓練場玩家文字與法術名稱資料分離

狀態：`READY`

## 範圍

本輪把訓練場作為第 452 輪精確基線建立後的第一個完整設施切片。訓練規則仍
由 Go 執行，但玩家看見的提示、選項、失敗原因、升級結果、職業與可學法術
名稱均由 `assets/locale/zh-TW.json` 的 stable ID 取得。

`trainingSpell` 現在只保存原版全域 spell ID、施法職業、法術等級與 locale
key；不再同時保存一份中文 `Name`。第一級魔法師法術沿用既有
`spell_magic_user_1..8` 真相來源，其他訓練法術各有 stable key。

## 測試契約

- 測試執行時載入正式 `zh-TW.json`，不複製目前譯名作期望值。
- catalog coverage 測試逐一驗證所有訓練 UI ID、職業 ID 與 spell key 存在。
- 魔法飛彈選擇測試由 `spell_magic_user_7` 解析期望文字；改譯名時不必改測試。
- 經驗不足測試由 `training_insufficient_experience` 解析期望文字。
- spell ID、升級資格、費用、HP 與職業上限規則仍以 typed assertions 驗證。

## 可量測結果

第 452 輪基線由 1,315 降為 1,251 occurrences，共移除 64 個正式 Go 漢字
literals；沒有新增、搬動或複製漢字 literal：

| 分類 | 第 452 輪 | 本輪 | 差異 |
|---|---:|---:|---:|
| 本地化／劇情 fallback | 409 | 409 | 0 |
| Ebiten 前端 UI | 164 | 164 | 0 |
| runtime／工具 UI | 742 | 678 | -64 |
| 合計 | 1,315 | 1,251 | -64 |

這只證明訓練場切片完成資料分離，不代表全遊戲已清除硬編碼。後續仍須用同一
流程遷移其他設施與章節，且不得以英文 fallback 取代 JSON 資料化來美化數字。

## 驗證邊界

聚焦測試涵蓋訓練、locale 與 sourceaudit；正式 Docker／Xvfb gate、正常設施
玩家路徑及容器清理結果記錄於 `CONTEXT.md`。訓練規則與 DOS 原版的完整
逐項反組譯並非本輪新增聲明。
