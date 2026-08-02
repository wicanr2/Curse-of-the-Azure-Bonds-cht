# 476：紮營與 ALTER 介面資料分離

狀態：`READY`

## 範圍

本輪處理作品 runtime 的 CAMP、REST、FIX、VIEW、MAGIC 與 ALTER 選單。規則、
選項 stable token、角色與法術資料仍留在既有 typed State；玩家可見繁中只由
`assets/locale/zh-TW.json` 提供，不再由 Go 的 `catalog.Text` fallback 保存第二份。

這是介面資料分離，不代表紮營、法術準備、原版存檔選單或 ALTER 的 DOS 畫面已
全部像素級還原。

## 實作與資料契約

- 128 次 Go 漢字 literal 已移除；`catalog.Text(key, key)` 在缺資料時顯示 stable
  ID，而不是靜默顯示舊 Go 譯文。
- 新增角色列、改名、頭身 block、移除錯誤、排序目的地與 Cure Light Wounds
  施法流程所缺的正式 locale keys。
- 刪除 `characterClassName` 的七筆中文 switch，改用訓練、建角與紮營共用的
  `localizedCharacterClassName` stable resolver。
- 修正先前被 Go fallback 掩蓋的正式資料漂移：REST 完成格式補回法術記憶人數；
  無 roster 的 SAVE／VIEW／MAGIC，以及 ALTER DROP／ICON 的訊息改為符合現行
  State gate 的內容。
- `TestCampAndAlterCatalogCoversEveryDisplayedStableID` 逐鍵載入正式繁中 JSON；
  CAMP／ALTER 行為測試也改用正式 catalog，不再依靠四鍵合成 fixture 的 fallback。

## 依賴 checkpoint 稽核

正式 gate 另發現 CoAB `go.mod` 鎖定 `f06493f`，但該 Engine 版本尚無目前 game
pack 已使用的 `CharacterCreation` schema，無法編譯。README 與 nested Engine
實際 HEAD 均為 `f3c652a3882e`；本輪將 module pseudo-version 修正為
`v0.0.0-20260802101532-f3c652a3882e`。

驗證使用 exact `f3c652a`、版本限定的暫存 `go.work replace`，不是以相容但不同的
Engine commit 冒充 pinned dependency。`f06493f` 的失敗編譯是依賴漂移證據，不能
寫成產品測試失敗，也不能忽略。

## 量測與驗證

- Go 漢字 baseline：`518 → 388`。
- `runtime_ui_debt 360 → 230`；`frontend_ui_debt 135`、
  `localization_debt 23` 不變。
- focused CAMP／ALTER 測試通過。
- exact Engine `f3c652a` 的 Docker／Xvfb／`--network none` 正式套件 gate：
  `go test ./cmd/... ./gamepack ./internal/...` 通過，marker
  `ROUND476_FORMAL_EXIT=0`。
- 本輪未修改 renderer，不新增或替換 README 截圖。

## 尚未完成

- frontend 與其他 State vertical slices 仍有 388 次既有漢字債務。
- CAMP SAVE 的完整原版選單、所有紮營法術、REST 隨機遭遇 fidelity、ALTER 的
  DOS 動態畫面及訊息速度 wall-clock 校準仍待完成。
- `AppendRenameName` 的 DOS 15-byte 錯誤屬技術 error；若未來直接顯示給玩家，
  必須另建 typed error→locale adapter，不能把錯誤字串翻譯塞回規則層。
