# 1130 — 四張 README 截圖背後的四個對齊錯誤

- 證據等級：`exact`（每一條都由重拍的畫面逐像素量出來，並與已有規格對照：
  spec 406 的 88×88 可見區、spec 1006 的 11×11 牆位表、spec 326 的 CPIC 錨點、
  spec 88 的 `try_place_combatant`、round-75 的 CPIC／SPRIT 分工）
- 起點：使用者 2026-08-18 逐張指出 README 圖不對

## 起因

四張圖的問題看起來各自獨立，量下去卻是同一類：**兩層各自算座標，然後對不上。**

| 圖 | 症狀 | 真正的原因 |
|---|---|---|
| `gold-box-layout-adventure` | 黃框裡的人像沒佔滿 | sub-image 目的地座標被扣了兩次 |
| `gold-box-layout-combat` | 人站在綠色灌木上 | 地形層沒過相機與鏡像 |
| `burial-glen-red-web-spiders` | 人站在牆裡、四隻蜘蛛全不見 | 同上 ＋ 編制格在視野外 ＋ 佈陣不看地面 |
| `tilverton-first-person-remake` | 像沒有外框、3D 組成可疑 | 框與幾何都沒錯；但一大半牆磚沒畫出來（spec 1131）|

## 一、sub-image 只裁切，座標仍是母圖的

`screen.SubImage(clip)` 當繪製目的地時**只做裁切**，`DrawImage` 的座標仍在母圖
座標系（Ebiten：「If a sub-image is used as a rendering destination, the region
being rendered is clipped」）。`drawSceneCharacter` 與 `drawImageCover` 都額外減了
一次 `clip.Min`，於是圖被畫到畫面左上角，只剩「剛好和黃框重疊的那一小塊」露出來。

量到的證據：修正前的旅店畫面，黃框內像素與 `character-area-*-head-03-body-03.png`
的第 `(27,21)` 格起完全相同（3432 格 0 不符），正是「畫在 `(2,4)` 再被裁切」的結果。
`imageCoverTransform` 的既有測試名字就叫 `UsesGlobalDestinationOrigin`——回傳的
本來就是全域座標，呼叫端不該再扣一次。

## 二、地形層要跟戰鬥員走同一條座標路徑

戰鬥員的畫面格是 `鏡像(相機(地圖格))`＝`(6 − (x − 原點x), y − 原點y)`
（鏡像見 spec 326），地形層卻直接拿畫面欄列當地圖座標。相機一開，兩層就整個錯開。

Burial Glen 那張的實測：戰鬥員在地圖 `(0,0)`，該格是可通行地板；相機原點 `(−3,−3)`
把他畫到畫面 `(3,3)`，而畫面 `(3,3)` 的地形是照「地圖 `(3,3)`」畫的——那格是牆。
**人沒有站在牆裡，是牆被畫到他腳下。**

修法是把地形查詢改走反函式 `(原點x + 6 − 欄, 原點y + 列)`，並把地形迴圈搬到
相機算完之後。測試 `TestCombatMapTileForScreenInvertsFighterAnchor` 釘住這條
往返關係。

## 三、佈陣：候選格要通過地面檢查，編制帶要在視野裡

兩件事：

- **地面**：原作 `try_place_combatant` 是在候選格通過 occupancy 與 ground 檢查
  之後才寫座標（spec 88）。remake 的 fallback 只算格號。實測 `-encounter` 開場
  28 個戰鬥員裡有 4 個站在 `MoveCost = 0FFh` 的格上。現在從自己的編制號往後找
  第一個沒被佔用且可通行的格。
- **視野**：敵方縱帶原本是 `7..9`，而視野只有 `0..6` 七欄。**兩隊永遠不可能同時
  入鏡**——相機必然被迫開啟，對準哪一隊，另一隊就被推出畫面。Burial Glen 那張
  因此只剩一個人、四隻蜘蛛全在畫面外。改成 `4..6` 之後兩隊都在視野內，
  也不再強制開相機。

⚠ 這一條與函式自己的註解（「distinct halves of the current combat viewport」）
本來就矛盾。**這仍是 fallback**：沒有 SETUP MONSTER 的距離與 occupancy 表之前，
掃描順序與縱帶位置都是本作自訂的，不宣稱與原作逐格相同。

## 四、戰場圖示是 CPIC，不是 SPRIT

round-75 早就寫明分工：「CPIC 負責靜態／攻擊 icon，SPRIT 負責 encounter animation」。
但 `drawFighterSprite` 把 SPRIT 動畫排在 CPIC 前面。

兩者的畫布尺寸就已經說明了差別：

