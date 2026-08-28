# 1244：正常入口採原版建角流程並加入繁中引導

狀態：`READY`
日期：2026-08-27

## 使用者決定

隊伍與人物建立要依原版流程，但可以加入不改變選項與結果的引導。正常玩家不再先看
remake 的 40 筆角色模板清單；模板 adapter 只保留既有測試與內部相容用途。

## 原版流程與實作

證據沿用 spec 1093 的 DOS `overlay-17:00782h`：

```text
種族 → 性別 → 職業 → 陣營 → 擲能力值／HP → 名字 → 是否儲存
```

- `OpenCharacterCreation` 現在直接啟動上述 guided state machine。
- 選完陣營立即完成第一次擲點；`R` 重擲、`Enter` 接受，不要求玩家猜第一步要按 R。
- 名字支援 Unicode；空名字仍依原版重問。
- `Y`／`Enter` 儲存，`N` 放棄；存完或放棄都回隊伍組裝頁，不再跳出建角。
- 隊伍組裝頁列出目前角色、種族、職業與人數；`G` 建立下一人、`D` 完成隊伍、
  `Esc` 取消。
- guided flow 使用獨立 `guidedReturnMode`，不再覆蓋整個角色建立畫面的
  `creationReturnMode`。

新增引導只說明目前步驟、方向鍵、Enter、Esc、R、Y／N、G、D；沒有新增種族、
職業、陣營、屬性調整或不同的儲存結果。

## 驗證

- `TestNormalCreationEntryUsesOriginalFlowAndReturnsToPartyAssembly` 固定正常入口與返回頁。
- `TestKeysDriveARealSessionFromTheTitle` 改由按鍵 seam 完整走過四段選單、自動擲點、
  中文姓名、儲存與完成隊伍，不再以「連按 Enter 六次加入模板」冒充正常流程。
- 文字輸入納入 `keySource.InputChars`，中文姓名可以由測試重播。
- Docker/Xvfb 視覺抽樣：種族頁與隊伍組裝頁均為 640×480，提示留在下方獨立石框，
  沒有壓到選項或名冊。

本規格不宣稱原版外層所有 disk／character-file 管理選單、刪除角色或匯入 Pool of
Radiance 已完整重建；本輪範圍是新隊伍的正常建角與重複加入角色。
