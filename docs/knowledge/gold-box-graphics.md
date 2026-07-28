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

- `HEAD<area>`／`BODY<area>`：一般場景人物 layer；reference 使用 unmasked DAX
  pictures，BODY 相對 HEAD 向下 5 個 8-pixel rows，也就是 40 pixels。不可把
  reference 的 row 單位誤當成 pixel，否則頭像會被塞進胸口。

### CombatantKilled skull overlay

`seg001.Init` 建立 26 個 `combat_icons`，其中迴圈 `block_id=0x00..0x0B` 將
`COMSPR` normal block 載入 icon index `0x0D..0x18`，attack block 則使用
`block_id+0x80`；接著 index `0x19` 載入 `COMSPR 0x19`／`0x99`。因此 reference
`ovr033.CombatantKilled` 的兩個來源可確定為：

| reference call | DAX source | derived asset |
|---|---|---|
| `combat_icons[24].GetIcon(Attack, 0)` | `COMSPR.DAX` block `0x8B` | [`comspr-block-8B-item-00.png`](../../assets/sprites/comspr-block-8B-item-00.png) |
| `combat_icons[25].GetIcon(Normal, 0)` | `COMSPR.DAX` block `0x19` | [`comspr-block-19-item-00.png`](../../assets/sprites/comspr-block-19-item-00.png) |

`CombatantKilled` 以 9 次短 delay 交替繪製 attack／normal icon；目前 Ebiten 已沿用
這個 100ms 交替 contract；共用 `DeathOverlayFrame` 固定九次後結束 flash，renderer 再
依 `DownedCorpse` 顯示 corpse marker。這個 mapping
只描述來源與 renderer 資產，不把 skull 的閃爍狀態塞進共用 combat core。

原作另以同一 layer block 加 `0x80` 取得 attack state；direction `> 3` 使用水平翻轉版本。

SPRIT frame 的 `x/y` 是 icon canvas 內的 placement metadata；應在 indexed frame 載入後保留，繪製時再依 renderer scale 套用，不能把它誤當成地圖座標。

PIC／FINAL 的後續 frame 不是獨立 pixels，而是相對於第一幀的 packed-byte XOR；SPRIT 不能套用這個規則。共用 parser 必須以來源 family 明確選擇 full-frame 或 XOR-from-first mode。

ECL 的 `PICTURE` block 是事件層 request，不應在 VM 內直接依賴 Ebiten；目前 `RunResult.PictureRequested`／`PictureBlock` 保持 renderer-neutral，game state 再把它轉成可恢復的 localized event screen。

PICTURE 的 block threshold 是重要的 family dispatch：`block < 0x78` 使用 `PIC<area>.DAX` animation，`block >= 0x78` 使用 `BIGPIC<area>.DAX` 的 unmasked static picture。兩者不能只靠同一個 filename key 或同一種透明設定處理。

PICTURE 還有第二個 dispatch：Area2 `HeadBlockId == 0xFF` 才使用上述 PIC/BIGPIC；若 head block 存在，body block 由 PICTURE operand 提供，改用 `HEAD<area>`＋`BODY<area>` scene composite。

目前可重用的資料邊界是 Area2 raw record `0x5C2` → `area.State.HeadBlockID` → `game.State.SceneHeadBlock`；這讓 renderer 不需要直接讀 DOS record，也讓後續 Gold Box 遊戲可替換自己的 Area state codec。
ECL script 可能在 PICTURE 後立刻把執行期 mirror `0x7EE1` 清回 `0xFF`，因此 VM 的
PICTURE signal 必須同時保存 opcode 當下的 HeadBlockId；不能等整段 ECL 因 menu／EXIT
停止後才從共享 memory 重建 scene selector。

戰鬥圖示的方向與位置也應分開：direction 是 0–7 的 facing，tile position 是 combat map 座標，screen position 是 camera transform 後的結果。不要用 fighter list ordinal 取代真實 map position；目前 ordinal 只作 deterministic fallback。`CombatCamera` 以 active fighter 對齊 viewport，後續 Gold Box 遊戲可替換 viewport／scroll policy 而沿用資料與 renderer 分層。

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

`.FX` effects 應沿用 `monster.ParseAffects` 的固定 9-byte codec：kind、little-endian duration、strength 與四 bytes effect-specific data 都先保存；`DOSPlayerRecord.ApplyEffects`／`Character.ApplyDOSEffects` 不應在 parser 階段直接修改 fighter。各遊戲再用 `ChineseAffectName` 與 rules adapter 顯示／套用效果，因為同一 kind 的 duration、strength 可能依遊戲版本不同。

欄位注意：duration 是 bytes 1–2 的 little-endian minutes，strength 是 byte 3，`0xFF` strength 是 permanent；不要再把 byte 3 當 duration。`monster.AdvanceAffects` 只做時間消耗／移除，effect kind 的 AC、攻擊、HP、免疫與 status 仍應由各遊戲 rules adapter 決定。

