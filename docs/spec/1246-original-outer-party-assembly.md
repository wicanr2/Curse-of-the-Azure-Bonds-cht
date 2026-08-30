# 1246：原版外層角色檔與隊伍組裝流程

狀態：`CONFORMED`
日期：2026-08-30

## 玩家可見範圍

正常新遊戲入口不再直接進入種族選單，也不在角色儲存後自動入隊。這一輪重建
原版的本作角色垂直鏈：

```text
功能選單 → 建立角色 → 儲存為獨立角色 → 回功能選單
          → 從 Curse 清單加入角色 → 回功能選單 → 出發
```

繁中介面保留目前步驟與按鍵提示，但選擇、返回點及交易結果遵循原版。

## RE 證據

所有位址均為原始 DOS 版 overlay 的段內位址；輸入雜湊與工具版本沿用
[`coab-function-index.md`](../audit/coab-function-index.md) 的固定清冊。

| 結論 | 原始定位 | 等級 | 出處 |
|---|---|---|---|
| 出發前外層是 `Choose a function`，建立與加入是兩個命令 | `overlay-17:001DDh` | `exact` | spec 1085 的 DOS 對側與 DOSBox 正常流程截圖 |
| 建角一次只產生一名角色，儲存後返回 caller | `overlay-17:00782h` | `exact` | spec 1093、1244 |
| 加入角色先問 `Add from where?`，`C` 代表本作 | `overlay-17:03680h` | `exact` | spec 1066 |
| 清單從角色檔讀名字，已入隊項目前置 `* ` | `overlay-16:0008Dh`、`overlay-17:03680h` | `exact` | spec 1066、1068 |
| 本作角色檔為 `.guy`、DOS 記錄 422 bytes | `overlay-16:00453h`、`0228Eh` | `exact` | spec 995、1038、1068 |
| 玩家隊伍上限八人；本作正常建隊保留兩格供劇情 NPC | `overlay-17:03680h` | `exact`／`strong inference` | spec 1066；六名玩家角色上限沿用現行已驗證流程 |

原版實機重播：標題 `Enter` → `CREATE NEW CHARACTER` → 完成建角與角色檔命名 →
返回 `CHOOSE A FUNCTION` → `ADD CHARACTER TO PARTY` → `ADD FROM WHERE? CURSE` →
挑選剛建立的角色。這是正常玩家路徑，不使用 `STING Wooden` 測試入口。

## Typed state 與交易

- `CreationOuterMenu`：顯示建立、加入與出發；空隊伍時出發失敗並留在原頁。
- `CreationSourceMenu`：顯示 Curse／Pool／Hillsfar／離開。本切片只有 Curse 可進清單；
  另外兩項明確顯示尚未提供相容匯入，不得靜默改讀本作 JSON。
- `CreationCharacterList`：列出獨立保存的本作角色；方向鍵移動、Enter／A 加入、
  Esc 返回。已在隊伍中的角色顯示 `*`，再次選擇不得重複加入。
- `ConfirmGuidedSave(true)`：將角色加入本作角色庫，而不是隊伍；寫入成功後才返回
  外層選單。寫入失敗時留在儲存頁並顯示錯誤，不得只存在記憶體。
- 角色庫採 remake 的版本化 JSON，與原版 `.guy/.swg/.fx` 分開，避免把現代 Unicode
  角色名稱或高階 typed 欄位冒充 DOS byte parity。
- `AddSavedCreationCharacter`：把角色庫中的角色複製入目前隊伍；同一角色 ID 不可重複，
  玩家建立角色上限六名。成功後清單項標成已加入。
- `FinishCharacterCreation`：只有隊伍非空才能開始冒險，並沿用既有正式 ECL 新遊戲入口。

角色庫預設與 remake 隊伍存檔放在同一資料目錄，檔名為
`azure-bonds-characters.json`；可用命令列參數指定，測試使用暫存目錄。

## UI 契約

- 角色管理是全寬清單，不套冒險畫面的 viewport／party HUD。
- 標題、選項、狀態訊息與底部操作提示各有獨立安全矩形。
- 選取反白不得改變字串換行；四語系使用同一幾何。
- 現代 theme 可以換框線與色彩，但不得改變選單順序、選取項或返回點。

## 驗收

1. State 測試：建立角色儲存後隊伍仍空、角色庫有一人；重新載入 State 後仍可列出。
2. State 測試：從 Curse 清單加入一次成功、第二次拒絕且標記維持。
3. State 測試：空隊伍不能出發；加入後由正式 `BeginAdventure` 進入 `(7,13)` 面東。
4. 前端按鍵測試：標題 → C 建角 → 儲存 → A → C → 選角色 → B，全程只走按鍵 seam。
5. Docker/Xvfb 抽拍外層選單與角色清單，確認文字安全矩形與繁中提示。

## 明確不處理

- Pool of Radiance 285-byte `.cha` 與 Hillsfar 188-byte `.sav` 的轉換器；選項保留但
  fail closed。這兩條需各自完成原始 fixture → parser → typed Character 的 READY spec。
- DOS 8.3 檔名、磁碟代號、實體 `.guy/.swg/.fx` 寫回；remake 不修改原版資料。
- 刪除、調整、訓練、轉職、移除角色等其他外層命令；它們不屬於本次新隊伍阻塞點。

## 實作與驗證結果

- `internal/save/character_library.go`：版本化 JSON 角色庫 codec。
- `internal/game/creation.go`：外層／來源／角色清單狀態、原子寫入、重複與空隊伍閘。
- `cmd/azure-bonds-game/main.go`：C／A／B 正常按鍵路徑與三個全寬頁面。
- State 測試通過獨立儲存、重載、加入及重複拒絕；真實按鍵測試由標題走到
  提爾佛頓 `(7,13)`，400 幀走過 47 格、英文回退 0、全滅 0。
- 四語系 640×480 外層頁已由 Docker/Xvfb 抽拍；第一版沿用冒險提示基線而被下框
  裁切，改為建角安全基線 `y=438` 後繁中實拍完整。
