# 第一千二百四十六輪：正常角色 VIEW 版面勘誤與 polish

狀態：`READY`

## 原版證據與勘誤

2026-08-27 以現有 DOS 原版 A 隊伍，依標題 `LOAD → A → BEGIN → ADVENTURE → VIEW`
正常操作取得 [`character-view.png`](../reference/original-dos/character-view.png)，
SHA-256 為 `f8cb14f024fddf1db36064e1786dc2b975a3a702b5ced33d4b3f13a95a87671a`。

原版 VIEW 是全頁角色資料表，顯示姓名、性別／種族／年齡、陣營／職業、六項能力、
金錢、等級／經驗、AC／THAC0／負重、HP／傷害／移動力與狀態；底部命令是
`TRADE／DROP／EXIT`。它沒有 HEAD／BODY 立繪。先前 spec 391 與 406 把一張 NPC
冒險訊息圖誤認為角色資訊頁，本輪已刪除該錯誤外推，保留那張圖對 NPC 舞台仍有效
的證據。

## Remake 實作

- `State.CampViewCharacter` 只在正常 `CAMP → VIEW → 角色` 狀態公開 roster 複本，
  renderer 不能繞過選單或回寫隊伍。
- 前端以全頁數值表呈現原版可見欄位；下半頁保留 remake 已有的裝備／卸下操作，
  這是明示擴充，不冒稱原版 `TRADE／DROP` 行為逐項一致。
- `-character-view` 只建立 typed fixture，之後呼叫公開 `Camp()` 並經 `Select`
  進入 VIEW；它是可重生的 renderer checkpoint，不是完整新遊戲通關證據。
- `PrepareWorldMapPreview` 同步重設英文原選項，避免先前 ECL 選單殘值讓畫面上
  的 `CAMP` 被路由到錯誤事件。

## 驗證

- [`character-view-remake.png`](../screenshots/character-view-remake.png)，640×480，
  SHA-256 `f753146d827bace5a4c3594dd01ba8a7f97859ac1a17060456f2d07be3820a1c`。
- 人工檢視確認主資料區沒有 HEAD／BODY、所有文字留在石框內，下半頁裝備選項不再
  壓住分隔框。
- `go test ./cmd/azure-bonds-game ./internal/game ./internal/localeaudit` 通過。
- `CampViewCharacter` 測試確認狀態進入／離開與回傳複本不可回寫。

證據等級：原版頁面構成為 `exact` runtime observation；remake 幾何與繁中重排為
`material-exact/layout-reconstructed`。原版 `TRADE／DROP` 的完整交易語意不在本輪
完成範圍。
