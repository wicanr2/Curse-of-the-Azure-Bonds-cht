# 第 482 輪：玩家存讀檔與文字輸入 locale 契約

狀態：`READY`

## 問題

Ebiten 曾直接組合 F5／F9 存讀檔結果、音訊續跑錯誤、ALTER 改名欄與 ECL
`INPUT STRING` 欄位的繁中文字。這讓平台 adapter 同時持有狀態語意與翻譯，
也使 DOS 地城版與一般事件版的輸入提示形成兩份中文副本。

## READY 契約

1. `FileOperation` 表示 save、load、audio restore；`FileOperationResult` 表示
   succeeded／failed。State 依兩個 typed 值選 stable locale ID，再插入路徑或
   錯誤細節。
2. Audio restore 沒有成功訊息；若誤傳成功結果，仍解析同一失敗 ID，不讓
   renderer 猜測新的未定義狀態。
3. ALTER 改名與 ECL 字串輸入由 State 提供 value／help 文字。Ebiten 只選字型、
   座標與顏色；地城版和一般版共享同一 ECL contract。
4. 玩家輸入值、最大字數、路徑與技術錯誤仍是 runtime 動態資料，不寫入 locale。
5. 本輪不變更 save v11 格式、F5／F9 時機、音訊 snapshot、ECL continuation、
   rename 15-byte 限制、輸入按鍵或 UI 幾何。

## 驗證

- 正式 catalog 測試涵蓋 save／load 成功與失敗、audio restore 失敗、rename
  value／help、ECL value／help；期望值在測試執行時由 stable ID 解析。
- 既有 save round-trip、ECL `INPUT STRING`、ALTER rename 與 Ebiten tests 由
  正式 Docker／Xvfb gate 重跑。
- Go 漢字稽核 `313→300`；frontend `101→88`、runtime 212、localization 0
  不變。

本輪沒有版面、圖像或 renderer 幾何變更，因此不新增 README 截圖，也不擴大
宣稱完整原版 save 或完整 GUI fidelity。
