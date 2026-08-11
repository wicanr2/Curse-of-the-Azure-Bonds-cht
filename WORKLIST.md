# 《青色枷的詛咒》目前工作清單

更新日期：2026-08-11（第 547 輪後盤點）

本檔是 compact、交接與每輪開工時的目前摘要入口。詳細的反組譯證據仍在
[`docs/knowledge/golden-box-reverse-engineering-worklist.md`](docs/knowledge/golden-box-reverse-engineering-worklist.md)；
可驗證的歷史與每輪規格見 [`docs/project-status.md`](docs/project-status.md)。
本檔只保留目前有效的工作，不把歷史輪次的舊 blocker 當成現況。

## 一句話結論

重製尚未完成整作通關。現在已經有多條可重播的正常玩家垂直切片，並完成
`SEARCH`／`LOOK`、wall=09 候選橋接、E2、火刀 E1、戰後世界地圖與 save/load
的 engine＋JSON 接線；本輪再完成 25 個 ECL block／125 個 entry 的 parser／控制流
稽核、16 個原始 GEO block 的 game-pack 宣告、ECL 戰鬥開始／隊伍全滅音效意圖，
以及 14 個世界點位的 ECL1 到達／JSON 有向路網基線；第 542 輪把同一新遊戲
session 從火刀首領後接到阿沙本福德城內、立石群與艾森布拉城外；第 543 輪再把
同一 session 接到 Hap 村落、熔岩洞、巫師塔、回洞穴與熔岩池第二次戰鬥。仍缺
完整 ECL side effects／外部 routine、全城市／全房間 coverage、完整結局同
session、完整戰鬥與原機音訊、全量繁中校對、完整存檔相容與三平台發行。

第 547 輪以 DOS IDA Pro 原始 bytes 閉合 `C04B..C04F` 的虛擬地圖 bridge：
`C04B/C04C/C04D` 是 live map 位置／half-facing 的 ECL adapter，`C04E/C04F` 是
wall／terrain cache 的讀取介面。這修正先前錯誤的 Cave E1 `(4,5,N)` 錨點：同一
新遊戲 session 現在由 E1 `(5,7,W)` 一般走格到 `(5,9)`，再由 ECL4 block `0x22`
的位置交易及 JSON `set_map_position` 抵達死精靈格 `(13,1,W)`，保留原始
`wall=08`／`terrain=C0`。這不是完整洞穴：Dexam、其他傳送／隨機事件、出口與
重訪仍待正常路徑驗證；`0x4C00` 仍是與目前玩家結果無關的 `unknown`，不列為
remake blocker。

## 狀態與證據規則

- `已完成（remake contract）`：重製程式、JSON、測試與玩家路徑已閉合；不代表
  原版每個 byte 已逐一證明。
- `exact`：原始 bytes 與 consumer／runtime trace 已閉合。
- `strong inference`：多項證據一致，但仍少一段原版資料流或 runtime oracle。
- `待實作`：目前玩家路徑或產品功能仍未完成。
- `待研究`：只有在要支援該功能或原版 fidelity 時才逆向；不可先把假說寫入
  正式規則。

## 第 540 輪已關閉的工作

| 項目 | 目前可宣稱的範圍 |
|---|---|
| ECL corpus parser／控制流稽核 | 原始 `ECL1..6.DAX` 的 25 個 block／125 個 lifecycle entry 都可由 `EntryPoints` 取得並交給 `TraceGraph`；靜態可達 opcode 都有 command metadata，`0x00..0x40` table 也有 coverage。這不是完整 opcode side effect、外部 routine 或整作通關。 |
| 全原始 GEO block 的 game-pack identity | 16 個原始 `GEO2..6.DAX` block 都有 first-person declaration，且 `script_block`／`geometry_block` 分離；ECL3 `0x12` 共用 GEO3 `0x11` 的幾何也有明確映射。這不是所有地形事件、出口、世界旅行或持久重訪。 |
| 戰鬥開始／隊伍全滅音效 intent | ECL encounter 進入戰鬥排入 `SoundCombat`，`PROGRAM 3` 排入 `SoundCrash`；PC-98 selector 對應留在 adapter，DOS 缺少 14／15 WAV 時安全略過。這不是完整原機音效、混音、時序或全場景音樂。 |

