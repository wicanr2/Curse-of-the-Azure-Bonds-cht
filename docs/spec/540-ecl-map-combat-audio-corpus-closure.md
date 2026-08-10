# 第五百四十輪 ECL corpus、全 GEO 宣告與戰鬥音效邊界

日期：2026-08-10

狀態：`READY`（本輪的資料覆蓋與音效生命週期邊界已可驗收；不等於整作已
完成、完整可通關或原版逐 byte／逐幀還原。）

## 本輪關閉的範圍

### ECL corpus 的 parser／控制流覆蓋

- 原始 DOS `ECL1.DAX` 至 `ECL6.DAX` 共 25 個 block、125 個 lifecycle entry
  皆由 `internal/ecl/corpus_coverage_test.go` 讀取。
- 每個 entry 以 `EntryPoints` 取得，交給 `TraceGraph` 建立有界控制流圖；所有
  靜態可達 instruction 都必須在 `KnownCommands` 中，且 command metadata 覆蓋
  `0x00..0x40`。
- 這證明「原始 corpus 可枚舉、entry framing 可解析、控制流 walker 不會把已知
  opcode 誤判成 unknown」，不證明每個 opcode 的外部 routine、畫面副作用、規則
  或正常玩家路徑都已完成。
- 另外的 reachable audit 確認 `0x0F INPUT NUMBER`、`0x1F UNKNOWN_1F` 與
  `0x23 SURPRISE` 沒有從本作 125 個 lifecycle entry 的控制流圖抵達；它們仍
  保留在通用 `KnownCommands` table，不能因「本作未使用」而宣稱通用 engine 已
  支援其完整語意。
- 英文攻略的技術表可作交叉參考：`INPUT NUMBER`、`0x1F` 與 surprise 變體在
  本作的使用狀態不能由線性掃描命中推定；攻略明列其中未用於《青色枷的詛咒》或
  未定義的項目。這是 guide corroboration，不取代本機 bytes／runtime 證據：
  [英文攻略技術章節](https://gamefaqs.gamespot.com/pc/564786-curse-of-the-azure-bonds/faqs/78365)。

### 全部原始 GEO block 的 game-pack 資源宣告

- CoAB game-pack 現在為原始 `GEO2` 至 `GEO6` 的 16 個 block 建立 first-person
  map declaration，並保留 `script_block` 與 `geometry_block` 的獨立欄位。
- ECL3 block `0x12` 與 ECL3 block `0x11` 共用 GEO3 block `0x11` 的映射已明確
  宣告；不能以 script ID 等同 geometry ID。
- `TestPackDeclaresEveryOriginalGEOBlock` 會從原始 ZIP 解析 DAX block，再與
  game-pack 的 `(area_id, geometry_block)` 集合比對。這關閉的是資源／identity
  coverage；每個地形事件、入口出口、世界旅行、持久 map state 與逐區正常路徑
  仍須另外驗收。

### 戰鬥開始與隊伍全滅音效意圖

- ECL encounter 進入 `StartEncounterWithAffects` 後排入 `SoundCombat`。
- `PROGRAM 3`（隊伍全滅）排入 `SoundCrash`，並由測試確認事件只消費一次。
- DOS selector `14／15` 已保留在 sound catalog；目前沒有對應的抽取 WAV，
  因此 DOS adapter 會安全略過。PC-98 的 `COMBATFX=14`、`CRASHFX=15` 映射
  已由既有 PC-98 sound spec 保存，semantic event 不在 State 內硬編 selector。
- 這補的是事件生命週期的 producer→platform intent 邊界；不宣稱 DOS／PC-98
  原機混音、音量、等待時間、每一個戰鬥 phase 的 cue 或所有場景音樂已 exact。

## 尚未關閉的邊界

1. `CALL`／`PROGRAM`／`NEWECL` 與地圖服務的完整 producer→state→consumer，
   以及從開場到結局的正常玩家 session。
2. 16 個 GEO 的所有房間、地形事件、外部出口、世界旅行、重訪與存檔持久化；
   block 宣告不等於全地圖可玩。
3. 完整 AD&D 戰鬥 AI、弓箭與每種法術的 saving throw、持續區域、死亡動畫、
   ECL handoff、逐幀 timing 與音效次序。
4. DOS／PC-98 的完整音樂與音效 producer、原硬體 timing／mixer 校準，以及全量
   繁中資料與 UI fidelity。

## 驗證

Docker 內已通過 focused，並在修正音效事件測試期望值後通過正式受影響套件：

```text
go test -modfile=workplace/coab-test.mod ./gamepack ./internal/ecl ./internal/sound ./internal/game
go test -modfile=workplace/coab-test.mod ./cmd/... ./gamepack ./internal/...
```

其中 `cmd/azure-bonds-game` 以容器內明確啟動的 Xvfb 執行；本規格不把套件 gate
擴大成整作完成證據。