角色檔案共用介面應先停在 sidecar bundle：必要 `.SAV/.GUY` record、optional `.FX`、optional `.SWG` 分別解析後再由 `party.ParseDOSPlayerFiles` 組合。DOS player `head_icon @ 0x141`、`weapon_icon @ 0x142`、`icon_id @ 0x143`、`icon_size @ 0x144` 先以 metadata 保存；不要把 `icon_id` 當成 CHEAD／CBODY block 或 slot ordinal。不要因為文件名 `SAVGAM?.DAT` 已知就假設其 header、slot、area pointer 或 memory address；container 必須有 sample bytes／反組譯證據後才接入。

`cmd/azure-bonds -import-character` 是 bundle 的可重現入口，輸入只讀、輸出目前 versioned party JSON；後續 Gold Box 遊戲只需替換 sidecar parser／class mapping，不應讓 CLI 直接解析未知的 save container。

`cmd/azure-bonds-game -dos-character-record` 直接重用同一個 `party.DOSPlayerFiles` boundary，將 imported HP、icon、equipment、effects 接進 startup state；它只建立單一角色，不能冒充完整 SAVGAM party／area restore。

effect projection 目前將 active `0x01` Bless（attack +1）、`0x02` Curse（attack -1）、`0x21` Blind（attack -4、AC +4）、`0x24` Bestow Curse（attack -4）與對 party 的 `0x31` friendly Prayer（attack +1）投影到 `party.Character.Fighter`。hostile Prayer、Protection、Mirror Image、Haste 與需要 target／phase／saving throw 的部分仍由後續 combat rules layer 處理；這個界線可沿用到其他 Gold Box 遊戲。

場所 service 也應保持資料中立：目前 `game.State` 的 `INN` 只做已由手冊確認的安全 HP restore，並同步 `partyRoster` 與 renderer fighter；Camp Menu 的 spell recovery、毒／疾病、守夜與中斷規則不可因客棧訊息而被假設完成。

商店共用 adapter 應把 offer price 與 stock 交給遊戲／script 層；`ITEMS` descriptor 只有 combat／usability 欄位，不能從 base item bytes 猜售價。`party.Character.BuyItem`／`SellItem`／`PayIdentifyFee` 目前只處理 party inventory、gold、readied lock、overflow 與已確認的 200 GP ID fee，後續 Gold Box 遊戲可重用同一 transaction boundary。

`game.State` 的 STORE UI 已保存原版七個 command 的繁中 mapping。後續遊戲可沿用 `shopMenu` state 與 return-to-place contract，只替換 stock、money pool、character selection 與 appraisal service；未知 command 不應直接 fall through 成一般 ECL event。

CAMP 也應採相同的資料／UI 分層：`CAMP` 先進入 command state，`REST` 呼叫遊戲專屬 safe-rest service，完成後回到 CAMP Menu，`EXIT` 才離開 menu。`SAVE`、`VIEW`、`MAGIC`、`ALTER`、`FIX` 應各自注入遊戲專屬 service；已解出的窄 boundary 可以先接入，但不要把不同 Gold Box 作品的存檔格式、法術恢復或角色修改規則硬編在共用 renderer。

640×480 remake 使用原 320×240 的 2× 邏輯空間：原始 24×24 combat icon 以
nearest-neighbor 放大為 48×48，WALLDEF／8X8D 與事件圖也只採整數倍率，不做
bilinear filtering。中文不是放大 DOS 8×8 字模，而是在放大後的 canvas 以
16×15 或 24×24 等高解析 Unicode glyph 重新排版；目前主 UI 使用 24px font，
事件 caption 每行最多 22 個 Unicode code point。素材層與文字層必須分離，
以便後續 Gold Box 作品沿用 pixel-art renderer 而替換 locale/font。

目前 `CAMP VIEW` 已形成可重用的只讀 adapter：以 roster 為 source，角色選單與摘要畫面分離，equipment label 經 base-item catalog 映射；查看本身不得改動 gold、treasure、equipment 或 spell state。後續 Gold Box 遊戲只需替換 roster codec 與名稱 catalog，即可沿用這個 UI/state boundary。

`CAMP MAGIC` 可沿用同一個 roster selector，但 spell layer 必須保持三段分離：DOS／save adapter 提供 ordered memorized slot IDs，catalog 提供名稱與 level，rules service 才處理 prepare／forget／cast／recovery。UI 顯示已保存的 ID 不代表已完成法術規則，也不應由 slot ordinal 推導 spell name。

CAMP SAVE 的跨遊戲介面也應是 intent signal：game state 只提供一次性 save request，platform adapter 再決定 configured path、container codec、atomic write 與錯誤呈現。remake party JSON 可作目前 fallback，但不能把它當成原版 SAVGAM slot／area 格式。

