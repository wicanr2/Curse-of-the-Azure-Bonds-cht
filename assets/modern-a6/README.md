# A6 現代重繪資產

本目錄只放本專案新製作的現代 theme 素材，不覆蓋 `assets/sprites/` 或
`assets/runtime-images/` 的原版匯出 PNG。F2 切回忠實 theme 時不得讀取本目錄。

美術契約：

- `pictures/`：高清手繪，保留原版構圖、姿態、辨識色與敘事功能。
- `sprites/`：現代像素畫，保留 Gold Box 戰術格、輪廓與朝向辨識。
- `tiles/`：高清手繪材質，保留圖塊索引、通行語意與俯視網格辨識。
- `ui/`：A6 細石雕文飾、雙層結構與極細明亮金色內框。

目前已完成場景人物、AREA-map tileset 與 CPIC 戰術 sprite。已接入項目如下：

- `pictures/character-area-5-head-03-body-03.png`；原版同 bytes 的 area 2／3／5
  `HEAD 03 + BODY 03` 共用此重繪。
- `pictures/character-area-2-head-04-body-04.png`；提爾佛頓旅店的第二名常見 NPC。
- `pictures/character-area-2-head-01-body-01.png`；灰帽、面罩、紅袍施法人物。
- `pictures/character-area-3-head-16-body-16.png`；金髮、紅披風與青藍甲女性。
- `pictures/character-area-2-head-02-body-02.png`；青綠衣、持匕首的警覺人物。
- `pictures/character-area-2-head-05-body-05.png`；金飾紫袍的宮廷女性。
- `pictures/character-area-5-head-31-body-31.png`；白髮、紫膚、黑甲與金護腕卓爾。
- `pictures/character-area-5-head-33-body-33.png`；白髮、綠寶石額飾的卓爾女性。
- `pictures/character-area-5-head-3A-body-3A.png`；紅髮紅袍的雙手施法女性。
- `pictures/character-area-3-head-10-body-10.png`；紅冠白盔與銀藍重甲戰士。
- `pictures/character-area-2-head-00-body-00.png`；深膚藍髮、白甲與紫徽戰士。
- `pictures/character-area-2-head-41-body-41.png`；紅鬍翼盔、藍衣金獅戰士。
- `pictures/character-area-3-head-11-body-11.png`；金髮鮮綠兜帽女性。
- `pictures/character-area-4-head-20-body-20.png`；橙金肩甲與黑月徽重甲男性。
- `pictures/character-area-3-head-12-body-12.png`；藍衣鎖甲、側望警戒的守衛。
- `pictures/character-area-3-head-13-body-13.png`；銀翼肩與淺綠圓徽聖職戰士。
- `pictures/character-area-4-head-21-body-21.png`；白鬍金棕兜帽的紫袍神職者。
- `pictures/character-area-4-head-22-body-22.png`；橙金小帽、紅袍與布包侍從。
- `pictures/character-area-2-head-06-body-06.png`；紅鬍純白兜帽修士。
- `pictures/character-area-4-head-23-body-23.png`；藍頭巾、綠袖與棕皮甲商旅。
- `pictures/character-area-4-head-2A-body-2A.png`；黃袖紫披肩、抱青綠書的學者。
- `pictures/character-area-4-head-31-body-31.png`；亮綠短袖與黑色小徽記青年。
- `pictures/character-area-4-head-46-body-46.png`；深膚、深藍兜帽與棕腰帶人物。
- `pictures/character-area-2-head-09-body-06.png`；深鬍純白兜帽修士。
- `pictures/character-area-5-head-32-body-32.png`；紅棕髮亮青藍衣青年。
- `pictures/character-area-5-head-3B-body-3B.png`；白頭巾、紫袍與大型金墜男性。
- `pictures/character-area-6-head-40-body-40.png`；藍紫長袍、洋紅內襯旅人。
- `pictures/bigpic*.png`；4／4 張 BIGPIC 均為 1216×480 高清手繪，保留原版
  304:120 構圖比例與地圖／骷髏／三女法師／夜村的辨識配置。
- `sprites/cpic1-block-01-item-00.png` 至 `cpic6-block-C7-item-00.png`；156 個原始入口
  全部有 2× edge-aware 現代像素重繪，涵蓋 122 種原始唯一視覺；`cpic1/01` 保留
  手工精修 override，因此現代輸出本身共有 123 種唯一 PNG。
- `tiles/tiles-block-01-item-000.png` 至 `tiles-block-02-item-021.png`；48 個 AREA-map
  索引全部以 96×96（原 24×24 的 4 倍）材質化重繪保留道路／符號、紅黑陰影與邊界色。
- `ui/adventure-frame.png`、`ui/combat-frame.png`；戰鬥框是獨立透明層，沿用核准的
  18px 精修石材、雙層亮金雕線、藍寶石節點，不以平面灰框替代。
- `sprites/pic*-frame-*.png`、`sprit*-frame-*.png`；PIC 152 檔／107 種與 SPRIT
  138 檔／121 種全部保留原動畫 metadata 並以雙解析度載入；現代調色降低 EGA
  螢光飽和度，新增可重生的材質微差、受光邊與背光陰影，避免四格同色假放大。
- `sprites/chead*`、`cbody*`、`head*`、`body*`、`party*`、`comspr*`；角色建立、
  場景分層、隊伍合成與戰鬥符號的所有唯一視覺均有現代檔。
- `combat/`、`symbols/`、`sky/`；65 張戰場圖塊、1,625 張第一人稱符號與 3 張天空
  均完成雙解析度材質層；牆符號透明色必須依固定 EGA index 13 遮罩，不可猜
  top-left。材質深度不可改變 WALLDEF 索引、牆面幾何或戰場碰撞語意。

## 覆蓋台帳

2026-08-28 以原始 PNG SHA-256 去重後的批次分母與目前完成量：

| 類別 | 原始檔案 | 唯一視覺 | 已完成現代資產 |
|---|---:|---:|---:|
| 場景人物 picture | 31 | 27 | 27 |
| 戰鬥 CPIC sprite | 156 | 122 | 122 |
| TILES tile | 48 | 48 | 48 |
| BIGPIC | 4 | 4 | 4 |
| PIC 動畫 | 152 | 107 | 107 |
| SPRIT 動畫 | 138 | 121 | 121 |
| 角色建立頭／身 | 184 | 153 | 153 |
| 場景頭／身 | 71 | 60 | 60 |
| 戰鬥符號 | 26 | 26 | 26 |
| 隊伍合成 | 18 | 12 | 12 |
| 戰場地形 | 65 | 63 | 63 |
| 第一人稱符號 | 1,625 | 1,185 | 1,185 |
| 天空 | 3 | 3 | 3 |

表中的「已完成」按唯一視覺計；重複入口仍以同名檔或已證實相同的別名完整解析。

全資產族的機器可讀分母在
[`docs/audit/modern-a6-asset-inventory.json`](../../docs/audit/modern-a6-asset-inventory.json)，
由 `tools/modern-a6-inventory.py` 依目前工作樹重生並以 `--check` 驗證。該台帳亦含
角色建立頭／身、動畫、天空、第一人稱符號與戰場地形；完成聲明必須以全台帳為準，
不能只引用上面的三類先行數字。

場景人物的 31 個原始入口現已全部解析至 27 種現代重繪。少於 31 並非漏檔：
area 2／3／5 的 `03/03`、area 2／6 的 `41/41`、area 4／6 的 `46/46` 原始合成圖
分別經 SHA-256 證實逐 byte 相同，由 `modernScenePictureKey` 明確共用同一重繪。
