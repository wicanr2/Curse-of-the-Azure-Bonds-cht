# 第五百四十一輪：ECL 外部 routine 副作用與全世界旅行邊界

狀態：`READY`（engine／CoAB adapter 分層與世界到達／路網驗證）

日期：2026-08-11

## 結論

本輪把 ECL 外部 routine 的責任切成兩層：

1. 共用 engine 只保存可跨 SSI Gold Box 重用的「有序邊界訊號、續跑狀態、資源
   選擇、資料請求與驗證契約」。
2. CoAB game pack／State adapter 保存 raw address、PROGRAM caller context、
   work address、劇情旗標、城市／地城語意與文字。

因此本輪**沒有把 `0x2E10`、`0xC01E`、`0xB200` 或 `PROGRAM` 數字語意硬塞進
獨立 `golden-box-remake-engine` repo，也沒有因地址相同就宣稱其他 SSI 遊戲共用
同一 routine。現有 engine 的 `MapDefinition`、`Event`、`MapPositionTransition`、
`SearchEdge`、`ExternalExit` 與 ECL typed signal 已足夠承載目前證據；尚無第二款
遊戲的 producer→consumer 證據支持新的通用地址 API。

同輪完成：

- `arriveAtWorldLocation` 在 ECL1 到達 entry 前後提交原生目的地 `4C9B`，避免部分
  世界點位沿用舊城市的路由列；
- `TestRealOverlandArrivalAndRouteGraphCoverage` 從原始 ECL1 到達 entry 執行全部
  14 個 `moonsea.overland` 原生點位；
- 同一測試驗證 JSON 有向 adjacency 的所有目的地都已宣告，且從 Tilverton 可達
  全部 14 點；測試不複製翻譯文字，透過 game-pack stable ID／native value 驗證；
- 已有各城市／地城正常 ECL tests 繼續負責事件、戰鬥、房間與 continuation。本輪
  不把 14 點到達／路網完整擴大成「所有地圖事件或整作通關完成」。

## 證據與位址空間

| 證據 | 可證明內容 | 等級 |
|---|---|---|
| `docs/spec/276-ecl-call-state-adapter.md`、`internal/ecl/runtime.go` | `CALL` 保存 raw operand、PC、block；`NEWECL` 保存下一個 PC 與目標 block；`PROGRAM` 保存 ID 並在已知 boundary 停止 | `exact`（VM framing／continuation） |
| ECL1–ECL6 25 個 block／125 個 entry corpus gate | 本作生命週期 entry 能被 parser／control-flow trace 建立；不等於每個 external side effect 已完成 | `exact`（corpus） |
| `CALL 0x2E10` → `AE11h` → overlay control vector | raw dispatch／map redraw candidate；尚未閉合完整 DOS producer→projection→consumer | `exact` raw，`strong inference` routine role |
| `CALL 0xC01E` → forced forward movement | CoAB adapter 的 cardinal wrapped forced move；不是玩家 wall collision transaction | `exact` remake contract；原版 routine 行為依既有 spec |
| `CALL 0xB200` | external sound boundary；`word_1EE76` 選 A/B 的 B 分支尚未投影 | `strong inference` |
| `PROGRAM 0/3/8/9` | ID 與 external boundary 可保存；訓練、全滅、勝利、CAMP 由 CoAB caller context 解讀 | `exact` boundary；ID 語意為 CoAB adapter |
| `LOAD FILES`／`LOAD PIECES` | 三個 raw selector 可保存並交給 title resource adapter | `exact` request；實際資源 identity 屬 game pack |
| `moonsea.overland` 14 點與 adjacency | native value、點位與 directed destinations 可由 JSON 驗證；不是自由 25×50 tile map | `exact` game-pack contract；原版像素地圖仍另驗證 |

## 外部 routine 副作用決策

### 應放入共用 engine 的部分

| 機制 | engine 形式 | 原因 |
|---|---|---|
| ECL 續跑 | `BlockSession` 的 PC、block、memory、menu／combat／treasure continuation | 各 Gold Box 都需要跨 pause 保存控制流；不含 CoAB 劇情名稱 |
| 有序 external call | `CallRequest{Address, PC, BlockID}` 與未知地址保留 | engine 可保存觀察到的訊號；不猜地址語意，其他作品可自行註冊 adapter |
| `NEWECL` 邊界 | `NewECLBlockID`＋resume PC | global block identity 與資源載入由 title adapter 決定 |
| `PROGRAM` 邊界 | ID／ordered boundary signal | ID 的 caller context 不可跨作品假設，但「外部服務會暫停 VM」可重用 |
| 資源請求 | `LOAD FILES`／`LOAD PIECES` typed selectors | DAX／GEO／piece selection 是共用資料流；selector 對應哪張地圖由 pack 決定 |
| map edge／travel contract | `MapDefinition`、`MapPoint`、`ExternalExit`、`SearchEdge`、`MapPositionTransition` | lookup、native coordinate、confidence 與資料驗證可跨作品重用 |
| typed side-effect request | `DAMAGE`、`TREASURE`、`NPC`、`SAVE`、`CLOCK` 等 raw request | VM 只提交資料，角色／規則／存檔 consumer 留在 title adapter |