權威規格：[`第五百四十輪 ECL／GEO／戰鬥音效邊界`](docs/spec/540-ecl-map-combat-audio-corpus-closure.md)。

## 第 541 輪已關閉的工作

| 項目 | 目前可宣稱的範圍 |
|---|---|
| ECL 外部 routine 分層 | 已將 `CALL`／`NEWECL`／`PROGRAM`／資源請求／typed side-effect request 與 CoAB address／caller context 分開；共用 engine 保留有序 raw／typed boundary，`0x2E10`、`0xC01E`、`0xB200`、`PROGRAM` 語意不下沉成跨作品事實。完整決策與推論等級見 [`spec 541`](docs/spec/541-ecl-external-routine-engine-boundary.md)。 |
| 全世界點位到達 | `TestRealOverlandArrivalAndRouteGraphCoverage` 由原始 ECL1 arrival entry 執行 `moonsea.overland` 全部 14 個 native location values，並驗證 Area／Location／world state 投影。 |
| 世界旅行路網 | JSON adjacency 的所有 destination 都有宣告，且從 Tilverton 的 directed graph 可達全部 14 點；`arriveAtWorldLocation` 在 ECL1 entry 前後提交 `4C9B`，修正部分抵達後沿用舊城市路由列的 bug。 |

這不是「全地圖事件完成」：所有城市設施、隨機遭遇、區域／地城房間、出口、重訪
旗標與完整主線仍要沿正常輸入逐區驗證；既有城市事件測試與後續 vertical slice
繼續累積，不能用路網可達性替代劇情事件證據。

## 第 542 輪已關閉的工作

| 項目 | 目前可宣稱的範圍 |
|---|---|
| 火刀首領後正常世界出口 | 同一新遊戲 ECL session 完成戰後夢境、`PATROL FOREST` 戰鬥續跑、`JOURNEY ON` 與阿沙本福德抵達；`4C03=0x80` 的前置共享旗標被保留，沒有用 frontend 特判清除。 |
| 阿沙本福德正常城市 handoff | 正常抵達後進城、進河畔酒館、選 `RELAX` 觸發 Tavern Tale 28、按鍵續跑、離開酒館與離城均由 game-pack stable option ID 驅動。 |
| 立石群／艾森布拉正常主線骨架 | 阿沙本福德離城後沿 `THE STANDING STONE`，完成提爾隘口戰鬥、灰袍男子／尋紅線索，再沿 `ESSEMBRA` 到達城外 edge；同一 ECL session 未重播 block 起點。 |
| 固定事件與正常路徑的證據分層 | 長固定整合測試仍涵蓋哈普、熔岩洞、法師塔、希爾斯法、尤拉什、摩安德之坑、散提爾堡等大量事件；第 542 輪規格明確標出它們不能取代一條從新遊戲到結局的正常 session。 |

權威規格：[`第 542 輪正常主線與城市／地城 handoff`](docs/spec/542-normal-campaign-spine-and-city-dungeon-handoff.md)。

## 第 543 輪已關閉的工作

| 項目 | 目前可宣稱的範圍 |
|---|---|
| Hap／熔岩洞／法師塔正常 session | `TestRealNewGameContinuesFromHapToBeholderCaveEntrance` 從第 542 輪的艾森布拉城外沿正常世界路由進入 Hap，以 `MoveDungeon` 逐格完成民宅、阿卡巴、旅店、伊弗利特與村落出口；經 JSON external exit 進入熔岩洞，完成伏擊／守門戰；到 `(6,15,W)` 進巫師塔，完成庭院、德拉坎德羅斯、黑龍敘事、攻擊巫師與屋頂 CAVES 回程；同一 session 再完成熔岩池 `WAIT→PARLAY_NICE`、重訪 `COMBAT`、15 隻火蜥蜴、防火桶 WHO 與耐熱失敗，並沿正常故事 handoff 經 Cave E1 `(5,7,W)` 到死精靈格 `(13,1,W)`。未直接寫入劇情旗標、未 direct-entry 戰鬥；Dexam／洞穴其他事件尚未納入此 gate。 |
| ECL 故事重繪座標邊界 | `original.geo5.block-33` 由 JSON 宣告 `spawn=(7,15,W)`；engine 以 map anchor 保留地城 live cursor，避免同區塊 ECL redraw 的暫存 `C04B/C04C/C04D` 改寫玩家位置。這是可重用契約，不是 CoAB 座標特判。 |
| 外部出口 presentation selector | 獨立 engine 新增 `ExternalExitDefinition.RoofType` 與 schema；CoAB Hap `(15,5,E)` 宣告 `roof_type=2`，正常邊界使用 `wall_type`／`roof_type` 而不把原始 GEO terrain 假稱 exact。engine 本地提交為 `9cf5fa5`；GitHub push 需待外部目的地審核通過。 |
| 正常路徑與固定夾具分層 | 新增 spec 543，明確把同一 session coverage 與既有固定 Hap／Myth Drannor／`PROGRAM 8` fixture 分開；目前仍不能宣稱全城市、全地城或整作結局。 |

