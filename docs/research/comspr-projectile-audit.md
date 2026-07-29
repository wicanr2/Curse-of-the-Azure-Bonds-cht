# CoAB DOS `COMSPR.DAX` projectile audit

狀態：研究草稿（2026-07-29）

## 問題與結論

本次稽核要回答 `COMSPR.DAX` 的 `0x00..0x0B`／`0x80..0x8B` 是什麼，
以及原版弓箭與 Magic Missile 是否直接存在於這批圖形。

結論：

- **弓箭直接存在。** `0x00/0x80`、`0x01/0x81`、`0x02/0x82` 是同一支
  綠色箭矢的六個方向來源；另外兩個斜向由 renderer 水平翻轉取得。這不只靠
  PNG 外觀：reference decompile 的 `DrawRangedAttack` 對 `Arrow`、`Dart`、
  `Javelin`、`Quarrel`、`Spear` 明確選取 combat icon `13..15` 的
  `Normal/Attack` state，而初始化 mapping 可換算回上述六個 DAX block。
- **Magic Missile 的飛行圖形也直接存在。** 成功的目標型施法在進入實際
  spell resolver 前固定呼叫 `load_missile_icons(0x12)`；combat icon `0x12`
  對應 normal block `0x05` 與 attack block `0x85`。`SpellMagicMissile`
  隨後進入 `DoSpellCastingWork` 結算傷害。因此 `0x05/0x85` 是 Magic
  Missile 實際沿用的四格飛行素材來源，不應再以 renderer primitive 代替。
- `0x05/0x85` 是多種目標型魔法共用的 spell missile，不是只供 Magic
  Missile 專用。文件與程式應稱為「generic spell missile（Magic Missile
  會使用）」；不能反推每一種法術都有獨立 projectile。公開 DOS 影片另在
  `00:36:15.40` 直接拍到 Fireball 使用相同青色 travel，之後才逐目標播放
  `0x0A/0x8A` 外觀的紅白 impact 與必要的死亡骷髏。

以上 consumer mapping 是 **code-backed strong inference**：它由公開的
CoAB decompile、DAX 初始化關係及本機原始圖形三者互相吻合；本次沒有把
decompile 的函式逐條對回 `GAME.OVR` raw instructions，所以不提升為 raw-byte
proven。DAX 容器、block bytes、尺寸和 PNG 外觀本身則已由本機原檔直接驗證。

## 輸入與可重現性

本機輸入：

| input | SHA-256 |
|---|---|
| `curseoftheazurebonds.zip` | `c98698a6271c17177dfdb27f34b0389b7d34f58ef206e92575393f4655f5b26d` |
| ZIP member `COMSPR.DAX` | `6bff177f6776339a627be3c0199518a41330f01f462e8994f649804e488a86d5` |
| ZIP member `GAME.OVR` | `53507d95f65e773ebc0934490e8dd180613f10c9cf4bbad3eed1cf90a9858215` |
| ZIP member `START.EXE` | `dd79b58f872f6f2fae94b96d20b9f82b25dfd33c38e0f9b886891c4994a0e3c5` |

交叉參考：

- Simeon Pilgrim `coab` repository，commit
  `9dc46f1d5911710fb2fcb7a9c0ec0ef74264d17c`。
- 本機 checkout：`/tmp/coab-reference`（只讀研究，不是本 repo 的 dependency）。
- 關鍵檔案：`engine/seg001.cs`、`ovr034.cs`、`ovr025.cs`、`ovr014.cs`、
  `ovr023.cs`、`ovr033.cs`。

方法：

1. 由 ZIP 直接抽出 `COMSPR.DAX`，解析 9-byte DAX directory entries 及 SSI
   RLE；不以衍生 PNG 反推 raw block。
2. 驗證 26 個 block 都解成 305 bytes：17-byte picture header 加
   `24×24÷2 = 288` packed nibbles，皆為一張 masked 24×24 EGA picture。
3. 逐張檢視 repo 既有 `assets/sprites/comspr-block-*.png`，並量測 opaque
   bounding box／palette；這只支持形狀辨識。
4. 追蹤 reference 初始化的 block→combat-icon mapping，再追
   `load_missile_dax`、`load_missile_icons`、`DrawRangedAttack` 與 spell
   consumer。程式路徑用於命名；視覺不吻合時不得靠名稱覆蓋 raw evidence。
