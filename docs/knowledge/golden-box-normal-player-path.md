# SSI Gold Box 正常玩家路徑知識

本頁記錄如何把反組譯／原始資料解碼成果接成可重播的正常玩家路徑，供青色枷
與後續 SSI Gold Box game pack 共用。這不是攻略，也不是把測試 shortcut 包裝成
通關證明。

## 路徑的四層

每一段玩家路徑都要分開保存：

1. **來源層**：原始 executable、ECL／GEO／DAX bytes、符號、手冊與 runtime
   observation。記錄檔名、SHA-256、工具版本、位址空間。
2. **狀態層**：作品中立 engine 的 typed state、ECL session、座標／方向、
   pending continuation 與 save snapshot。
3. **資料層**：CoAB game-pack 的 block、map selector、事件 ID、stable message
   ID、人物／怪物／戰利品與翻譯；不得把劇情常數塞進共用 runtime。
4. **驗收層**：由新遊戲或正常地圖輸入抵達，保存每個 boundary 的 state、畫面、
   音效 intent、Journal／獎勵與離開後重訪結果。

## 地城移動交易

第 510 輪的 `State.MoveDungeon` 是一個可重用的範例。前端不應自行做一半：

```text
player input
  → cardinal delta／direction validation
  → decoded GEO wall／door legality
  → geometry coordinate wrap
  → ECL-local projection（若有 map adapter）
  → wall／roof register projection
  → transient per-step guard clear
  → ECL per-turn
  → ECL search-location／event boundary
  → renderer refresh after success
```

`GEO` 的 `(x,y)`、ECL 的 `C04B／C04C` 與檔案 offset 不可混稱。先以
`DungeonGeometryView` 取得 geometry 空間，再由 `SetDungeonGeometryView` 投影回
script 空間；日後其他 Gold Box 作品若有不同鏡像／局部地圖，只新增 typed adapter，
不要在劇情 State 內寫一個座標特例。

### 提爾佛頓 E2：橋接牆與外部邊界要分開

第 536 輪以原始 `GEO2.DAX` block 3 重生出下列候選路徑：

```text
(13,10) → (12,10) → (11,10) → (10,10) → (10,11) → (10,12)
        → (11,12) → (11,13) → (11,14) → (11,15) → (10,15)
        -- (10,15,S) 的 E2 boundary → block 4 (8,1,S) --
```

⚠ 往火刀據點的來源格是 `(10,15)` 不是 `(8,15)`：`(8,15)` 的南面走不出去
（移動遮罩 `3`），而腳本去程 `X := X − 2`、回程 `X := X + 2`，回程落點正是
`(10,15)`。逐格證據與下水道三塊區域的圖緣傳送見
[spec 1199](../spec/1199-tilverton-sewers-map-edge-handoffs.md)。

`wall=09/detail=0`（`(10,12)↔(9,12)`）是另一件事：它通往下水道**中間那一塊
區域**（X 5..9），不在往據點的路上。正常玩家驗收必須分成：

1. Search／Look 輸入如何讓 `(10,12)↔(9,12)` 成為可走邊；
2. 走到 `(10,15)` 後如何發出 E2 boundary attempt；
3. block 4 初始化、返回與 save-load 持久性。

原版手冊的 `SEARCH` 是持續 toggle，`LOOK` 是目前格子的單次檢查；第 537 輪
remake 已以 `ToggleDungeonSearch`／`LookDungeonLocation` 分離兩個 service，並
把發現的 edge ID 存入 save v12。不要把 `wall=09` 的原版 writer 或 E1 座標的
`strong inference` 誤寫成原版 `exact`；remake 正常路徑的完成證據見第 537 輪規格。

## 測試規則

- 最強的路徑測試從 `ActionStart`、角色建立或版本化 save 開始，使用正式 game pack
  與原始 bytes；不要先設定目標座標、旗標或 ECL PC。
- 若為了鎖定單一 opcode 而直接設定座標／flag，名稱必須包含
  `coordinate-assisted` 或 `direct-entry`，並且不得列入完整正常玩家路徑證據。
- 事件輸入、選項、文字、裝備與法術期待值要以 stable ID／JSON resolver 取得；
  測試不能複製目前的繁中顯示字串。
- 每段路徑都要驗證 continuation：戰鬥後、圖片後、手札後、離開後重訪，而不是
  只驗證第一個畫面。
- 通過一段路徑只提升該段的 evidence level，不會自動提升整作完成度。

## 提爾佛頓設施路徑與 one-shot 群組

