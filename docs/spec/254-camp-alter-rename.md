# 第二百五十四輪：CAMP ALTER RENAME

## 狀態

READY

## Contract

`CAMP → ALTER → RENAME` 會列出目前 party roster，選取角色後進入繁中 Unicode input
editor。名稱限制為最多 15 bytes，符合目前 DOS `.SAV/.GUY` name field writer；Enter
確認、Backspace 刪除、Esc 取消。

確認後只修改 `Character.Name` 與同 ID 的 combat fighter display name，不修改角色 ID、
能力、年齡、裝備、spell slots 或 `.FX/.SWG` sidecar。既有 `PatchDOSPlayerRecord` 會在
SAVGAM staged save 時把新名稱寫入 `0x00..0x0F`，未知 bytes 仍保留。

## 驗證與邊界

- State regression 覆蓋 ALTER menu、中文名稱 editor、roster／fighter projection 與返回
  ALTER menu。
- SAVGAM regression 覆蓋 rename 後 `.SAV` 名稱 writeback、角色 ID／basename 不變與 unknown
  byte preservation。
- 原版 rename error dialog、Big5／DOS code-page transcoding、多職業 serializer 與完整
  delete semantics 仍需後續反組譯證據；目前 UTF-8 remake UI 只允許能寫入既有 15-byte
  boundary 的名稱。