5. 搜尋本 workspace 與 `/home/anr2` 的 `.idb/.i64/.asm/.lst`。沒有找到
   CoAB 專用 IDA database 或 assembly listing；repo 既有 knowledge 只有
   consumer 結論，沒有保存原始 artifact。

## DAX directory 與 raw block anchors

`COMSPR.DAX` 的 directory 長度欄為 `0x00EA`，所以 payload base 是
file offset `0x00EC`。下表的 offset 是 packed block 在 `COMSPR.DAX` 內的
絕對 file offset；所有 decoded size 都是 `0x0131`（305）。

| block | file offset | packed bytes | decoded SHA-256 |
|---:|---:|---:|---|
| `0x00` | `0x00EC` | 96 | `d69e79fa703a12089bd935c5da3aec1feb8c504dfc17ceb534f07107fb4f7c60` |
| `0x80` | `0x014C` | 96 | `8bdd6e067e5d36dda0abccb03a6d9837e1a0f0d29464b64b651d51789f0cba88` |
| `0x01` | `0x01AC` | 89 | `5bdc6220c9906c9aa707fbe9eeabf2d5e4eccc33ec9bde71686ceab02f7d1b24` |
| `0x81` | `0x0205` | 89 | `eba4d4677db1b70957b5b7368833a6a6dc9f375c83901b41d2e10286d4e8ccfe` |
| `0x02` | `0x025E` | 73 | `7010719ab97ef86ed47f848c67e1819f31b3b34c9331aa0c81f02554bcc24398` |
| `0x82` | `0x02A7` | 73 | `b269f296ad82be3be03cc7c772e00387dcdbb04f9498746feac866c28cc20f00` |
| `0x03` | `0x02F0` | 124 | `c078d01b00f5f4b392ad31ff3b6153fb78138ad0907a23c9b2a75fbf4b0ded83` |
| `0x83` | `0x036C` | 121 | `b21ee99e1c4e6f11fe1cef0241475a0f9b1a7e8c57b09f85137a3c43256a3c57` |
| `0x04` | `0x03E5` | 82 | `a505746585a28396011103c008c03d389f1a6abf8a132d110dc4d76cf214d99c` |
| `0x84` | `0x0437` | 90 | `4e87cd7b71b115c084a128d9109e9391a8a9c58c1c655f0470208c40dfc7e91c` |
| `0x05` | `0x0491` | 126 | `87738079cda33458ece082a5b68be3fcdd70537b66769c8ad5bc8f635dca63f7` |
| `0x85` | `0x050F` | 126 | `1649b72d4fafbb21904ca1562af6d50ed0258c30e46157c71a843863f0470853` |
| `0x06` | `0x058D` | 171 | `78a84103fc3a02fea445081551049b6d225e90e1323a440f7d580b602202e5fd` |
| `0x86` | `0x0638` | 191 | `b234b12b13021ab509cd95599d1c92ba28e751b5af1c30ed2e1a81ac1f42009e` |
| `0x07` | `0x06F7` | 145 | `b5668ca976ab8e558a83d8aabf81d4cf13658da5fe7a3b2ab6b8eb2f3333c044` |
| `0x87` | `0x0788` | 148 | `24ef570400478442873563c257b8e1de3007b9fd14e687b1d7928fa1874ae47f` |
| `0x08` | `0x081C` | 37 | `b5d246908d0ba68d12e4a435ac189ad6201f0eb896aceb18699f636eb2670286` |
| `0x88` | `0x0841` | 37 | `e8c2e81233a82cf4bd93785be54f436f458559cc58875d77845c94ef7c7e7aec` |
| `0x09` | `0x0866` | 167 | `986fbca853ffdc19e427a10fd67162a1ad1fb3ac43f24010a80b812fc6505b57` |
| `0x89` | `0x090D` | 164 | `ec39d40f93b417a000845af07c3e4ddbb58da1ee5ff6fcb2ee8e02e7e537eb20` |
| `0x0A` | `0x09B1` | 71 | `299209de273519d93e5ac576d3dc507ccd35bf89813178f3c790561973963da9` |
| `0x8A` | `0x09F8` | 284 | `7e233f9fa33bbf7e66e11a2ba61fb7eba0063bf150494d327ed964c708e3c1e7` |
| `0x0B` | `0x0B14` | 254 | `d3dfc2c27932d2e4ab74df4011ce1d85e936338d12110b90c9fac401e6fe3bb6` |
| `0x8B` | `0x0C12` | 254 | `460bf2762115c0bc6830f11fcff8aeefe584f54b77216ef70a1b89d399df4265` |