第 511 輪把第 510 輪的一格交易延伸到提爾佛頓設施之間。從開場中斷休息後的
`(4,13)`，正式測試透過 `State.MoveDungeon` 逐步抵達 Filani、Weaponers、Gond
altar、Training Hall、Tavern 與高階祭司所在格；途中招牌與下水道傳聞的繁中內容
由 locale JSON 提供，未知 pause 不可被測試忽略。

高階祭司是正常路徑驗證中特別容易誤判的例子。原始 ECL2 block 1 的 SearchLocation
先由 terrain 得到 `7F7Ah`，高階祭司入口在 payload `+1104h` 以
`AND 4C03h,7F7Ah,7F79h` 後於非零時 `EXIT`。因此 terrain `0x0D` 招牌先把
`4C03h` 設為 `0x80` 後，之後抵達 terrain `0x8F` 可能按同一 one-shot 群組保持
靜默；這是原始控制流，不是 remake 應清掉的 stale flag。若要驗證祭司本身，應以
未先消耗該群組的 fresh ECL session 進行 branch regression，再另外驗證正常重訪
不重播。

這裡的「共用事件群組」是 `strong inference`：AND／EXIT bytes 與 runtime 前後
狀態已閉合，但 `4C03h` 在所有 SSI 作品的泛用名稱、存檔格式與所有城市事件仍未
完整解出。不要把 `0x80` 直接寫成跨作品固定劇情旗標。

## 城門封路與皇家馬車的移動邊界

第 512 輪把高階祭司後的城門段落接回同一個正常移動交易。decoded `GEO2.DAX`
block 1 證明 `(1,10)` 到 `(1,0)` 有一條 16 步可行路徑：向東兩格、向北三格、
向東一格、向北七格，再向西三格。測試逐次呼叫 `State.MoveDungeon`，不設定目標
座標、不注入 ECL PC，也不在最後一格手動呼叫城門事件。

最後一個西行步驟抵達 `(1,0,W)` 時，原始事件產生
`tilverton.carriage-gate-closed`。這是移動途中發生的 boundary；按下唯一繼續
選項後仍留在同一格，接著只轉向北方，再執行 lifecycle，才會得到皇家馬車的
`PICTURE 11`。因此不要把「先轉身、再執行 lifecycle」誤寫成第二次城門觸發，也
不要在測試中把封路訊息吞成任意按鍵。

城門後的事件鏈仍沿同一 `BlockSession`：皇家馬車喊話、青色枷強制攻擊、皇家衛兵
戰鬥、紅袍人綁走假國王、投降／入獄與盜賊 `PICTURE 2` 救援，最後轉入盜賊公會
地城 block 2。這證明了本段的 ECL continuation，但不代表公會與下水道內部已經
全部改成正常移動。

本輪遇到的綠袍女人傳聞也正式放入 CoAB game-pack `text_rules`：英文片段、
`tilverton.green-robes-rumor` stable ID 與繁中 locale 同時保存。這是把事件文字
從 State fallback 拉回資料層的範例；後續新事件應沿用此方式，不在 State 或測試中
新增中文劇情字串。

## 公會內部與下水道痕跡

第 513 輪從公會 block 2 handoff 的正常回返位置繼續，不再把 `(11,7)`、`(15,7)`、
`(10,13)` 等事件格直接寫入 State。實際玩家交易是：`MoveDungeon` 驗證 GEO 雙側
牆／門後移動，`TurnDungeonWithGrid` 只改變面向，下一次正常移動才讓 ECL
per-turn／search 決定是否出現事件。這個順序很重要；「把角色放到事件格再執行
lifecycle」只能叫 coordinate-assisted probe，不能列入正常路徑證據。

這段的可重用路徑模型是：

```text
公會戰後回返
  → 訪客簿／半身人
  → detail 2 門：撬門／GEO 雙側解鎖
  → 犬舍實戰戰鬥與戰後續跑
  → 猴籠事件
  → 經回廊繞過實心牆
  → 公會內部隨機／提示 boundary
  → detail 2 門
  → 下水道痕跡
```

`GEO2.DAX` 在 `(13,7)` 東側是實心牆，不能因畫面上兩格相鄰就假設可以直走；
正常路徑必須依 cell 的 wall／detail 資料繞行。鎖門的測試使用力量 25 的
確定性 fixture 以確保撬門成功，這只是在測試中提供可重播輸入，不是把該數值寫成
原版開門公式。之後仍由 `UnlockDoorWrapped` 更新兩側 detail，不能只改玩家所在
側的暫存器。