ALTER ORDER 的跨遊戲 boundary 應以 stable character ID 重排 roster，再由 combat projection 依 ID 重新建立 fighter order；不能只交換 renderer ordinal。這同時保留後續遊戲替換 formation／deployment rules 的空間，並讓 DROP、ICON 等具有不可逆或素材 side effect 的 command 各自接 adapter。

ALTER DROP 應沿用同一個 stable ID transaction：先在 UI 層二次確認，再由 party service 同步刪除 roster、combat projection 與 remake save snapshot。不可把取消、空 roster 或最後一名角色誤當成成功刪除；原版 save disk 的實體刪除則留在 container adapter。

PICS preference 應在 game state 與 renderer 之間保持兩個獨立布林：picture visibility 決定是否建立／繪製事件 picture，animation visibility 決定是否使用 frame timing；關閉動畫不應刪除 frame asset 或改變 PIC／SPRIT codec。

SPEED preference 應以 renderer-neutral 的 1–5 級保存，文字 adapter 再依 Unicode rune reveal；不要以 byte offset 切中文訊息，也不要把速度設定混入 PIC／SPRIT frame delay。這讓後續 Golden Box 遊戲可替換字型與 UI 而保留相同設定語意。

ICON selection 應只提供 manifest 已驗證的 layer pair，並以 character ID 將 roster metadata 投影到 combat fighter；不要把 block ordinal、weapon icon 或 `+0x80` attack variant 混為同一欄位。後續遊戲可替換 block family manifest，而沿用同一個 selection／projection contract。

`LOAD PIECES A,B,C` 的 selector 順序已由公開 CoAB reference 收斂：A/B/C 分別填入 WALLDEF symbol set 1/2/3；單一 WALLDEF record 的 8×8D block 使用 selector，多 record 使用 `selector*10 + recordIndex + 1`。這是素材載入邊界，不代表 WALLDEF table 已直接等於畫面座標；後續 renderer 仍需解碼 row／column 到牆面組合的規則。

WALLDEF graphic IDs 在載入後不是原始值直出：reference `Offset` 只平移 `>=0x2D` 的值，三個 dungeon symbol slot 的 global base 分別為 `0x2E`、`0x74`、`0xBA`。共用 `gfx.PieceSet.WallSymbol` 可沿用這個 cell-to-item lookup；area-map 的 `0x01` base 是另一個 8×8D set，不能混用。

3D wall layout 可跨 Gold Box 共用 `WallStamp` contract：reference 的十個 viewport metadata 是 `idxOffset=[0,2,6,10,22,38,54,110,132,154]`、`colCount=[1,1,1,3,2,2,7,2,2,1]`、`rowCount=[2,4,4,4,8,8,8,11,11,2]`。先由 wall type 分流到 WALLDEF slot／slice，再以 row-major stamp 輸出；方向遍歷與遮擋應留在作品 renderer adapter。

`Draw3dWorld` 的方向遍歷也可共用：`left=(partyDir+6)%8`、`behind=(partyDir+4)%8`、`right=(partyDir+2)%8`，從前方兩格依 Far→Mid→Near 消費 GEO wall。`gfx.WallLayoutCall` 保存 depth、map 座標、wall type、layout 與起始 cell；GEO wrap、sky／roof、door 與遮擋仍由各作品 Area／renderer adapter 決定。

dungeon preview 的 position service 可沿用 `geo.Grid.CanMove`：先驗證目前 cell 與相鄰 cell 的雙側 wall，再更新座標並重建 floor／wall view。這是安全的 renderer preview contract；正式遊戲仍需由 Area／save 注入 party position、direction、movement cost 與 encounter。

CAMP FIX 應拆成可重用的 healing service 與各遊戲的 spell catalog adapter：目前 CoAB 只能以已確認的一級牧師表順序將 `Cure Light Wounds` 映射到 one-based ID `3`，並以 memorized slot 數量決定 cast 次數。治療應以 roster 順序選擇受傷角色、以 `1d8` 封頂 MaxHP，再以 stable character ID 同步 combat projection；spell slot 是否消耗、重記憶、時間推進與中斷則由遊戲規則層注入，不能寫死在共用 renderer。測試可注入 seed 保持重現性，但正式遊戲仍需接原版 random／time source。

HEAD／BODY scene selectors 是兩個獨立欄位，不能假設同號。CoAB Tilverton Gond
祭壇實際使用 HEAD2 `0x09`＋BODY2 `0x06`。合成 canvas 高度至少是
`max(headHeight, 5+bodyHeight)`；BODY 放在 `y+5`，再以 HEAD 的非透明 pixels 覆蓋，
否則以 BODY 當固定 destination 會裁掉較高 layer，使用 OR merge 也會讓白色衣領吞掉臉。