檔案另有 `0x19/0x99`，不在本題主要範圍；兩者是實色 24×24 background
tiles。死亡 routine 以 `0x8B` 骷髏和 `0x19` 灰底交替，而不是
`0x0B/0x8B` 互相交替。

## 初始化與 frame 組合

Reference `seg001.Init`：

- `seg001.cs:312-315` 將 block `0x00..0x0B` 載入 combat icon
  `0x0D..0x18`。
- `ovr034.cs:69-71` 將每個 normal block `b` 與 `b+0x80` 一起交給
  `LoadIcons`。因此：

```text
combat icon index = 0x0D + normal block
Icon.Normal       = COMSPR block
Icon.Attack       = COMSPR block + 0x80
```

`Normal/Attack` 是 `CombatIcon` 的兩個通用 storage slots。在 COMSPR
projectile consumer 中，它們代表方向或 animation phase，不應依欄位名稱
誤解成 projectile 在「普通／攻擊姿勢」。

`ovr025.cs:873-878` 的 `load_missile_icons(index)` 組成四格：

1. normal；
2. normal 的水平翻轉；
3. attack 的水平翻轉；
4. attack。

`draw_missile_attack` 再沿 attacker→target stepping path 疊圖。這證明
`0x05/0x85` 等 pair 是四方向／四 phase 的來源，而非兩張互不相關圖片。

## Block 語意表

| blocks | PNG 外觀（visual） | consumer evidence | 判定 |
|---|---|---|---|
| `00/80`, `01/81`, `02/82` | 綠色箭桿與箭頭，垂直／斜向／水平 | `DrawRangedAttack` (`ovr014.cs:1590-1631`) 對 `Arrow` 等選 icon `13..15`，依 8-direction 選 Normal/Attack/flip | **code-backed strong inference：箭、矢、標槍等共用方向 projectile** |
| `03/83` | 旋轉中的帶柄刃器 | `ovr014.cs:1633-1640` 對 Hand Axe／Club／Glaive 載 icon `16` 的四格 | **code-backed strong inference：旋轉投擲武器** |
| `04/84` | 小型青白彈體／容器的兩個角度 | `ovr014.cs:1642-1648` 對 Flask of Oil／unknown `Type_85` 載 icon `17` | **code-backed strong inference：油瓶 projectile**；具體 `Type_85` 名稱仍 unknown |
| `05/85` | 青色四叉／飛散形，黃點綴；兩張是不同 phase | 施法共同路徑 `ovr023.cs:739-762` 固定載 icon `0x12`；`SpellMagicMissile` 在 `ovr023.cs:1166-1171` 進入該結算流程；DOS 影片 `00:36:15.40` 亦直接顯示 Fireball 使用同形 travel | **code-backed + video-backed：generic spell missile；Magic Missile、Fireball 直接使用** |
| `06/86` | 青白分叉電弧 | electrical damage 路徑在 `ovr023.cs:1951-1961` 載 icon `0x13`，Lightning Bolt／dragon electricity 亦使用 | **code-backed strong inference：電擊／Lightning Bolt projectile** |
| `07/87` | 灰色圓盤／石塊的旋轉 phase | ranged default branch `ovr014.cs:1661-1667` 載 icon `20` 的 Normal/Attack | **code-backed strong inference：未分類 ranged fallback**；不能只靠外觀命名為石頭 |
| `08/88` | 4×4 小彈丸 | Sling／Staff Sling／Spine branch `ovr014.cs:1650-1659` 載 icon `21` 的兩格 | **code-backed strong inference：投石索／spine pellet** |
| `09/89` | 散佈的青白星芒，兩格有位移 | `MagicAttackDisplay(showMagicStars=true)` 選 icon `0x16`，原地反覆播放四格（`ovr025.cs:1118-1158`） | **code-backed strong inference：magic status/heal stars**，不是 Magic Missile 飛行彈 |
| `0A/8A` | 小紅命中叉 → 大型紅白爆點 | 同 routine 在 `showMagicStars=false` 時選 icon `0x17` | **code-backed strong inference：magic hit/status burst**；不是 Magic Missile 飛行彈 |
| `0B/8B` | 紅／亮紅底的骷髏 | `CombatantKilled` 只取 combat icon `24` 的 Attack（即 `0x8B`），並與 icon `25` Normal（即 `0x19` 灰底）交替，見 `ovr033.cs:548-568` | **code-backed strong inference：死亡 skull family**；本次沒有找到 `0x0B` 的 runtime consumer |