權威規格：[`第五百四十三輪正常主線 Hap／熔岩洞／法師塔 coverage`](docs/spec/543-normal-campaign-coverage-and-ida-map-cell-audit.md)。

## 第 544 輪已關閉的工作

| 項目 | 目前可宣稱的範圍 |
|---|---|
| opaque raw memory engine contract | Golden Box engine 新增 `set_memory`／`Runtime.MemoryWrites`／schema validation；CoAB JSON 宣告 `zhentil-keep.inner-city.route-memory-reset`。`0x4C00` 原版語意仍是 `unknown`，不列為 D&D 規則，也不再為它建立反組譯 blocker。 |
| 散提爾堡至眼魔洞穴死精靈格 | 同一正常新遊戲 session 通過 Olive、暗神殿、Dimswart、hooded woman 與 block `0x22`／GEO4 `0x25` handoff，抵達 Cave E1 `(5,7,W)`，再以 ECL 的 `C04B/C04C/C04D=13/1/3` 位置交易到 `(13,1,W)`；測試沒有 `0x4C00`／`4BF2` 直接注入。 |
| 洞穴驗收邊界 | 依公開攻略只保留 Dexam／傳送／隨機事件作 nearby 導航線索；未把攻略座標直接寫成 GEO edge，洞穴內部房間、戰鬥續跑與出口不宣稱完成。 |
| 角色 saving throw 資料分層 | 五欄 saving throw threshold 已由 character template JSON、engine schema 與 State projection 驗證；不與 `0x4C00` 混為同一語意問題。 |

權威規格：[`第五百四十四輪 raw memory route boundary／Dexam 洞穴入口`](docs/spec/544-opaque-memory-route-boundary-and-cave-entry.md)。

## 第 547 輪已關閉的工作

| 項目 | 目前可宣稱的範圍 |
|---|---|
| DOS `C04B..C04F` virtual-map bridge | IDA overlay-07 原始 bytes 證明 setter 將 `C04B/C04C/C04D` 對應到 `DS:720F/7210/7211`，getter 將 `C04E/C04F` 讀回 `DS:7212/7213`；既有 vector 4／6 與 GEO four-plane evidence 將後兩者接到 wall／terrain cache。這消除玩家路徑所需的 map-register blocker，但不宣稱每個 renderer dirty flag 或 DOS redraw 都已 pixel-exact。 |
| E1 至死精靈格正常 handoff | 正常新遊戲 session 由 `(5,7,W)` 一般走到 source cell `(5,9)`，game-pack 只投影原始 ECL position transaction，最終為 `(13,1,W)`、`wall=08`、`terrain=C0`，並驗證 `C04B..C04F`／event exactly-once。這不含 Dexam 雙戰、出口、重訪或完整通關。 |
| `0x4C00` 範圍 | 本輪沒有新增 `0x4C00` 逆向或命名；依第 546 輪結果維持 `unknown`，只要不影響玩家可見 D&D 規則／路由／存檔，就不列為 remake blocker。 |

權威規格：[`第五百四十七輪洞穴位置交易畫面狀態`](docs/spec/547-normal-beholder-cave-presentation-state.md)。

## 第 539 輪已關閉的工作