| 來源 | 尺寸 | 對得上什麼 |
|---|---|---|
| CPIC | 24×24、48×24、24×48、48×48 | **恰好是格子與 footprint**（原始格 24px）|
| SPRIT | 24..64 寬 × **73..80 高**，圖案貼在畫布底部 | 不可能是格子圖示 |

當成格子圖示畫出來，怪物會落在自己那格下方約一格半，選取白框裡因此空無一物。
`animation.json` 的 frame offset 只有 `x = 0..5`、`y = 0..1`，不是這個位移的來源
——**位移來自畫布本身**。

## 五、第一人稱：框在、幾何也對

使用者指出的兩點，量完都是誤會，但誤會有原因：

- **外框在**。`dos-first-person-stage-frame.png` 是一圈點狀白線，原版
  `tilverton-first-person-demo.png` 的 native 第 23 列與第 119 列也是同一種
  `#.#.#.` 點狀線。看起來像沒有，是因為 checkpoint 當時面向東、整個場景是
  一片平坦灰牆，框線與牆之間沒有對比。
- **3D 組成與 spec 1006 逐項相符**。引擎的 `wallLayoutIndex`／`Columns`／`Rows`
  ＝ `0,2,6,10,22,38,54,110,132,154` ／ `1,1,1,3,2,2,7,2,2,1` ／ `2,4,4,4,8,8,8,11,11,2`，
  與規格的十個牆位表完全一致；三層掃描的起點、步數、走向與 `var_12` 序列
  （深度 2 的 `0,±2,±4,±6`、深度 1 的 `∓6,∓3,0`、深度 0 的 `∓7,0`）也逐條對得上。
  `wallStampNativePosition` 是 `((欄+3)×8, (列+3)×8)`，11×11 格 × 8px 正好鋪滿
  spec 406 的 `(24,24)..(111,111)`。

新增 `-dungeon-facing`（0..7，0 為北），與既有的 `-dungeon-x`／`-dungeon-y` 同一族，
用來從任一朝向檢查視野。

★ 幾何對了不等於畫面對了：**該畫的格子有一大半根本沒畫出來**（天空格與側牆的
斜邊全部落在沒有載入的第 0 段符號），見 [spec 1131](1131-wall-symbol-group-zero.md)。

⚠ `docs/reference/original-dos/tilverton-first-person-demo.png` **不能**當這個
位置的 oracle。它的位置列印是 `7,13 N`，畫面卻是洋紅天空的戶外林地與帳篷，
地名是 `NOWHERE IN THE REA…`——那是另一張圖，不是提爾佛頓街道。
**檔名相符不等於場景相符。**

## 六、順帶：checkpoint 先前完全沒有地形

三支戰鬥地形投影（`SetCombatLineTerrain`／`SetCombatMovementTerrain`／
`SetCombatScanMapProvider`）原本裝在 `RunGame` 前一行，而 `-encounter`、
`-burial-red-web-battle` 這些確定性 checkpoint 在那之前就已經 `StartCombat`。

也就是說**那幾場戰鬥完全沒有地形**：佈陣不看地面、AI 移動不算成本、閃電也不會
撞牆反彈（`combatLineTerrain` 的 `Reflect` 條件是 `DUNGCOM` 且 `MoveCost = 0FFh`，
規則一直都在，只是沒被裝上）。README 的戰鬥圖正是這條路拍的。

現在三支投影在任何 checkpoint 開戰之前就裝上，`app.state` 改成指標讓
checkpoint 對 state 的改動與投影看到的是同一個物件。`combatLineTerrain` 的
模式與座標命名空間也改成**每次查詢時取**，不再凍結在安裝當下。

`-encounter` 另外補上荒野座標 `(25,12)`：先前只有 `-combat-terrain WILDCOM` 才設，
預設路徑停在 `(0,0)`，查表一半落在地圖外，戰場大半沒有地形。

## 明確不宣稱

- 沒有宣稱 fallback 編制格的縱帶位置、列起點或掃描順序與原作相同。
- 沒有宣稱 `6 − x` 水平鏡像是原作的螢幕慣例；它是 spec 326 既有的 CPIC 錨點，
  本輪只是讓地形層跟上它。
- 沒有宣稱 SPRIT 畫布相對於格子的正確錨點。本輪只確定它**不是**格子圖示；
  沒有 CPIC 時仍退回 SPRIT，那條路的錨點還沒量。
- 沒有宣稱第一人稱的每一塊牆磚選對了圖；本輪比的是**幾何**（牆位表、掃描序列、
  格子鋪滿範圍），不是 `WALLDEF` 的美術對應。
- 沒有宣稱閃電反彈在畫面上已驗；本輪只確認規則現在有地形可查。
