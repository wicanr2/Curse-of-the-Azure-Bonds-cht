# 第 107 輪：DOS player `icon_id` preservation

狀態：`READY`

## 目的

補齊 DOS player record 的 combat icon metadata 邊界。`head_icon`、`weapon_icon`、
`icon_id`、`icon_size` 是不同欄位；`icon_id` 先保留為 renderer-neutral metadata，
不能直接當成 CHEAD／CBODY DAX block 或 party slot ordinal。

## 實作

- `ParseDOSPlayerRecord` 讀取 `icon_id @ 0x143`。
- `DOSPlayerRecord.Character` 將它保存到 `Character.IconID`。
- `Character.Fighter` 將它保存到 `combat.Fighter.PartyIconID`。
- 舊 party JSON 沒有此欄位時仍可載入；零值不被推測成另一個素材 block。

## 邊界

原始 runtime 的 icon slot allocation、recolor palette、角色編輯器選擇與真正的
CombatMap 座標／camera 仍需另外反組譯。這一輪只建立可跨 Golden Box 遊戲沿用的
欄位保存契約。