shop stock／money pool 的共用邊界現在由 `game.ShopOffer`、`State.PoolPartyGold`、`TakeGold`、`ShareGold`、`BuyShopOffer` 提供。offer price 必須由各遊戲資料層注入；gold pool 的平均分配與 remainder 順序可直接沿用到其他 Gold Box 遊戲。

BUY renderer/state 已使用 `monster.ChineseName` 顯示 item offer；CityShop 購買會
clone item 且不移除 stock entry，inventory item 一律先保持未 ready。active character
selector 仍是明確 API，不可用 shop list ordinal 假裝完整 VIEW menu。

VIEW 目前只讀 party roster 並顯示 HP／gold／ChineseName equipment summary；不要在這個摘要層猜未識別 item、魔法 effect 或 ALTER menu side effects。

TAKE UI 目前使用 bounded 1／10／100／全部選項，底層 `State.TakeGold` 仍是任意 amount 的通用 contract；其他 Gold Box 可替換輸入 widget，不應把 bounded UI 當成原版數字格式的證據。

APPRAISE 以 `AppraisalOffers` 的 Ready flag／offer value 連接 gems／jewelry 到 money pool；沒有報價時 UI 不應把 treasure 數值直接當 GP，也不應省略拒絕報價的後續分支。

APPRAISE confirmation 現在分離 accept／reject／cancel；只有 accept 才呼叫 `AppraiseTreasure`。這個 transaction boundary 可跨 Gold Box 遊戲沿用，避免 UI 選取本身造成不可逆的 treasure mutation。

## 640×480 中文重製畫布

原版 320×200／320×240 級畫面適合 8px 拉丁字，但不足以容納可讀繁中。remake
固定使用 640×480 logical canvas，原始 tile、PIC 與 combat sprite 只採 2×／3×
nearest-neighbour 整數放大，保留清楚像素邊緣；中文字不從 DOS bitmap 拉伸，而以
16×15 compact tier 或 24×24 reading tier 直接重繪。圖片層與 Unicode 文字層必須
分離，才能讓後續 Gold Box 作品共用素材 decoder，同時各自調整中文行寬與 HUD。
640×480 是遊戲內部固定 logical resolution，不只是把 320×240 最終 framebuffer
平滑拉大：原素材在圖片層指定 `FilterNearest` 後做整數倍率 transform，24px／16px
CJK 字則直接 rasterize 到 640×480 文字層。如此原版小人保留硬邊像素，中文也不受
8px Latin cell 限制。

## 原版畫面拓撲

DOS Gold Box 的冒險畫面不是自由置中的圖片頁：上半部固定為左圖、右隊伍表，
下半部為敘事列，最底是 command line。戰鬥則改成左戰術地圖、右 active／target
狀態，下方訊息及 command line。640×480 中文 renderer 應保留這個拓撲，只放大
區域和文字容量；不能為了中文改成與原版不同的全畫面 card。

HEAD／BODY 的 reference `row + 5` 是五個 8px 顯示列，即 BODY y+40px。
合成順序為 HEAD 後 BODY，讓肩頸遮住頭部下緣；透明 index 0 不得清除底下 HEAD。
這個 row-to-pixel 換算可沿用於同系 Gold Box 的人物 scene layer。

戰術地圖必須是 clipping region。CPIC 大型怪物、CHEAD／CBODY 玩家、死亡 overlay
與 marker 都畫入同一 clipped target，避免 sprite 跨進右側狀態欄。combat terrain
selector 未證實前應顯示中性格線，不可把 TILES icon atlas 當成地板依序鋪設。

DOS 戰鬥畫面的可重用 native geometry 是 `8 + 168 + 8 + 128 + 8 = 320`：
168×168 戰場正好容納 7×7 個 24px movement cells，中央與外框各 8px，右欄
128px；底部 `y=184..199` 是兩列 8px 文字。2× renderer 對應戰場
`(16,16,336,336)`、中央框 `x=352..367`、右欄 `(368,16,256,336)`。
640×480 多出的 80px 可放中文戰鬥紀錄，但不能改變上方 320×184 的 2× 拓撲。
移動格是隱藏座標，畫面不應顯示 checkerboard；斜視感來自 terrain artwork。

戰鬥 terrain 不在一般 `TILES.DAX`：`DUNGCOM`／`WILDCOM`／`RANDCOM` 是共用
17-byte picture header 加多張 24×24 4bpp items，CoAB 分別有 25／34／6 張。
palette 0 是 overlay transparency。地城的 `BackgroundTile.TileIndex` 可直接查
DUNGCOM atlas；WILDCOM 與 RANDCOM 的選擇／擺放仍應由作品 engine adapter
提供，不能用 atlas index 當地圖順序。
