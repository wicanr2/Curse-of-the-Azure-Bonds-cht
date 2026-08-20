# 第五百四十八輪：眼魔洞穴 A2 續跑與資料驅動地圖 handoff

狀態：`READY`（限 ECL4 block `0x22` 的 A2→死精靈選單續跑；不是完整洞穴或整作通關）
日期：2026-08-12

## 勘誤與結論

第 547 輪正確閉合了 `C04B..C04F` 的 DOS map-register bridge，但把 A2 地形
觸發後的座標寫入誤說成同一時刻完成。新的原始 ECL trace 證明正確順序是：

1. ECL4 block `0x22` 的 `+050A` 以 `C04F=0xA2` 進入砲擊文字序列；進入時
   `4C03=1`。
2. 玩家先閱讀「彷彿被砲彈射出」的頁面，並經過三次 `PRESS` boundary。
3. 第三次續跑才抵達 `+061B`，依序寫入 `C04B=13`、`C04C=1`、`C04D=3`、
   `4C06=1`，再呼叫原始 `CALL 0x2E10`。
4. 同一份 ECL session 隨即進入死精靈的 `EXAMINE REMAINS／LEAVE` 選單；不是
   可在地圖 handoff 時被吃掉的第二個事件。

因此 `4C03=0` 不是此 transaction 的前置條件；初始 `4C03=1` 在 `LEAVE` 後仍
保持。舊規格將它視為 handler 會清除的值，是已被 raw trace 推翻的斷言。

## Engine／game-pack 邊界

獨立 Golden Box engine 新增中立 `continue_result` 契約：

- 僅 `set_map_position` action 可設定；schema 與 runtime 會拒絕其他 action 使用。
- action 會先投影資料包宣告的位置、朝向及 map cache，再把同一份 ECL result
  留給前端／State 續跑。
- 它不包含洞穴名、座標、密語、劇情旗標或任何 CoAB 遊戲規則。

CoAB 的 `zhentil-keep.beholder-cave.same-block-launch` 只宣告 raw trace 已驗證的
`C04B/C04C/C04D`、`4C06` 條件與目的地 `(13,1,W)`／`wall=08`／`roof=C0`。
故事文字、死精靈選項與翻譯仍由 ECL／CoAB JSON 的 stable ID 解析，沒有回寫到
Go frontend。

## 證據與驗證

| 項目 | 證據 | 等級 |
|---|---|---|
| A2 分派與三次 `PRESS` | 原始 `ECL4.DAX` block `0x22` 的 bounded session trace | `exact` |
| `+061B` 的 map 寫入與 `CALL 0x2E10` | raw opcode trace 與 `SaveWrites` | `exact` |
| `4C03=1` 在 `LEAVE` 後保留 | 同一 raw session 的 state assertion | `exact`（本 trace） |
| C04 map-register 寫入投影 | 第 547 輪 DOS overlay-07 IDA bridge | `exact`（writer/getter bridge） |
| `(13,1,W)` map cache | 原始 GEO4 block `0x25` | `exact` |
| 玩家由 E1 到 A2、三次續跑並在同一 result 抵達死精靈提示 | `TestRealNewGameRunsToTheEnding` | `exact`（remake normal path） |

Docker focused gate：

```text
go test -v ./internal/ecl \
  -run '^TestRealECL4CaveA2CannonContinuesToDeadElfHandler$' -count=1
go test -v ./internal/game \
  -run '^TestRealNewGameRunsToTheEnding$' -count=1
go test ./gamepack \
  -run '^TestBeholderCaveMapHandoffContinuesSameECLResult$' -count=1
```

## 範圍限制

- 這只證明 A2 事件的文字、位置 handoff 與死精靈選單 continuation；不等於
  Dexam、所有洞穴房間、出口、重訪或完整主線已完成。
- `0x4C00` 仍是與這條玩家結果無關的 `unknown`；不因本輪被命名或列為阻塞。
- `(15,1)` 的資料驅動搜尋邊仍只是 `strong inference`；它不是這條正常玩家
  路徑的證據，也不能被當作 Dexam 或出口的 direct-entry 驗收捷徑。

## 本規格的後續

死精靈之後的正常 session 由 [`spec 550`](550-ecl4-dead-elf-journal59-treasure-continuation.md)
接手：皮袋、氣體陷阱、手札 59 與戰利品服務結束後回到同一洞穴座標。
本規格其餘 A2 map-register 與 continuation 證據仍然有效。

第 547 輪的 raw map-register 證據仍有效，但洞穴 route 時序與完成敘述由本規格
取代。