路途中遇到的文字與選項都由 game-pack stable ID 解析：
`tilverton.running-thieves`、`tilverton.option.remain-calm`、
`tilverton.running-thieves-warning`、`tilverton.fire-knives-spot-you`、
`tilverton.guild-assassins-attack`、`tilverton.guild-metal-and-animals` 與
`tilverton.guild-bodies-after-battle`。測試只要求 ID 命中目前 locale，不複製
繁中內容；未來改譯文不應迫使測試同步改字串。

隨機戰鬥必須沿同一 dungeon return mode 續跑。若生命週期中間經過 engine-only
`CALL` 才建立 encounter，單看 `ModeEvent` 的勝利訊息會誤判成戰鬥已回到地城；
remake 需保留 caller 的 dungeon return context，並在勝利後明確 `Continue()`。
這是 runtime transaction 的經驗，不是可寫入 CoAB 劇情 JSON 的旗標。

本輪的終點是 `tilverton.guild-sewer-traces`。選擇繼續並進入 block 3 只能證明
入口 handoff；現有測試在 `(1,8)` 與更深處仍使用座標輔助，因此後續要從 block 3
入口以相同移動交易重新建立完整下水道路徑，不能把本輪標成下水道完成。

## 下水道入口到火刀檢查站

第 514 輪把上述「更深處」的第一段改回正常玩家交易。block 3 入口回返為
`(0,1,S)`，decoded GEO 的合法路徑是：

```text
(0,1) → (1,1) → (1,2) → (1,3) → (1,4)
      → (0,4) → (0,5) → (0,6) → (1,6) → (1,7) → (1,8)
```

每一步都經 `State.MoveDungeon`；不能因已知終點而直接寫入 `(1,8)`。最後一步的
per-turn 邊界會先顯示 `tilverton.sewers.guild-battle-echoes`，中文為「你們仍不時
聽見公會大廳傳來戰鬥聲」。這不是檢查站選單：PRESS 返回地城後，玩家再使用
正式 `LookDungeonLocation`，才會看到 `tilverton.sewers-checkpoint`；舊的
`SearchDungeonLocation` 名稱只保留相容 wrapper。因此 per-turn text、按鍵
continuation 與 explicit LOOK 是三個可追蹤的交易階段，不能在 renderer 或測試中
合併成一次自動觸發。

拒絕檢查站投降後，五名 Fire Knife 由真正 combat turns 完成，勝利後再按下
`tilverton.sewers-hide-bodies` 才返回相同地城 session。這一段證明的是入口到
檢查站與戰後 boundary。第 515 輪已把火刀戰後第一個 `(1,8)`→`(13,10)`
handoff 移入 engine `MapPositionTransition` 與 CoAB JSON event；State 不再由
測試直接寫入騎士座標。這筆轉移目前是 `strong inference`：raw ECL
`CALL 2E10h`、PC-98 movement service、GEO 關閉狀態可行圖分析與既有事件路徑一致，
但 DOS ECL work address 與 overlay 22 `[di+4BF0h]` indexed table 的
writer→projection→consumer 尚未 exact 閉合。不能只因看見同一張
GEO map 就假設左右 component 有開路相連，也不可用 BFS 穿過實心牆。騎士後到
往 block 4 的第二個外部 handoff（下水道 `(10,15,S)`）已接成正常玩家路徑，
逐格落點由 `ECL2/0x03` 的圖緣處理常式直接讀出（spec 1199）；`wall=09` 的原版
writer 仍未 exact 閉合。

## 下水道 E2 與火刀 E1 回返

第 537 輪以 persistent `S` 搜尋模式沿同一 `BlockSession` 完成：

```text
(13,10) W→(12,10) W→(11,10) W→(10,10)
         S→(10,11) S→(10,12)
         （持續 SEARCH 在 (10,12) 發現 game-pack 的 wall=09 候選邊）
         E→(11,12) S→(11,13) S→(11,14) S→(11,15) W→(10,15)
         S→E2 → block 4 (8,1,S)
```

E2 不是前端先包裝的 teleport：`MoveDungeon` 先驗證 pack external exit，再把
邊界嘗試交給 ECL；block 4 入口會保留 `LOAD PIECES 1,2,4`。火刀據點北側
`(8,0,N)` 是第一個短路 E1 候選，玩家先正常處理安靜／刀刃事件，再北行越界，
ECL 回到下水道 `(10,15,N)`。`(11,0,N)` 與 `(13,0,N)` 也已放入同一資料契約，
但目前只以 `strong inference` 標記，沒有宣稱三個原版座標逐像素／逐指令 exact。

