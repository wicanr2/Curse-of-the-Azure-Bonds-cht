# 第九十七輪：DOS player spell record parser

狀態：`READY`（限已公開記錄格式中的 spell 欄位）

## 已確認欄位

公開的 Curse of the Azure Bonds PC creature file format 同時適用於 saved-game character（`.SAV`）與 party 外角色（`.GUY`）：

- `0x01E–0x071`：memorized spell slots；`0` 是空 slot，其他值是 spell ID。
- `0x079–0x0DC`：known-spell flags；每個 byte 的 `0` 代表未知、非零代表已知，表格順序對應 one-based spell IDs。

本輪 parser 要求輸入已解壓的 record，最少 `0x0DD` bytes；不足即回傳錯誤。它只讀上述兩段，保留 slot 順序、移除空的 memorized slots，並將 known flags 轉成 one-based spell ID。未猜測 record 其他欄位，也未處理 DAX 壓縮或 saved-game container。

## Remake adapter

`party.Character.ApplyDOSSpellRecord` 將非空 memorized slots 複製到目前 data-neutral 的 `SpellSlots`。因此 `Roster.FindSpell`／`game.State.ResolveSpellSearch` 可以先使用真實 DOS spell slot 資料；原始 ECL memory writeback 與完整 character import 仍是後續工作。

## 驗證

`internal/party/character_test.go` 覆蓋：

- 空 slot 過濾與 memorized slot 順序。
- known flag → one-based spell ID。
- character adapter 與 truncated-record rejection。
