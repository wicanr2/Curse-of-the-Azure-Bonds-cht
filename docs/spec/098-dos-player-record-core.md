# 第九十八輪：DOS player record 核心角色匯入

狀態：`READY`（限公開 `.SAV`／`.GUY` fixed record 欄位與單職業 remake projection）

## 已確認欄位

依 [CoAB PC creature file format](https://gamefaqs.gamespot.com/pc/564786-curse-of-the-azure-bonds/faqs/78365) 的公開欄位表，`party.ParseDOSPlayerRecord` 讀取已解壓且至少 `0x1A6` bytes 的 record：

- `0x000–0x00F`：長度前綴與 ASCII 姓名。
- `0x010、0x012、0x014、0x016、0x018、0x01A`：六項原始能力值。
- `0x074–0x075`：種族與職業。
- `0x078`、`0x1A4`：最大／當前 HP。
- `0x079–0x0DC`、`0x01E–0x071`：known／memorized spell 欄位。
- `0x101–0x102`、`0x105–0x108`：gold、gems、jewelry。
- `0x109–0x10F`：各職業 current level；目前單職業 projection 使用對應欄位。
- `0x141–0x144`：玩家 icon head、weapon、size。

姓名長度、record 長度、種族、職業與 level 都做 bounded validation。多職業／druid／monk 等超出目前 `party.Character` enum 的 raw class 會拒絕，不會靜默映射成錯誤職業。

## Remake projection

`DOSPlayerRecord.Character()` 將核心欄位、HP、icon 與 memorized slots 投影到 `party.Character`，因此既有 `Fighter`／`FighterWithEquipment`、combat icon 與 `Roster.FindSpell` 可以使用匯入資料。若 `MaxHitPoints` 已有值，當前 HP 為 0 會保留為 0，支援死亡／失去意識角色。

本輪沒有解析 `0x14D` item pointer、`.SWG` inventory、`0x0F2` effects pointer、`.FX` effects、next-character chain 或 ECL memory writeback；完整 DOS save/import 仍未完成。

## 驗證

`internal/party/character_test.go` 使用 synthetic `0x1A6` record 覆蓋姓名、能力、種族／職業、等級、HP、gold、icon 與 spell projection，並保留 truncated／unsupported boundary tests。
