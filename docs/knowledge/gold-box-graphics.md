# Gold Box 圖像格式知識庫

## 共用處理順序

1. 從原始 ZIP 讀取 DAX member。
2. 用 DAX index 取得 block payload。
3. 依 picture header 解析 indexed pixels；SSI picture 的 width 是 `widthUnits × 8`，height 是 rows。
4. masked graphics 將 palette index `0` 映射成 renderer 內部的透明 sentinel `16`。
5. 單層 picture 轉 EGA16 RGBA；多層 icon 先在 indexed domain 合成，再轉 RGBA。

## 戰鬥小人分類

- `CPIC*`／`COMSPR`：可直接顯示的 combat icon／static sprite。
- `SPRIT*`：以 frame count、delay 與 picture frame 組成的動畫 stream。
- `CHEAD`：玩家頭部 layer，通常比 body canvas 矮，左上對齊合成。
- `CBODY`：玩家身體／武器 layer，通常提供 24×24 的 destination canvas。

- `HEAD<area>`／`BODY<area>`：一般場景人物 layer；reference 使用 unmasked DAX pictures，BODY 相對 HEAD 向下 5 rows。

原作另以同一 layer block 加 `0x80` 取得 attack state；direction `> 3` 使用水平翻轉版本。

SPRIT frame 的 `x/y` 是 icon canvas 內的 placement metadata；應在 indexed frame 載入後保留，繪製時再依 renderer scale 套用，不能把它誤當成地圖座標。

PIC／FINAL 的後續 frame 不是獨立 pixels，而是相對於第一幀的 packed-byte XOR；SPRIT 不能套用這個規則。共用 parser 必須以來源 family 明確選擇 full-frame 或 XOR-from-first mode。

ECL 的 `PICTURE` block 是事件層 request，不應在 VM 內直接依賴 Ebiten；目前 `RunResult.PictureRequested`／`PictureBlock` 保持 renderer-neutral，game state 再把它轉成可恢復的 localized event screen。

PICTURE 的 block threshold 是重要的 family dispatch：`block < 0x78` 使用 `PIC<area>.DAX` animation，`block >= 0x78` 使用 `BIGPIC<area>.DAX` 的 unmasked static picture。兩者不能只靠同一個 filename key 或同一種透明設定處理。

PICTURE 還有第二個 dispatch：Area2 `HeadBlockId == 0xFF` 才使用上述 PIC/BIGPIC；若 head block 存在，body block 由 PICTURE operand 提供，改用 `HEAD<area>`＋`BODY<area>` scene composite。

目前可重用的資料邊界是 Area2 raw record `0x5C2` → `area.State.HeadBlockID` → `game.State.SceneHeadBlock`；這讓 renderer 不需要直接讀 DOS record，也讓後續 Gold Box 遊戲可替換自己的 Area state codec。

戰鬥圖示的方向與位置也應分開：direction 是 0–7 的 facing，tile position 是 combat map 座標，screen position 是 camera transform 後的結果。不要用 fighter list ordinal 取代真實 map position；目前 ordinal 只作 deterministic fallback。

共用 fighter adapter 應保存 `pos` 與 `size`，而不是只保存最後的 screen pixel；這樣 camera、occupied tiles、碰撞與不同 Gold Box 戰場尺寸可以在上層替換。

目前已知的 reference position 生成公式可由 `combat.ReferencePlacement` 重用，但它依賴 team／candidate／occupancy 資料；在資料未解出前，不能把 deterministic formation 宣稱為原版 placement。

目前已確認 `PlaceCombatants` 的 team layout：party origin 固定為 `(0,0)`；enemy origin 為 `encounterDistance × MapDirectionDelta[mapDirection]`；party facing group 是 `mapDirection / 2`，enemy facing group 是 `((mapDirection + 4) % 8) / 2`。`combat.EncounterTeamStart` 已封裝這個不含 renderer 的轉換。`mapDirection` 與 occupancy table 尚未從 Area／Player record 接入，因此它只提供 team origin／facing，不代替候選格搜尋。

新建玩家的初始 icon 欄位不是依隊伍 slot 變化：`head_icon=0`、`weapon_icon=0`；種族只先決定 `icon_size`，small races 是 dwarf／gnome／halfling，其餘是 normal。

## 合成規則

對 destination `a` 與 source `b` 的每個 indexed pixel：

```text
a = transparent, b = transparent → transparent
a = transparent                  → b
b = transparent                  → a
a、b 都不透明                   → a OR b
```

這是目前從參考 Gold Box 重寫程式確認的 `MergeIcon` 語意。不要先把透明色畫成黑色，也不要用一般 alpha overwrite 取代 indexed merge。