火刀首領勝利後的夢境、Tilverton 城外世界選單與 save v12 重訪另有 State integration
回歸；該測試的首領戰是 deterministic fixture，不冒稱已從所有火刀房間逐格走到
首領。完整證據見 [`docs/spec/537-search-look-e2-fire-knife-normal-route.md`](../spec/537-search-look-e2-fire-knife-normal-route.md)。

## 世界地圖 14 點 arrival／路網基線

第 541 輪新增 `TestRealOverlandArrivalAndRouteGraphCoverage`，將原始 ECL1–ECL6
載入同一個測試 session，逐一執行 `moonsea.overland` 的 14 個 native location
arrival。測試同時驗證 JSON directed adjacency 的每個目的地都已宣告，並以 Tilverton
為起點做可達性走訪；這不是把世界地圖當作任意座標游標，也不使用翻譯後文字當
真相來源。

荒野抵達的 adapter 需要在 ECL1 arrival entry 前後提交 `4C9B` native destination，
`4C9C` 則保留為 CoAB arrival selector。若只在 entry 後更新，部分 location 的
route menu 會沿用舊 current-location row；若只在 entry 前更新，ECL 返回後的
save／world state 又可能殘留 dispatcher 暫存值。這項時序是 CoAB adapter contract，
不是獨立 engine 的固定位址語意。

本基線只證明「14 點能抵達、路網不丟點」。各城市的設施、事件、隨機遭遇、地城
入口／出口、世界 travel branch、save／revisit 與劇情旗標仍須由正常玩家 session
逐一驗證；現有事件測試不能被這個 graph gate 取代。

## 從新遊戲接到城市／主線 handoff

第 542 輪提供目前最長的同一 session 範例：

```text
火刀首領勝利
  → PATROL FOREST（保留前置 4C03=0x80）
  → JOURNEY ON → 阿沙本福德
  → ENTER CITY → BAR → RELAX → Tavern Tale 28 → EXIT → LEAVE
  → THE STANDING STONE → TRAIL／提爾隘口戰鬥
  → 灰袍男子 → THANK HIM → 尋紅線索
  → JOURNEY ON → ESSEMBRA → TRAIL → 艾森布拉城外
```

這段的重點不是選項名稱，而是 handoff 的狀態連續性：首領戰後的 block 50 menu
由前置城市事件的共享工作旗標決定；固定只建立首領狀態的測試可能看到
`ENTER CITY`，不能因此清除正常 session 的 `4C03`。城市／世界選項必須從
game-pack stable ID 解析，測試期待值不能複製 locale 顯示文字。

事件驗收要把三種證據分開：

1. **正常 session**：從 `ActionStart`／角色建立或正式 save 開始，沿 `MoveDungeon`
   與世界旅行輸入抵達。
2. **固定 ECL fixture**：只驗證某個 raw block、flag 或戰鬥的 producer／consumer；
   可加速研究，但不得算進「新隊伍從開場到結局」。
3. **coordinate-assisted**：為了縮小地圖事件直接設定座標／旗標；名稱要明示，
   只能作研究或 branch regression，不能寫進 P0 正常路徑完成數。

第 542 輪後，正常骨架已到艾森布拉城外；哈普、熔岩洞、法師塔、尤拉什、摩安德
之坑、散提爾堡與 Myth Drannor 的既有測試仍須逐段轉成同一 session，最後才可宣稱
完整主線。完整證據見
[`第 542 輪正常主線與城市／地城 handoff`](../spec/542-normal-campaign-spine-and-city-dungeon-handoff.md)。

## 目前 CoAB checkpoint

已由新遊戲進入提爾佛頓 GEO2 block 1，使用原始 west step 抵達 Windlord’s Inn，
並回歸圖片、HEAD／BODY 舞台、繁中訊息與 Journal 31。第 511 輪再從正常地城
狀態逐步走過提爾佛頓多個設施，記錄招牌／祭司共享 one-shot 群組；第 512 輪
接續真正 GEO 步行到城門，回歸封路、皇家馬車、衛兵戰與盜賊救援的 continuation。
這段 map／ECL integration 是 `exact`（remake 對原始資料）；State movement
transaction 與 DOS 逐幀 loop 的對應仍是 `strong inference`。其餘由 Tilverton
到結局的路徑仍必須逐段建立，不能把本頁的 checkpoint 寫成完整通關。
