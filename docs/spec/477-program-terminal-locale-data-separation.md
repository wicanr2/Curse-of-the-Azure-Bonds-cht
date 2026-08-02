# 477：PROGRAM 終局介面資料分離

狀態：`READY`

## 範圍與既有 oracle

本輪資料化 ECL external `PROGRAM 0／3／8` 的玩家介面：返回主選單、隊伍全滅、
提朗瑟克斯終戰勝利、保存勝利進度與不保存結束。PROGRAM ID、角色恢復、
`GameWon／PartyKilled`、save request 與 session 清理仍是 typed State 行為；只有
prompt、choice 與 message 由正式繁中 locale stable IDs 提供。

原始 PROGRAM 8 caller、37 名敵軍與正常 GEO／戰鬥 continuation 已由 READY
spec 408 證明。本輪不重新推論 executable 語意，也不把終局文字放入共用 Engine。

## 資料契約

新增十個 stable IDs：

- `program_main_menu`
- `program_party_killed_prompt／message`
- `program_return_title`
- `program_victory_prompt／message／save／save_requested`
- `program_end_without_save`
- `program_adventure_ended`

State 使用 ID 作明確 fallback；正式 JSON 缺鍵時顯示 ID，不再由 Go 中文掩蓋。
`currentOriginalChoices` 仍保存 `PROGRAM_WIN_SAVE／PROGRAM_END`，因此翻譯改字不會
改變 transaction。

## 測試與 save／restore 發現

- PROGRAM 0 驗證 session／block 清除及正式主選單訊息。
- PROGRAM 3 驗證全滅 prompt、返回標題 choice、message 與結束後標題訊息。
- PROGRAM 8 分別驗證全隊恢復、保存後發出一次 save request，以及不保存時不發
  request；兩條都回到標題。
- `TestProgramCatalogCoversEveryDisplayedStableID` 直接載入正式繁中 JSON。
- Standing Stone 至 Burial Glen／內部遺跡的長路徑以正式 scheduler 打敗 37 名
  敵人，並比較正式 PROGRAM 8 prompt／兩選項／message。

長路徑第一次加入 locale 斷言時顯示 stable IDs。根因不是 runtime 清掉 catalog，
而是兩次 save／restore harness 以 `testCatalog()` 重建 State；save 正確地不序列化
應用程式 locale。測試現於初始、一般 restore 與臨時盟友 restore 都注入同一正式
catalog，再完成終戰。不得在終局局部重新塞中文修補。

## 量測與驗證

- Go 漢字 baseline：`388 → 378`。
- `runtime_ui_debt 230 → 220`；frontend 135、localization 23 不變。
- focused PROGRAM 測試與正常 37 人終戰長路徑通過。
- Docker／Xvfb／`--network none` 正式 gate marker 記於 `CONTEXT.md`。
- 本輪沒有 renderer 變更，不新增 README 截圖。

## 尚未完成

- 由全新角色從開場走完全部章節的單一通關 gate。
- 原版 PROGRAM 3／8 終止畫面的像素、輸入節奏與 save 檔案選單 fidelity。
- 終戰敵軍完整 AI、所有法術／特殊能力、死亡演出與聲音。
