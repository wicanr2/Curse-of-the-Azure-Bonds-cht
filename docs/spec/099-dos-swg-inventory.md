# 第九十九輪：DOS `.SWG` inventory 匯入

狀態：`READY`（限連續 item records 與 player equipment projection）

公開的 CoAB PC item format 將 character `.SWG`、placed treasure 與 monster item data 都表示為連續的 `0x3F` bytes item records。每筆 record 的已使用欄位由 `monster.ParseItems` 解碼：display name、item type、identified name components、plus、readied、cursed、weight、quantity、value 與三個 properties bytes。

本輪新增：

- `DOSPlayerRecord.ItemsPointer`（raw `0x14D–0x150` pointer）與 `EffectsPointer`（raw `0x0F2–0x0F5 pointer`）保存，但不自行解參照。
- `DOSPlayerRecord.ApplyInventory` 將外部 `.SWG` stream 解成 `Inventory`。
- `Character.ApplyDOSInventory` 將同一 stream 接到 remake `Equipment`。
- `DOSPlayerRecord.Character()` 將已附加的 inventory 帶入 party projection，因此原始 readied item 可繼續進入 `FighterWithEquipment`。

空 stream 仍合法（代表沒有 item）；非 `0x3F` 倍數立即錯誤。`.SWG` stream 的來源、pointer/address space、`.FX` effects 與完整 save slot container 尚未由本輪猜測補上。

格式參考：[Curse of the Azure Bonds creature/item file format](https://gamefaqs.gamespot.com/pc/564786-curse-of-the-azure-bonds/faqs/78365)。
