# 第五百零四輪：PC-98 Quick 目標優先級重試邊界

狀態：`READY`（限 `04CCh..0624h` 的候選順序／優先級重試控制流與 CoAB
bounded adapter；不宣稱 random helper 身分、完整 pointer-chain tie 或完整
Quick AI）

## 本輪結論

第 503 輪只把 PC-98 Quick 目標候選的 `LegacyObjectID` 順序移入 engine。本輪
沿用同一份非破壞性 IDA 連續報告，補上 `04CCh` 已能由 raw bytes 支持的搜尋
形狀：

1. `04DEh` 將最低優先級起點設為 `7`。
2. `04E2h..04E8h` 以 `1` 與 `7` 呼叫未命名的 resident helper，將返回值保存
   到區域 byte；`052Fh..0536h` 拒絕小於 `1` 的值。
3. `0539h` 將 pass counter 設為 `1`，從施法者 action 的 `+14Eh` 取得候選鏈。
4. `0551h..05F1h` 沿候選 `+52h` next pointer。只有在尚未有 best pointer 時，
   才依候選 predicate 呼叫 local `03D3h`；`03D3h` 讀法術 record priority，
   並與傳入的最低優先級比較。
5. 候選鏈掃完後，`05F4h` 降低最低優先級；`05F7h..05FFh` 以 pass counter
   與前述 helper 結果比較，未達抽樣上限才重新掃描。
6. `0602h..061Eh` 只有 best pointer 非空才呼叫 `00FA:0048h` handoff。

這證明它不是「穩定排序後永遠取第一個」；但 `3E01:142Dh` 的正式亂數演算法、
候選欄位 `+65h／+56h／+66h／+5Ch` 的完整名稱、pointer chain 建立順序與
同分 tie 仍未閉合。

## 輸入與位址空間

| 輸入／產物 | SHA-256 | 用途 | 推論等級 |
|---|---|---|---|
| PC-98 `GAME.EXE` | `8bca0b50f47b5a41193584d3d4d1cd7361562ca3daf5360d3691620cc1b752c0` | resident symbols／資料 | `exact` |
| PC-98 `GAME.OVR` | `de87ea5311e1c2c45000c76472db6e280eadead7d1e8a00d14b243a7ec2bb38a` | overlay／TPOV | `exact` |
| overlay 09 | `c014bcbf9faf3acc4877386529d3b0aa74beac81f05d48e87d7f01de61031c20` | Quick selector／target helper | `exact` |
| `/tmp/pc98-quick-magic-overlay09.txt` | `3f7f6160e9aac8140ecf320ae45f34572b6d77a39bc7b5f1eb6c8f417c30b36c` | IDA Pro 9.4 連續指令報告 | `exact bytes／control flow` |

所有數值均為 overlay-local offset；不可與 resident effective address、file
offset、ECL work address 或角色 record offset 混用。原始 executable、overlay、
Borland symbol 與基準 database 唯讀保存；研究只在 Docker 的副本／暫存資料庫
進行，沒有 rename、patch 或覆寫原始位址。

## Remake 契約

- Golden Box engine `combat/quicktarget.Rule` 由 game pack 宣告
  `candidate_order=legacy_object_id`、`retry_roll_sides=7` 與
  `minimum_priority_start=7`。
- `Select` 使用同一個 Battle PRNG stream 擲一次重試次數，依 threshold 掃
  legal candidates；engine 不認識 CoAB 法術、中文、地形或角色欄位。
- CoAB suitability 階段改成只檢查「是否至少有一個合法候選」，不消耗 target
  retry roll。法術選定後，area／line／Curse／Cause／Protection 的 final target
  handoff 才執行一次 `SelectQuickTarget`；Magic Missile 與 Cure 專用流程不在
  本輪改寫。
- 若 selector 已選法術、但有限 priority pass 沒找到目標，法術格不消耗，State
  回到一般戰鬥 action；這是可恢復的 no-target 邊界，不是資料錯誤，也不代表
  原版完整 AI 已還原。

## 驗證

- engine `combat/quicktarget` 測試：驗證 legacy order、priority pass、抽樣面數、
  輸入不被修改與非法 roll fail-closed。
- CoAB `internal/combat`／`internal/game` 測試：驗證 Quick target 經 Battle
  PRNG、area／line／targeted pending handoff，以及無目標時不消耗 slot。
- Docker focused gate：`internal/combat` 與 `internal/game` 通過；正式
  `./cmd/... ./gamepack ./internal/...` 與 `coab-audit` 仍需在本輪收尾重跑。

## 尚未完成

1. `3E01:142Dh` helper 的演算法與 DOS／PC-98 RNG 對應。
2. `+52h` chain producer、candidate field consumer、同分 tie 與完整 target
   pointer→grid projection。
3. Magic Missile、Cure 特殊選擇與敵方 Quick AI 的完整 target 行為。
4. 弓箭／法術動態演出、全部 ECL 玩家路徑、全規則、全翻譯、音訊與發行包；
   本輪不能支撐「整作 remake 已完成」聲明。
