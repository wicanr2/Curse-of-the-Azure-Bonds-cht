# 第五百四十四輪：不命名 raw memory 的玩家路徑邊界與 Dexam 洞穴入口

狀態：`READY`  
日期：2026-08-11

> 勘誤（第 547 輪）：本規格的 `0x4C00`／`set_memory` 結論仍有效；洞穴入口
> `(4,5,N)` 已被更正為 Cave E1 `(5,7,W)`，並由
> [`spec 547`](./547-normal-beholder-cave-presentation-state.md) 取代其
> `C04B..C04F`／傳送描述。

## 結論

本輪的目的不是替每個 ECL work address 找一個看似合理的名稱，而是讓重製在
不污染 D&D 規則的前提下繼續可玩。

- `0x4C00` 的原版語意仍是 `unknown`。目前證據只支持它會影響散提爾堡
  Olive 故事路徑的前置狀態；沒有證據支持它是年齡、能力、命中、AC 或任何
  AD&D 規則欄位。
- engine 新增作品中立的 `set_memory` action 與 `Runtime.MemoryWrites`。
  它只接受 `address`／`value`，不替 raw word 取語意名稱。
- CoAB game pack 以
  `zhentil-keep.inner-city.route-memory-reset` 宣告
  `ECL block 0x20 ∧ [0x4C00]=2 → [0x4C00]=0`。這是重製的玩家路徑契約，
  不是對原版欄位語意的宣稱。
- `TestRealNewGameRunsToTheEnding` 已由同一新遊戲
  session 正常抵達 Cave E1 `(5,7,W)`，並接通原始 ECL 寫出的死精靈格
  `(13,1,W)`。完整洞穴、隨機遭遇與出口仍不以猜測 GEO 邊連線取代。

## 證據與推論等級

| 項目 | 等級 | 證據 | 可宣稱範圍 |
|---|---|---|---|
| `0x4C00` 原版欄位名稱 | `unknown` | 目前只有 raw ECL／runtime work value，尚未閉合 writer→projection→consumer | 不可命名為年齡、規則或永久旗標 |
| Olive 路徑對 `[0x4C00]=0` 的依賴 | `strong inference` | 移除直接測試注入後正常路徑停在 Olive；由 JSON 宣告 raw transition 後同一 session 通過 Olive、暗神殿與後續 handoff | 支持重製需要保存此 raw route dependency；不支持原版語意命名 |
| engine `set_memory` | `exact`（remake contract） | engine schema、loader、Runtime test 與 State 寫回測試 | 作品包可宣告不透明的 VM word 寫入 |
| Cave E1 `(5,7,W)` | `exact`（remake path contract）／`strong inference`（原版座標／朝向對照） | `zhentil-keep.beholder-cave` 的 JSON spawn、ECL block `0x22`／GEO4 block `0x25` handoff、同一 session 測試 | 可宣稱 E1→死精靈格已接通；不能宣稱洞內全路徑 |
| 洞穴內傳送與房間順序 | `nearby`／`layout-only` | 公開攻略描述 Cave E1 的傳送與隨機遭遇，但攻略座標系統不等於本機 GEO 座標 | 可用來安排下一個玩家驗證，不可直接新增 map edge |

## 實作邊界

### Golden Box engine

`golden-box-remake-engine/engine/pack.go` 的 `Action` 增加 `address`／`value`；
`engine/runtime.go` 將 action 結果放入 `MemoryWrites`。`Pack.MemoryAddresses()`
也會收集 action address，使 host adapter 能在不複製整個 VM 的情況下取得必要輸入。
schema 只驗證 16-bit address／value，不知道 `0x4C00` 的作品語意。

engine 測試 `TestMemoryActionEmitsOpaqueWrite` 驗證：

1. JSON 可載入 `set_memory`。
2. predicate 命中後只產生 raw write。
3. write address 可被 `MemoryAddresses()` 發現。
4. action 不會自行切換地圖、戰鬥或 D&D 規則。

### CoAB adapter

`internal/game/state.go` 只把 `Runtime.MemoryWrites` 寫回目前
`BlockSession`；若 action 沒有 `mode`，保留玩家目前的 mode／地圖／選單。故事
條件、ECL block 與 raw address 都在 `gamepack/events/pit-of-moander.json`，
沒有把 `0x4C00` 寫入通用 State 的劇情 `switch`。

角色建立的 saving throw 是另一條已閉合的資料路徑：五個原始 saving throw
threshold 放在每個角色 template 的 `saving_throws`，engine schema 驗證五欄，
`creation.go` 只投影資料。它與 `0x4C00` 沒有語意關係，不應互相推論。

## Dexam 洞穴的停止點

正常 session 已驗證：

`開場 → Hap → 熔岩洞 → 法師塔 → 熔岩池 → 世界旅行 → 散提爾堡 → Olive →
暗神殿 → Dimswart → hooded woman → ECL block 0x22／GEO4 block 0x25 → Cave E1
(5,7,W) → 死精靈格 (13,1,W)`。

公開攻略可作下一輪導航線索：

- [GameFAQs Curse of the Azure Bonds walkthrough](https://gamefaqs.gamespot.com/pc/564786-curse-of-the-azure-bonds/faqs/78365)
- [GameBanshee Beholder Cave walkthrough](https://www.gamebanshee.com/curseoftheazurebonds/walkthrough/beholdercave.php)

這些資料描述洞穴中可能有入口傳送、隨機戰鬥與 Dexam 房間，但其座標是攻略
自己的表示方式。下一輪要用 DOS／remake 正常移動、`SEARCH`／`LOOK`、事件續跑
與戰後 handoff 驗證後，才可新增 JSON `search_edges`、`external_exits` 或房間
事件。不能把 `(0,0)→(13,1)` 等攻略數字直接當成 GEO4 的 native edge。

## 驗證

Docker 內已通過：

- engine `go test ./...`。
- CoAB `go test ./gamepack ./internal/...`。
- 代表性正常玩家路徑：
  `TestRealZhentilOliveSecretPassage`、
  `TestRealZhentilRecruitDimswart`、
  `TestRealZhentilHoodedWomanReachesBeholderCave`、
  `TestRealBeholderCaveDexamAndZhentilBattles`、
  `TestRealNewGameRunsToTheEnding`。

Ebiten command package 仍須以有界 Xvfb 做獨立 build／smoke；它不是本輪 raw
memory 或 D&D 規則證據。

## 明確不做

- 不為 `0x4C00` 建立 `age`、`alignment`、`story_flag` 或其他語意名稱。
- 不因攻略畫出洞穴圖就宣稱完整 native GEO 路徑。
- 不把抵達入口的測試擴大成 Dexam 洞穴全房間、完整戰鬥、全地圖或整作結局。
- 不把 `set_memory` 當成 D&D 規則 API；若未來證據顯示某 word 影響戰鬥／存檔，
  才另開有 raw 位址、consumer 與推論等級的規格。
