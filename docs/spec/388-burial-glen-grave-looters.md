# 第 388 輪：Burial Glen 墳墓掠奪者與純珠寶 TREASURE

狀態：`READY`（限正常 GEO 抵達、隨機遭遇、原始六敵戰鬥、重新安葬／
搜刮分支、精靈好感及一件珠寶；四批上限、所有戰鬥規則與完整視覺 fidelity
仍未完成）

## 1. 地圖與事件入口

DOS `GEO6.DAX` block `40h` 的 `(6,12)` terrain 是 `04h`；它由紅網北方
`(6,14)→(6,13)→(6,12)` 的原始牆面路徑可達。`ECL6.DAX` block `40h`
dispatcher 會把 terrain selector `04h` 送到 decoded payload `+0BF9h`：

- `4CC1h==4` 時不再出現；
- `7F81h==1` 時同一步不重複處理；
- 事件含原始 `RANDOM` gate，因此來回踩格不保證每次都遇到；
- 發生時顯示 thri-kreen 正在挖掘墳墓，接著建立六名敵人。

正常玩家 regression 從 Standing Stone 進入 Myth Drannor、完成紅網雙戰
後，沿上述 GEO 路徑移動；若先遇到一般 ECL random encounter，玩家選
FLEE 後繼續行走，而不是由測試直接跳到事件 PC。

這條路徑也揭露固定 seed 的生命週期：每個 `RunEntry` 都重新建立相同 seed
會永遠得到同一 roll；每次改成 `seed+序號` 又會改變既有長路徑的遭遇組成。
本輪讓 `BlockSession` 的 shared `RuntimeState` 保存一條持續 PRNG stream，
同一基底 seed 仍可重現，跨 invocation 則消耗下一個值。PRNG 狀態尚未進入
save round-trip，因此讀檔後亂數序列 fidelity 仍列為未完成。

## 2. 戰鬥組成

`+0C59h..+0C7Ch` 使用 monster setup `40h`，再由共用 encounter resolver
載入：

- 2 隻 `GIANT SPIDER`（MON6 block `42h`）；
- 3 隻 `PHASE SPIDER`（MON6 block `41h`）；
- 1 隻 `THRI-KREEN`（MON6 block `40h`）。

勝利後 `+0C80h` 把 `4CC1h` 加一、`+0C89h` 把本步 guard `7F81h`
設為一，然後顯示骸骨與三選單。測試使用一般 `CombatAct` 完成戰鬥，
敵人種類、數量及資料均來自原始 ECL／MON6，不由 frontend 寫死。

## 3. 選單真實邊界

`HORIZONTAL MENU` 位於 `+0CDAh`，動態長度 operands 的 exact decoded
bytes 起始如下：

```text
0CDA  2B 01 79 7F 00 03
0CE0  80 08 ...                         LOOT GRAVE
0CEA  80 0C ...                         REBURY SKELETON
0CF8  80 02 1C F0                      GO
+0CFC 25 01 79 7F 00 03 ...            ON GOTO
```

因此 menu 的真正 `stringsEnd` 是 `+0CFCh`，不是固定 arity tracer 曾誤讀的
`+0CDBh`。`ON GOTO [7F79h],3` 的前兩個目的地是：

- index 0 `LOOT GRAVE` → `+0D0Bh`；
- index 1 `REBURY SKELETON` → `+0D28h`。

這次錯誤說明動態選單不可只用 opcode 表線性掃描；必須以 runtime operand
parser 量出字串終點，再追 branch table 與 resume PC。

## 4. 精靈好感與珠寶

好感工作位址 `4CBAh` 採偏移編碼。ECL6 兩處初始化指令
`+0053h`／`+0A5Ah` 都明確寫入 `80h`，所以 raw `80h` 才是邏輯中立：

- `REBURY SKELETON` 在 `+0D28h` 執行 `ADD 1,[4CBAh],[4CBAh]`，
  raw 值由 `80h` 變成 `81h`；
- `LOOT GRAVE` 在 `+0D0Bh` 執行
  `SUBTRACT 1,[4CBAh],[4CBAh]`；
- `+0D15h` 的 TREASURE 七個數量是 `0,0,0,0,0,0,1`，
  `ItemBlock=FFh`，即一件珠寶、沒有 ITEM record；
- `+0D26h` 的 COMBAT 是 treasure-service boundary，不是另一場戰鬥。

舊 adapter 只在 `pendingTreasureItems` 非空時開寶物介面，因此這種
「只有金幣／寶石／珠寶」的合法 TREASURE 會誤落成零怪物 COMBAT。本輪
改為只要 TREASURE 已成功解析就保留可見 service boundary；一件珠寶進入
treasure pool，玩家繼續後由同一 ECL session 執行 `+0D27h EXIT` 回地城。

GameFAQs 的 Burial Glen 攻略也記載重新安葬使精靈認可度 `+1`、搜刮使其
`-1` 並得到一件珠寶；它只作次級交叉驗證，實作數值仍以上述原始 ECL bytes
為準：
<https://gamefaqs.gamespot.com/pc/564786-curse-of-the-azure-bonds/faqs/78365>

## 5. 資料化中文

`LOOT GRAVE`、`REBURY SKELETON` 與 `GO` 透過 game-pack
`option_rules` 的 stable ID／`message_id` 翻譯。State 先查作品資料，
不再為這三個 CoAB 選項擴大中文 switch。測試從實際 pack stable ID 取得
期望譯文；不複製一份會隨 JSON 修改而失效的繁中文字串。

## 6. 畫面證據

`docs/screenshots/burial-glen-grave-looters.png` 是正式
`cmd/azure-bonds-game -burial-grave-battle` 在
Docker／Xvfb／ALSA null device 產生的 640×480 確定性畫面。

SHA-256：
`ef2c4219375e5793b32f126884cd9aa7f45fa0f6f0da34949f48d091cae7c5d5`。

畫面證明原版 cracked-stone combat frame、原始地城 combat terrain、
蜘蛛 CPIC 小人、THRI-KREEN 目標與正式 HUD 已由事件 State 送入 renderer。
目前 thri-kreen 小人被戰場地形／佈陣遮住，不能用這張圖宣稱六名敵人都
清楚可見，也不能證明 placement、遮擋順序或整場動畫已與 DOS 完全一致。

## 7. 尚未完成

- 四批事件全部跑完後的 `4CC1h==4` 正常玩家長回歸；
- GIANT SPIDER／PHASE SPIDER 毒素、phase 行為、thri-kreen 多臂攻擊與 AI；
- 戰敗、逃跑、全滅及 save/resume；
- treasure pool 的完整原版 TAKE／SHARE 金錢與珠寶操作；
- thri-kreen CPIC 清楚可見的 placement／occlusion 校準；
- Burial Glen Daemir、紅羽戰士、最後區域、Tyranthraxus 與結局。