## 弓箭方向 mapping

依 `DrawRangedAttack` 的 `dir=0..7` 分支，可把箭矢來源還原為：

| direction | source |
|---:|---|
| 0 | `0x00` |
| 1 | `0x01` |
| 2 | `0x02` |
| 3 | `0x81` |
| 4 | `0x80` |
| 5 | `0x81` horizontally flipped |
| 6 | `0x82` |
| 7 | `0x01` horizontally flipped |

這也解釋為什麼檔案只有六張箭，卻能覆蓋八方向。原版 ranged branch 對這組
projectile 使用 `frame_count=1`、初始 `delay=10`；它沿 stepping path 移動
同一個方向圖，不是用 `0x00→0x80` 交替做兩格箭矢動畫。`dir` 的實際畫面
方位仍應以 DOS runtime capture 校正，不在只有座標慣例的情況下硬寫成
北／東／南／西。

## Magic Missile mapping

成功施法的顯示順序是：

```text
SpellCastFunction succeeds
  -> load_missile_icons(0x12)
     -> combat icon 0x12
     -> normal block 0x05 + attack block 0x85
     -> [05, flip(05), flip(85), 85]
  -> caster Attack pose
  -> sound selector (通常 2)
  -> draw_missile_attack(delay=0x1E, frameCount=4)
  -> restore caster pose
  -> spell resolver
     -> SpellMagicMissile -> DoSpellCastingWork damage
```

因此 remake 可立即採用原始 `0x05/0x85`：

- travel frame 依原版四格順序播放；
- path 依 attacker→target 逐步移動，而不是只在兩格之間線性畫一條線；
- 原版 delay 參數是 `0x1E`，但其 wall-clock 單位仍需 DOSBox runtime
  trace；不能直接把 `0x1E` 宣稱為 30ms；
- 多枚 Magic Missile 是否逐枚重播、如何分配多目標，仍需 spell target
  runtime capture。這份 audit 只證明 projectile source 與共同播放入口。

## 與既有文件的關係

`docs/knowledge/gold-box-graphics.md` 已正確記錄：

- `0x00..0x0B` 載入 icon `0x0D..0x18`；
- attack block 使用 `+0x80`；
- 死亡顯示使用 `0x8B` 與 `0x19`。

但它尚未列出 ranged/spell consumers。本草稿補足的重要修正是：

- `0x00..0x02/0x80..0x82` 不是未分類 COMSPR effect，而是弓箭分支直接使用；
- `0x05/0x85` 不是只能視覺猜測的魔法圖，而是 Magic Missile 會經過的
  code-backed generic spell missile；
- `0x09/0x89` 的星芒與 `0x0A/0x8A` 的爆點是 target-local feedback，
  不應拿來取代 Magic Missile travel；
- death flash 是 `0x8B ↔ 0x19`，不是 `0x0B ↔ 0x8B`。

## 未解事項與下一個驗證

1. 用 DOSBox 原版錄影在同一場戰鬥逐格擷取弓箭與 Magic Missile，確認
   palette 0/8 swap、masked blit、tile anchor、path cadence 與命中銜接。
2. 以 IDA 載入本機 `START.EXE`＋`GAME.OVR`，把 reference comments
   `sub_40BF1`、`sub_67924`、`sub_67A59`、`sub_67AA4`、`sub_5E221`
   對回 raw instructions／overlay offsets。完成前，本文件的 consumer
   結論保持 strong inference。
3. 找到 `0x0B` Normal state 的 consumer，或明確證明在 CoAB DOS release
   中只被初始化而未使用。
4. 對 `0x07/0x87` 的 default ranged branch 收集實際 item type，避免把
   灰色圓盤只憑外觀命名成 stone。
5. 驗證 Magic Missile 多 projectile／多 target 的重播次數；目前 source
   只足以證明每次共同施法 travel 使用 `0x05/0x85` 四格。
6. Fireball 影片證明單次 travel 後逐一 impact／death；下一步把
   `sub_5F782` 的 radius target ordering、saving throw 與 damage dice
   對回 raw overlay instructions，再建立可玩的多目標 timeline。