### 保留在 CoAB adapter／JSON 的部分

| 機制 | 不下沉的原因 |
|---|---|
| `0x2E10` 的 redraw／position semantics | DOS overlay dispatch、`C04B..C04F` projection、dirty flags 與 map consumer 尚未在同一位址空間閉合；不同作品可能用不同 map service |
| `0xC01E` 的 address-to-movement 對應 | 「這個 raw address 代表前進」是 CoAB 證據；engine 可提供 forced-move transaction，但不應內建 `0xC01E` |
| `0xB200` 的 sound A/B | `word_1EE76` consumer 尚未閉合；sound selector、平台 driver 與缺檔 fallback 應由 title audio adapter 宣告 |
| `PROGRAM 0/3/8/9` 的畫面／流程 | `PROGRAM 0` 在訓練館不是主選單；`3`、`8`、`9` 的 roster／save／camp 副作用也與 CoAB State context 有關 |
| `0x7F6C`／`0x7EE2` shop／temple dispatch | 這些是 CoAB ECL work memory；共用 engine 不應直接讀固定作品位址來猜 service type |
| world native values、城市名稱、旗標、ECL block、事件文字 | 屬 CoAB JSON／ECL data；不進作品中立 engine，不在 Go switch 硬編劇情 |

目前的實作策略是「engine 保存 raw／typed boundary，CoAB adapter 解讀已證實
語意」。若未來第二款 SSI Gold Box 能以原始 bytes 與正常 runtime trace 證明相同
producer→consumer，才新增明確、無作品地址假設的 engine contract；在此之前不為
了抽象而抽象。

## 世界地圖與旅行執行

### 原生資料契約

CoAB pack 的 `moonsea.overland` 宣告 14 個 native location value：`0..13`。
每個 `MapPoint.Destinations` 是 ECL1 route dispatcher 的 directed adjacency row，
不是把世界地圖當作可任意走的 25×50 wilderness cursor。`syncWorldDestinationSelectors`
只在目前 ECL menu 顯示完整 adjacency row 時，把 JSON native values 投影回
`0x4C02..0x4C05`；有條件隱藏的原版選單仍以 ECL 自己的 selector 為準。

荒野抵達的 transaction 現在是：

```text
選取 destination
  → 4C9C = native destination
  → 進入 ECL1 arrival entry 1
  → 4C9B = native destination（entry 前的 dispatch state）
  → ECL 返回後再次提交 4C9B（避免 route menu 沿用舊地點）
  → Area／Location／world menu 投影
```

`4C9B` 與 `4C9C` 的數值仍是 CoAB ECL work address；engine 只看到 title adapter
提交的 native destination，不把它們命名成跨作品固定欄位。

### 正常驗證

Docker image：`coab-go-test:20260729`；測試從原始 archive 解出 ECL1–ECL6，沒有
直接注入畫面文字或翻譯。`TestRealOverlandArrivalAndRouteGraphCoverage` 驗證：

- 14 個原生點位均能由 ECL1 真實 arrival entry 執行並回到 world state；
- 所有 adjacency destination 都指向已宣告點位；
- 從 Tilverton 的 directed graph 可達全部 14 點；
- 每個到達 state 的 `Area.CurrentCity`、`Area.GameArea=1`、`InDungeon=false`
  與 locale 投影均成立。

既有正常玩家測試另外涵蓋 Tilverton、Ashabenford、Standing Stone、Essembra、Hap、
Hillsfar、Yulash、Zhentil Keep、Myth Drannor 的特定事件與跨地城 continuation。
本輪尚未用一條新隊伍 session 逐一完成 14 點的全部城市設施、隨機遭遇、區域地圖、
所有房間與重訪旗標，因此「全地圖事件」仍是 P1 待完成，不在本規格打勾。

## 未完成與下一步

1. 正常 session 的所有世界路線仍需逐一覆蓋 arrival event、TRAIL／WILDERNESS／
   EXIT 分支、隨機遭遇、存檔重訪與條件目的地；本輪只關閉原生點位與路網基線。
2. `CALL 0x2E10` 的 DOS map projection、`0xB200` sound B、完整 external routine
   producer／consumer 仍依 `docs/knowledge/golden-box-reverse-engineering-worklist.md`
   的 P0/P1 順序處理；沒有必要時不再逐行解讀 overlay。
3. 16 個 GEO block 的 identity 已有 pack 宣告，但各 block 的所有 terrain event、
   entrance／exit、持久狀態與原版畫面 fidelity 仍要用正常輸入和資產證據驗證。
4. 本輪沒有修改獨立 engine repo；若後續要新增跨作品 API，先以第二款 SSI 遊戲
   adapter 或同等第一級證據驗證，避免把 CoAB address mapping 變成 engine 事實。

## 驗證命令

```text
PATH=/usr/local/go/bin:$PATH go test -count=1 \\
  -modfile=workplace/coab-test.mod \\
  ./internal/game -run '^TestRealOverlandArrivalAndRouteGraphCoverage$' \\
  -timeout=240s
```

結果：Docker 內通過。這是全世界點位／路網與 ECL arrival boundary gate，不是完整
整作通關 gate。