| 項目 | 目前可宣稱的範圍 |
|---|---|
| 火刀據點入口→首領戰前正常路徑 | `TestRealNewGameBeginsAtGlobalBlockOne` 從真實開場、下水道、E2 block 4 `(6,1,S)` 出發，以同一個 `MoveDungeon`／ECL session 走 29 步到 `(3,13)`、terrain `0x87` 的首領事件；路徑實際經過 `0x99` 刀刃區、`0x9A` 冰凍房與 `0x94／0x95` 相位蜘蛛區，完成必要的選單／戰鬥／續跑後抵達首領戰前。測試以 stable message ID 與敵方物件資料驗證 20 名火刀＋首領，共 21 名敵人；沒有直接注入座標或直接進入首領。原版對照等級是 `layout／route reconstructed`，不是整張地圖 pixel-exact。 |
| 中文 GUI 溢框修正 | renderer 現在依倚天字形的實際 glyph advance 做 rune-safe 換行與單行裁切，主要冒險、地城、事件、手札、建立角色與戰鬥欄位都使用明確可用寬度；中文字不再以固定英文字元數硬塞。Docker／Xvfb 重新產生 640×480 冒險與第一人稱截圖，並替換 README 仍引用的舊版代表圖。這是 `layout-reconstructed`，不是所有狀態逐像素 exact。 |
| 讀檔位置投影 | `LoadPartyFile` 依保存的 `Area.CurrentCity` 重建 `LocationName`／`OriginalLocation`，不再只恢復數字 enum；火刀首領固定 fixture 的 save/load 另驗證原始地點保留。正常完整 session 的戰後世界選單尚未因此宣稱閉合。 |

權威規格：[`第 538 輪火刀入口至首領路徑`](docs/spec/538-fire-knife-normal-leader-route.md)、
[`第 539 輪中文 GUI 寬度與溢框`](docs/spec/539-cjk-gui-width-clipping.md)。

## 第 537 輪已關閉的工作

| 項目 | 目前可宣稱的範圍 |
|---|---|
| engine＋JSON 的 `SEARCH`／`LOOK` 分離 | `S` 持續切換 `DungeonSearchEnabled`，`L` 是一次性 `LookDungeonLocation`；地圖 edge、external exit、locale 與 save v12 均有資料契約。這是重製契約已完成；原版 Search 成功率與 wall writer 仍未 exact。 |
| 下水道至 E2 | 從 `(13,10)` 正常逐格移動，經 wall=09 候選橋接抵達 `(8,15,S)`，再由 E2 進 ECL2 block 4；未用直接設定座標完成這條重製路徑。 |
| 火刀 E1 回返 | block 4 北側 E1 候選可由正常移動越界，回到下水道 `(10,15,N)`；三個 E1 座標仍屬 `strong inference`。 |
| 戰後 handoff 與存檔 | 首領勝利後的 ECL 夢境、Tilverton 世界地圖選單與 Search／edge 狀態 save/load 已有固定 fixture 回歸；第 539 輪另外完成入口到首領戰前的正常路徑，但完整 session 在首領戰後仍會出現原始 `PATROL FOREST／JOURNEY ON／CAMP` 分支，尚未與預期的世界地圖出口契約閉合。 |
| 既有資料分層 | 開場選項、玩家法術入口與多個事件已使用 stable ID、locale JSON 與 engine resolver；不能因此宣稱全量中文化。 |
| UI 基線 | 640×480、原版裂紋石框、倚天粗體 16×15、人物 HEAD／BODY 分層與多張對照截圖已建立；目前多為 `layout-reconstructed`，不是所有狀態逐像素 exact。 |

權威規格：[`docs/spec/537-search-look-e2-fire-knife-normal-route.md`](docs/spec/537-search-look-e2-fire-knife-normal-route.md)。

## 剩餘工作總表

### P0：先讓主線繼續走，不再用座標輔助掩蓋缺口

| 工作 | 現況 | 下一個可驗收成果 |
|---|---|---|
| 火刀據點完整正常路徑 | 入口→首領→戰後世界→阿沙本福德→立石群→艾森布拉城外的正常 session 已接通；仍未覆蓋所有可選房間、全部寶物、失敗／重訪分支。 | 以原始 GEO 路徑補齊火刀可選房間與重訪，再把可驗收結果寫入 coverage matrix。 |
| 火刀據點出口、返回世界地圖與重訪 | 正常 session 的 `PATROL FOREST`、`JOURNEY ON`、阿沙本福德抵達與後續城市 handoff 已閉合；Tilverton 固定 fixture 的 save/load 回歸仍保留。 | 將同一正常 session 的存檔／重載與重訪延伸到世界路由，並分離固定夾具與正常主線證據。 |
| 開場到結局的正常玩家主線 | 已從開場走到眼魔洞穴死精靈格 `(13,1,W)`；固定事件測試另覆蓋哈普、法師塔、摩安德之坑、散提爾堡與 Myth Drannor 局部，但 Dexam 正常 handoff、完整章節、最終戰與結局尚未串完。 | 先從 `(13,1,W)` 以正常輸入接通 Dexam 與洞穴出口，再沿同一 session 續接尤拉什／摩安德之坑→散提爾堡→Myth Drannor，最後加入 `PROGRAM 8` 結局與 save/reload gate；未支援 boundary 必須明確 fail-closed。 |

