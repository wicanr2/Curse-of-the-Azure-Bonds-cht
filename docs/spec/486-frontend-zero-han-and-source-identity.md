# 第 486 輪：前端零漢字與來源身分控制流

狀態：`READY`

## 問題

Ebiten 前端最後二十六筆漢字包含 combat screenshot demo 名稱、世界地圖 preview
fallback，以及法師塔自動路徑直接比較「等待」「攻擊法師」「與龍群交涉」「狡猾」
與繁中劇情片段。只要 game-pack 改譯文，preview／驗證流程就可能選錯分支或失敗。

## READY 契約

1. 自動流程以 `currentOriginalChoices` 的 exact ECL source token 尋找選項：
   `WAIT`、`ATTACK WIZARD`、`PARLAY WITH THE DRAGONS`、`PARLAY_SLY`。
   `OriginalChoiceIndex` 不讀 `Choices` 顯示文字。
2. 劇情抵達判定以 stable game-pack message ID
   `wizard-tower.dragons-convinced` 解析目前 locale 文字；前端不保存中文 fragment。
3. combat visual demo 與 `-encounter` debug party 名稱使用十一個 stable locale IDs。
   Demo ID／戰鬥身分仍是英文 technical ID，不把譯名當 identity。
4. 世界地圖 deterministic preview 由 State `PrepareWorldMapPreview` 使用正式 catalog
   建立；renderer 不提供繁中 fallback。
5. 本輪不變更法師塔 ECL、選項順序、戰鬥、demo 數值、世界地圖 geometry 或
   screenshot 時間點。

## 驗證

- Source identity test 將顯示選項替換為 `TRANSLATION A/B/C`，仍能以原始 token
  找到正確 index；顯示文字本身不能被誤認 source token。
- Stable message ID 測試由正式 game-pack 解析目前 locale 全文，並驗證缺失 ID
  fail-closed。
- 十一個 demo name 與世界地圖 preview 全由正式 catalog 驗收，測試不複製中文。
- Go 漢字稽核 `238→212`；`frontend_ui_debt 26→0`、runtime 212、localization 0。
- `cmd/azure-bonds-game/main.go` 只剩註解中的非 ASCII 字元；AST gate 忽略註解。

這是「前端零硬編碼繁中」里程碑，不代表 runtime 212 筆債務、全遊戲翻譯、完整
可通關或 GUI 像素級終驗已完成，因此不新增 README 截圖。
