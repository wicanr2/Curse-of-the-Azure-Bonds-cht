# 第二百四十九輪：完整單職業角色建立選項

## 狀態

READY

## 證據與範圍

reference 的建立流程允許從種族／職業選擇開始；repo 已有 `raceAllowsClass` 的種族限制、
class minimum 驗證，以及依 `race_ages` 還原的 single-class starting-age table。本輪將
這些已驗證資料真正接到 remake 的建立選單，不再只提供三個固定人類／精靈模板。

目前列出 19 個可由現有 evidence 同時通過種族職業限制與起始年齡表的組合：人類 6、矮人
2、精靈 3、侏儒 2、半精靈 4、半身人 2。半獸人、多職業／dual-class 與完整原版建立
流程仍未臆測加入。

## 實作

- `starterCharacters` 產生所有 19 個單職業選項，使用各職業合法的 deterministic baseline abilities。
- `AddCreationCharacter` 延續既有 copied-character age roll／age effect transaction；最多六人隊伍不變。
- Ebiten 建立畫面以五列視窗顯示選項，游標上下移動時自動捲動；改名、能力編輯、重擲與完成按鍵維持原有流程。

## 驗證

- `TestCharacterCreationListsVerifiedSingleClassOptions` 驗證 19 個選項全部通過 `Character.Validate`
  與 `StartingAgeSpecFor`。
- `internal/game`、`internal/party`、`internal/area`、`internal/save`、`internal/ecl` 測試通過。