### P1：補齊可玩規則、資料與原版行為

| 工作 | 目前缺口 | 驗收方向 |
|---|---|---|
| 全 ECL 與外部 routine | 25 個 block／125 個 entry 的 parser／控制流 corpus gate 已完成；`C04B..C04F` virtual-map adapter 已閉合，但 `CALL`、`NEWECL`、剩餘地圖服務、劇情旗標、NPC 離隊、輸入與 continuation 的完整 consumer 仍未閉合。與玩家結果無關的 raw work address（目前如 `0x4C00`）不列為 blocker。 | 只逆向會改變玩家結果的 producer→state→consumer；每個完成事件都要有 raw bytes／runtime trace、JSON contract、stable ID 測試與正常輸入路徑。 |
| 全地圖與世界旅行 | 16 個原始 GEO block 已在 game-pack 宣告；14 個世界點位的 ECL1 到達、Tilverton→全點 directed adjacency，以及新遊戲→阿沙本福德→立石群→艾森布拉的正常主線已通過 Docker gate。仍缺所有城市／地城房間 coverage、TRAIL／WILDERNESS／EXIT 全分支、隨機遭遇、所有入口出口、持久 map state 與原版 fidelity。 | 建立每座城市／每個 GEO block 的正常事件 coverage matrix，保存 flag／座標／資源 handoff，並補全世界旅行與重訪回歸；不把攻略座標直接寫成規則。 |
| 戰鬥規則、AI、法術效果與動畫 | 已有部分 AD&D 數值、敵方選敵、延後施法、12 個玩家法術入口與視覺時間軸；ECL 戰鬥開始／隊伍全滅音效 intent 已接通，仍缺完整敵我 AI、弓箭／投射物、法術逐項效果、saving throw、持續區域、死亡動畫、回合節奏與原機音效 cue。 | 對近戰、弓箭、Magic Missile、Fireball、Lightning Bolt、Stinking Cloud／Cloudkill 等分開驗收 windup、travel、impact、damage、save、death、persistent effect、聲音與 ECL handoff；影片只能證明演出，數值要回 bytes／DOSBox。 |
| 存檔、角色規則與跨遊戲轉移 | remake save v12 已保存 Search／edge；DOS／PC-98 `SAVGAM`、角色 sidecar、完整 record、年齡／職業／特殊能力、刪除／改名與 `MOVEPARTY` 跨遊戲 transfer 尚未完整 round-trip。 | 先完成版本化 parser／serializer 與 save mutation diff，再以角色檔跨 Gold Box 來源做 stable transfer contract；不能把 `MOVEPARTY` 靜態 helper 直接當秘密門。 |
| 全量繁體中文化與遊戲內手札 | 多個事件、選項、手札與攻略已資料化；仍缺全 ECL／物品／法術／怪物／地名／UI 字串、中文校對、長文分頁與所有原版需查手冊的內容整合。 | 以 stable `message_id` 做 coverage／orphan／source-drift audit；玩家可在遊戲內讀手札，不要求查外部文件才能通過事件。 |
| 音樂與音效 | YM2203、S98、PC98 sound BIOS、cycle PCM 等 engine 知識與部分合成測試已有；戰鬥開始／隊伍全滅 semantic intent 已接通，但完整 DOS／PC-98 producer、播放生命週期、音效與戰鬥 phase 同步仍未完成。 | 先完成每個場景／戰鬥 cue 的資料綁定與可重播播放，再用 DOS／PC-98 runtime 對照 phase、音量、音效次序；合成器測試不能冒稱硬體 exact。 |
| UI、素材與原版 fidelity | 本輪補上依實際 glyph advance 的中文換行／裁切，並重新驗證原版裂紋石框、640×480 第一人稱與右側 party/status；README 代表圖已換成目前版本。冒險／戰鬥／地圖／對話／頭像的所有狀態仍未逐張比對，palette cycle、sprite timing、左上場景填格與 PC-98 密度仍需抽樣校準。 | 每張對照標示平台、狀態、save／seed、theme 與 `exact`／`nearby`／`layout-only`；原版 theme 與美化 theme 分開，先完成原版忠實驗收。 |

