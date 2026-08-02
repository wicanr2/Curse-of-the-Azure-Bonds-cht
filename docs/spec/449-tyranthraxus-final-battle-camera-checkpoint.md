# 449：提朗瑟克斯終戰畫面與戰鬥鏡頭

狀態：`READY`

## 目標與取樣範圍

本輪不重跑全新隊伍從開場到結局的長回歸，而是抽樣驗證高風險的終戰視覺
邊界：從 block `43h` 正常初始化、經 terrain `97h` 樓梯與第 408 輪證明的
十步二樓路線，抵達 `(6,1)` terrain `9Ah`，保留原 ECL 三段台詞並建立正式
37 名敵軍。

截圖專用隊伍只提高 HP、AC、命中與傷害，避免擷取流程在抵達邊界前戰敗；
它不改寫 terrain、ECL PC、敵軍種類、數量、座標或戰鬥 AI。因此本輪支持
「正式終戰資料已由正常樓梯與 GEO 路線載入」，不支持終戰平衡、完整怪物
能力或開場至結局完整通關。

## 鏡頭修正

[`spec 147`](147-combat-camera.md) 引用原版 RuleBook：大型 CombatMap 的正式
鏡頭應跟隨目前主動角色。renderer 過去卻以全體 fighter bounds 的中心當焦點，
會在大型戰場把主動角色擠到角落。本輪改為：

- 有主動角色位置時，以該角色的 CombatMap tile 為正式焦點；
- 沒有可用主動位置時，才使用全體 bounds 中心作 deterministic fallback；
- 只有 `-inner-final-battle` 且同時要求 `-screenshot` 時，啟用明確隔離的
  `SpriteBlock 47h` 首領觀察鏡頭，方便 README 同時看見原始終戰陣形；
- 首領觀察鏡頭只改 renderer transform，不移動 fighter，也不消耗回合。

## 實機畫面與證據等級

輸出：
[`myth-drannor-tyranthraxus-final-battle.png`](../screenshots/myth-drannor-tyranthraxus-final-battle.png)

- remake：Go／Ebiten，640×480，倚天粗體 16×15，DOS 忠實 theme；
- scenario：ECL6 block `43h`、GEO6 二樓 `(6,1)` terrain `9Ah`；
- enemy oracle：MARGOYLE `45h`×28、TYRANTHRAXUS `47h`×1、
  HIGH PRIEST `48h`×8；
- 原始 CPIC、combat terrain 與 DOS 石框是 `material-exact`；
- 640×480 排版及首領觀察鏡頭是 `layout-reconstructed`；
- 這不是原版同一 frame 的逐像素對照，也不是完整法術、AI、死亡、聲音與
  wall-clock timing 的完成證據。

## 驗證

- `TestCombatCameraFocusFollowsActiveFighterOnLargeMap`：正式大型地圖鏡頭跟隨
  主動角色，缺少位置時才使用 bounds fallback。
- `TestCombatPreviewFocusFindsOriginalBossSpriteWithoutMovingIt`：截圖觀察鏡頭
  依穩定 `SpriteBlock 47h` 取得既有座標；停用時不選取 fighter。
- `-inner-final-battle` runtime capture：實際走樓梯與十步 GEO 路線，並在
  擷取前斷言 `45h×28／47h×1／48h×8`。
- 提交前正式套件 gate 及 Docker 清理結果記錄於 `CONTEXT.md`。

## 尚未完成

- 原版 DOS 同一終戰首幀的逐像素／逐幀 capture。
- 37 名敵軍完整法術、特殊能力、AI、死亡、音效及回合節奏。
- 三神器與提朗瑟克斯所有 executable consumer 的完整資料流。
- 全新隊伍由開場至結局的發行前完整驗收；日常開發依使用者指示採代表性
  vertical slice 與高風險狀態轉換抽樣。
