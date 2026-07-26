# 格式規格

相關中文資料：[繁中遊玩手冊](../manual/curse-of-the-azure-bonds-zh-TW.md) ・ [中文金盒子歷史筆記](../history.md)

本目錄只收錄由原始映像、執行觀察、反組譯或可重現工具支持的規格。

狀態定義：

- `DRAFT`：有初步觀察，但仍有未驗證的關鍵假設。
- `READY`：可供實作，欄位、邊界、錯誤行為與驗證案例已足夠明確。

目前規格：

- [第一輪素材與格式盤點](./round-01-inventory.md)（`DRAFT`）
- [第二輪 DAX 容器與 ECL 文字取樣](./round-02-dax-container.md)（`DRAFT`）
- [第三輪 DOS loader 與 GAME.OVR](./round-03-loader-and-overlay.md)（`DRAFT`）
- [第四輪 Go 核心層與繁中資源](./round-04-go-core.md)（`READY`：限 DAX container／locale）
- [第五輪 ECL operand framing](./round-05-ecl-operands.md)（`DRAFT`）
- [第六輪 ECL 安全 trace walker](./round-06-ecl-trace.md)（`DRAFT`）
- [第七輪 ECL packed text extraction](./round-07-ecl-text.md)（`DRAFT`）
- [第八輪繁中開場狀態核心](./round-08-localized-state.md)（`DRAFT`）
- [第九輪 Ebiten opening prototype](./round-09-ebiten-opening.md)（`DRAFT`）
- [第十輪 ECL data-driven opening](./round-10-data-driven-opening.md)（`DRAFT`）
- [第十一輪 ECL branch target graph](./round-11-ecl-branch-graph.md)（`DRAFT`）
- [第二十六輪 ECL5 real NEWECL regression](./round-26-ecl5-newecl-regression.md)（`READY`：限已定位 transition entry）
- [第二十七輪 Shadowdale 荒野入口](./round-27-shadowdale-map-entry.md)（`READY`：限入口狀態與輸入 contract）
- [第二十八輪 Shadowdale 場所 menu](./round-28-shadowdale-place-menu.md)（`READY`：限選項與 state-level event contract）
- [第二十九輪 TREASURE bounded prefix](./round-29-treasure-prefix.md)（`READY`：限 operand framing 與安全前綴）
- [第三十輪 COMBAT request signal](./round-30-combat-request.md)（`READY`：限控制轉移 signal）
- [第三十一輪 AD&D combat core](./round-31-combat-core.md)（`READY`：限可注入骰點的核心規則）
- [第三十二輪 ECL monster spawn descriptors](./round-32-ecl-monster-spawn.md)（`READY`：限 ECL command descriptor）
- [第三十三輪 MON*CHA monster record](./round-33-monster-cha-record.md)（`READY`：限固定 record offsets 與 raw combat fields）
- [第三十四輪 MON*ITM／MON*SPC records](./round-34-monster-items-affects.md)（`READY`：限 raw record parser）
- [第三十五輪 monster item／effect 繁中名稱](./round-35-monster-item-localization.md)（`READY`：限本輪實際觀察 IDs）
- [第三十六輪 ECL-to-enemy encounter adapter](./round-36-ecl-encounter-adapter.md)（`READY`：限 enemy fighter 建立）
- [第三十七輪 可操作戰鬥狀態與 Ebiten 畫面](./round-37-playable-combat-state.md)（`READY`：限戰鬥垂直切片）
- [第三十八輪 ECL encounter 到 Battle 的資料橋](./round-38-ecl-encounter-to-battle.md)（`READY`：限 ECL1 direct-entry）
- [第三十九輪 PROGRAM 外部 routine 邊界](./round-39-program-boundary.md)（`READY`：限 VM 控制轉移）
- [第四十輪 遊戲內冒險手札與 CAMP state](./round-40-journal-and-camp-state.md)（`READY`：限資料呈現與控制邊界）
- [第四十一輪 可翻頁的繁中冒險手札](./round-41-journal-pages.md)（`READY`：限八頁摘要與 UI 導航）
- [第四十二輪 party 保存與 CAMP 恢復 state](./round-42-party-camp-state.md)（`READY`：限 party HP boundary）
- [第四十三輪 角色建立規則核心](./round-43-character-creation-rules.md)（`READY`：限 validation）
- [第四十四輪 繁中角色建立 UI](./round-44-character-creation-ui.md)（`READY`：限 starter slice）
- [第四十五輪 繁中角色姓名輸入](./round-45-character-name-input.md)（`READY`：限 Unicode 姓名）
- [第四十六輪 角色建立能力值編輯](./round-46-ability-editor.md)（`READY`：限能力值 slice）
- [第四十七輪 能力值 3d6 擲骰與重擲](./round-47-ability-rolls.md)（`READY`：限可重現擲骰）
- [第四十八輪 版本化 party JSON 存檔](./round-48-versioned-party-save.md)（`READY`：限 remake party descriptor）
- [第四十九輪 ECL COMBAT 到可操作 Battle](./round-49-ecl-combat-state-bridge.md)（`READY`：限已載入 party／MON*CHA bridge）
- [第五十輪 真實 ECL1 路徑戰鬥 regression](./round-50-real-ecl-combat-regression.md)（`READY`：限 block 0x51 journey slice）
- [第五十一輪 ECL RANDOM 與可重現 seed](./round-51-ecl-random.md)（`READY`：限 bounded VM random）
- [第五十二輪 ECL ENCOUNTER MENU](./round-52-encounter-menu.md)（`READY`：限 bounded encounter-menu bridge）
- [第五十三輪可恢復的 ECL runtime](./round-53-resumable-ecl-runtime.md)（`READY`：限 menu pause／resume execution context）
- [第五十四輪跨 ECL block runtime context](./round-54-cross-ecl-runtime-context.md)（`READY`：限 `NEWECL` context transfer）
- [第五十五輪全 ECL global block namespace](./round-55-global-ecl-loader.md)（`READY`：限全檔 loader 與 real transition）
- [第五十六輪 GEO 16×16 map geometry](./round-56-geo-map-geometry.md)（`READY`：限 GEO geometry parser）
- [第五十七輪 indexed picture 與 WALLDEF](./round-57-indexed-picture-and-walldef.md)（`READY`：限 picture／wall data parser）
- [第五十八輪 EGA indexed pixels 到 RGBA](./round-58-ega-rgba-adapter.md)（`READY`：限 palette／透明色 adapter）
- [第五十九輪 Ebiten 原始 tile gallery](./round-59-ebiten-tile-gallery.md)（`READY`：限 TILES graphics pipeline preview）
- [第六十輪 GEO geometry Ebiten viewport](./round-60-geo-ebiten-viewport.md)（`READY`：限 GEO raw geometry preview）
- [第六十一輪 GEO wall navigation contract](./round-61-geo-wall-navigation.md)（`READY`：限 raw GEO wall traversal）
- [第六十二輪 README 與可重現截圖](./round-62-readme-screenshots.md)（`READY`：限 repository progress evidence）
- [第六十三輪 wilderness floor construction](./round-63-wilderness-floor.md)（`READY`：限 reference wilderness floor generation 與目前 map slice）
- [第六十四輪 GEO dungeon floor composition](./round-64-dungeon-floor.md)（`READY`：限四段 tile composition 與 GEO2 preview）
- [第六十六輪原始 GEO map catalog](./round-66-geo-catalog.md)（`READY`：限 GEO DAX map ID catalog 與 preview selector）
- [第六十七輪 ECL LOAD FILES → GEO map request](./round-67-ecl-map-load.md)（`READY`：限第三 operand selector 與 renderer bridge）
- [第六十八輪 Area1／Area2 map-load boundary](./round-68-area-state.md)（`READY`：限 area state 與 `CMD_LoadFiles` branch contract）
- [第六十九輪 Area1／Area2 binary codec](./round-69-area-codec.md)（`READY`：限已定位欄位與未知 bytes preservation）
- [第七十輪可恢復的 remake game save](./round-70-resumable-game-save.md)（`READY`：限目前 game state 與舊 party JSON 相容）
- [第七十一輪三城市 location routing](./round-71-city-location-routing.md)（`READY`：限 opening city menu 與共同 map/place contract）
- [第七十二輪戰鬥小人 sprite extraction](./round-72-combat-sprite-extraction.md)（`READY`：限 CPIC masked picture extraction）
- [第七十三輪 Ebiten combat sprite renderer](./round-73-combat-sprite-renderer.md)（`READY`：限 CPIC asset mapping 與目前 playable combat slice）
- [第七十四輪 SPRIT animation codec](./round-74-sprit-animation.md)（`READY`：限 SPRIT frame stream decode）
- [第七十五輪 SPRIT runtime playback](./round-75-sprit-playback.md)（`READY`：限目前 combat renderer 的 SPRIT playback）
- [第七十六輪 CHEAD／CBODY 玩家小人圖層](./round-76-party-icon-layers.md)（`READY`：限 party icon 合成與素材管線）
- [第七十七輪玩家戰鬥 icon state](./round-77-combat-icon-state.md)（`READY`：限方向／攻擊選擇 state boundary）
- [第七十八輪角色建立玩家 icon 預設](./round-78-player-icon-defaults.md)（`READY`：限新建角色 icon defaults）
- [第七十九輪 SPRIT frame position](./round-79-sprit-position.md)（`READY`：限 animation frame offset）
- [第八十輪 PIC／FINAL XOR delta](./round-80-pic-final-delta.md)（`READY`：限 PIC frame stream decode）
- [第八十一輪 ECL PICTURE 事件畫面](./round-81-picture-event.md)（`READY`：限 PIC request 與 playback slice）
- [第八十二輪 BIGPIC PICTURE 分支](./round-82-bigpic-event.md)（`READY`：限 BIGPIC extraction 與事件畫面）
- [第八十三輪 HEAD／BODY 場景人物圖層](./round-83-head-body-scene-layers.md)（`READY`：限 scene character 素材與合成）
- [第八十四輪 PICTURE HEAD／BODY branch](./round-84-head-body-picture-branch.md)（`READY`：限 Area2 head sentinel 與 scene event）
- [第八十五輪 Area2 HeadBlockId codec](./round-85-area2-head-block-codec.md)（`READY`：限 `0x5C2` 與 PICTURE sync）
- [第八十六輪 combat tile placement](./round-86-combat-placement.md)（`READY`：限 formation 與八方向 delta contract）
- [第八十七輪 CombatMap position state](./round-87-combat-map-position-state.md)（`READY`：限 fighter position／size boundary）
- [第八十八輪 reference combat placement formula](./round-88-reference-combat-placement-formula.md)（`READY`：限 `try_place_combatant` adapter）
- [第九十輪 ITEMS base catalog](./090-items-base-catalog.md)（`READY`：限 128 筆 descriptor codec 與已知名稱 catalog）
- [第九十一輪 equipped fighter effect adapter](./091-equipment-effect-adapter.md)（`READY`：限 readied 武器／護甲投影）
- [第九十二輪 party equipment slot contract](./092-party-equipment-slots.md)（`READY`：限 class mask／slot transaction）
- [第九十三輪 inventory quantity／cursed mutation](./093-inventory-mutation.md)（`READY`：限 stack／readied／cursed）
- [第九十四輪 consumable item use signal](./094-consumable-item-use.md)（`READY`：限 scroll／potion／wand）
- [第九十五輪 ECL SPELL／PROTECTION signal](./095-ecl-spell-protection-signal.md)（`READY`：限 bounded runtime signal）
- [第九十六輪 party spell-slot resolver](./096-party-spell-slot-resolver.md)（`READY`：限 remake roster bridge）
- [第九十七輪 DOS player spell record parser](./097-dos-player-spell-record.md)（`READY`：限已公開 spell 欄位與 remake adapter）
- [第九十八輪 DOS player record 核心角色匯入](./098-dos-player-record-core.md)（`READY`：限單職業核心欄位與 remake projection）
- [第九十九輪 DOS `.SWG` inventory 匯入](./099-dos-swg-inventory.md)（`READY`：限連續 item records 與 party equipment projection）