### P2：完成後才做的發行工作

| 工作 | 門檻 |
|---|---|
| Windows／Linux／macOS 打包 | P0 主線、P1 規則／資料／音訊與存檔通過後，才做三平台可重現 build、資產授權檢查、存檔位置與首次啟動 smoke。 |
| README／截圖／40–60 秒推廣片 | 截圖只使用目前版本可重播狀態；推廣片只在可玩整合完成後製作。8 小時錄影不是本專案目標。 |
| 日後美化 theme 與 donate | 原版忠實 theme 永久保留；美化與 donate 只作後續／local 設定，donate 資訊不得上傳 GitHub。 |

## 仍需逆向，但不應阻塞目前 remake 路徑的項目

這些是原版 parity 或跨作品知識庫工作，不是重新打開第 537 輪已接通的路徑：

1. `wall=09` 第三平面 before／after writer、Search 成功率、原版 E1 精確座標與
   同版重訪 trace。現有 CoAB edge 是 `strong inference`，不要擴大成所有
   `wall=09` 都可走。
2. DOS `CALL 2E10h` 的剩餘 redraw／dirty-flag 時序與 runtime pixel trace。第 547
   輪已關閉 `C04B..C04F` 對 `DS:720F..7213` 的 adapter；這項只屬原版 fidelity
   稽核，除非某個玩家結果被阻塞，不再逐行追無關 overlay。
3. PC-98 `MOVEPARTY` 的角色轉移 selector／record／save round-trip。中文說明書
   已證明產品功能邊界，但尚未證明每個 raw helper 與 transfer record 的一對一
   runtime 對應。

## 明確不做的事情

- 不為了湊「完整反組譯」而逐行解讀與玩家結果無關的 function。
- 不把 `BDF1`、`SEARCHREC`、`MOVEPARTY`、相同十六進位數字或單一 xref 重新命名成
  秘密門、detail、年齡、旗標或地圖 owner。
- 不以 direct-entry、固定座標、注入戰鬥、測試模式或窄測試宣稱完整通關。
- 不在 JSON 尚未成為真相來源前，把劇情文字、裝備、法術或測試期望值硬編碼回 Go。
- 不在遊戲完整可玩前花時間做三平台 release 或長篇推廣影片。

## 完成聲明的共同驗收門檻

至少要通過：

1. 新隊伍／正式角色能由開場以正常輸入走到結局，包含移動、互動、裝備／使用、
   戰鬥、治療／休息、存檔、退出、重載與一個後期任務重訪。
2. 所有已宣稱支援的內容由 CoAB JSON／locale 與 engine contract 驅動；未支援
   行為明確失敗即關閉，不以 fallback 假裝完成。
3. 原版／remake 的畫面、動畫、音樂、音效、規則與存檔比較都有證據等級；近似畫面
   不標成 pixel-perfect。
4. Docker 內完成受影響套件、代表性正常玩家路徑、save round-trip、截圖／包裝 smoke；
   再集中 commit＋push 兩個 repository。

下一個最小可重現工作：沿第 543 輪同一個 ECL session 從防火桶返回的熔岩洞續接
尤拉什、摩安德之坑、散提爾堡與 Myth Drannor；同時從 Dexam 洞穴 `(13,1,W)`
以原版 ECL selector／正常輸入建立死精靈、皮袋與手札 59 的 coverage，再建立
城市／GEO 事件 coverage matrix，逐項標示 normal、fixed fixture 或
coordinate-assisted。不要把本輪
Hap／巫師塔局部路徑或既有 `PROGRAM 8` fixture 擴大解讀成完整結局。
不要把 static corpus／路網 gate 擴大解讀成完整 ECL，也不要先深挖與玩家結果無關的
反組譯。