## 跨遊戲重用介面

後續遊戲可沿用 `internal/dax`、`internal/gfx.Picture`、`gfx.MergePictures`、`gfx.MergePicturesAt`、`gfx.Picture.FlipHorizontal` 與 preview generator，只需建立每款遊戲的 DAX member 名稱、mask color、palette 與 icon layer mapping。

## 共用資料檔邊界

Gold Box 的 `ITEMS` 不是 DAX；目前 reference 格式是 2-byte header 加 16-byte descriptor table。descriptor index 就是 inventory item type，固定保存裝備欄位、hands、大小生物傷害、AC adjustment、weapon type、range、class usability mask 與 ammunition type。`monster.ParseBaseItems`／`BaseItemCatalog.Lookup` 可供後續遊戲沿用；若其他作品改變 descriptor 數量，仍應由 payload 長度計算，不要硬編碼 128。

裝備效果層應保持兩段式：`ItemRecord` 的 instance state（plus、readied、cursed、Affects）先查 `BaseItem` descriptor，再由上層決定是否投影到 fighter。現在 `ItemRecord.Effect`／`party.Character.FighterWithEquipment` 只實作 readied 基本武器／護甲，避免把 charges、magic effect 或 inventory mutation 混進共用 parser。

Party inventory transaction 可沿用 `party.ItemClassBit` 與 `Character.CanEquip`：class bit 來自原始 table，不可直接拿本地 enum 數字；slot 0/1 的 hands conflict 與 slot 9 的雙戒指限制也應留在資料層，renderer／商店只呼叫 `EquipItem`／`UnequipItem`。

Inventory mutation 也應以 instance state 為單位：`Count == 0` 是單件非堆疊物，正數才遞減 stack；readied 物品不可直接移除，cursed readied 物品不可卸下。`Character.RemoveItem` 提供這個跨遊戲可重用的安全邊界，消耗品真正 effect 仍由各遊戲 adapter 實作。

`Affects` 不是統一的 generic status：scroll 用 properties 保存 spell IDs，charged wand 用第一個 byte 保存 charges、第二個保存 effect ID。共用層應先回傳 `ConsumableUse` signal，再由各遊戲的 spell/effect engine 決定結果；不要在 parser 階段直接把 effect ID 當成傷害或狀態。

ECL `SPELL` 也應保持 signal boundary：它描述 spell ID 與兩個 runtime memory addresses，不等於已經找到施法者。`ecl.RunResult.SpellSearches` 可跨 Gold Box engine 重用，實際 party spell-slot lookup 再由各遊戲的 character record adapter 提供。

目前共用的 party adapter 是 `Character.SpellSlots`／`Roster.FindSpell`／`game.State.ResolveSpellSearch`；它用 ordered first-match 模擬 ECL search，但不取代原始 DOS player spell offsets。`ITEMS` catalog 則由 game bootstrap 載入，讓同一份 base descriptor 驅動 creation／save-load 的 equipment projection。

DOS player／creature record 的 spell 欄位可作為後續 Gold Box 遊戲共用的窄 adapter：公開 CoAB PC format 將 `0x01E–0x071` 定義為 memorized spell slots，`0x079–0x0DC` 定義為 one-based spell table 的 known flags。`party.ParseDOSPlayerSpellRecord` 只依這兩段資料工作，要求 decompressed record 至少 `0x0DD` bytes；不要在沒有版本證據時把其他 record offsets 猜成完整 character importer。`Character.ApplyDOSSpellRecord` 只替換 ordered non-empty `SpellSlots`，container／DAX／ECL writeback 由遊戲專屬層處理。

核心角色匯入可沿用同一個 record boundary：公開格式將姓名、六能力值、race/class、HP、各職業 current level、icon 與金幣放在固定 offset；`party.ParseDOSPlayerRecord` 目前只接受 remake 能表達的單職業 race/class，並把 `0x1A4` current HP 與 `0x078` max HP 保留到 `Character`。不要把 `0x14D` item pointer 或 `0x0F2` effects pointer 當成 inline data；它們必須分別透過 `.SWG`／`.FX` adapter 解析。

`.SWG` inventory 應沿用 `monster.ParseItems` 的固定 `0x3F` record codec，不要把 player `0x14D` pointer 當成 record 內嵌位置；`DOSPlayerRecord.ApplyInventory`／`Character.ApplyDOSInventory` 接受外部 stream，保留 pointer raw value，讓未來的 save/container loader 決定 address space。item instance 的 `Readied`、`Cursed`、`Count`、`Affects` 仍由 `Character` transaction／consumable adapter 處理，不要在 `.SWG` parser 裡直接套用戰鬥效果。
