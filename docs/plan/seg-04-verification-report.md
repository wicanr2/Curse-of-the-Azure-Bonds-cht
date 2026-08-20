# 驗證報告：`SEG-04` 統一直入旗標

- 日期：2026-08-20
- 範圍：主線分段驗證計畫的 `SEG-04`（`-segment <id>` 取代散在各處的專用旗標）
- 結論：**通過**。25 段全部進得去，逐段一條測試；過程中修掉兩個讓段入口
  畫面落空的素材鍵錯誤，剩一個真的缺口。

## 一、註冊表與旗標

段的入口資料集中在 [`internal/segment`](../../internal/segment/segment.go)：
每一段記 block 編號、直接進入時要把 `LastECL`（`4BF2h`）設成誰、章節編號、
在地城還是世界地圖上，以及對應的既有旗標名。

```
tools/go.sh run ./cmd/azure-bonds-game -segment list
tools/go.sh run ./cmd/azure-bonds-game -segment ECL5/0x33 ...
tools/go.sh run ./cmd/azure-bonds-game -segment 0x33 ...
tools/go.sh run ./cmd/azure-bonds-game -segment wizard-tower ...
```

三種寫法收同一段：完整 id、只給 block 編號、既有旗標名。
⚠ 旗標名對到的是它**實際進的段**，不是名字看起來的那一段：`-opening` 對到
`ECL2/0x01`，因為 remake 的新遊戲流程 `BeginAdventure` 直接 reset 到 `0x01`
（見 §五）。

⚠ 既有旗標**不只**做段的入口——`-lava-tube` 之類還會把隊伍走到段內某一格、
或先打完一場戰鬥。那些是段內的檢查點，所以旗標本身保留，`-segment` 只保證
進得到**段的入口**。

進入的動作是 `State.StartStorySegment`：切 block → 設 `LastECL` → 跑那一段的
initial lifecycle（`vm_init_ecl` 的第 5 個入口）。既有的
`StartDungeonStoryPreview` 變成它的地城特例，行為不變。

### `EnterFrom` 從哪來

取自 [`docs/audit/ecl-block-graph.md`](../audit/ecl-block-graph.md) 的「進入自」
欄位，挑順著劇情的那一條。**`0x00` 不是某個 block，是「全新開局」**：
spec 1141 的主迴圈只在 `LastECL ＝ 0` 時走到開場（`0x52`，需要 `DS:4FBCh` 非 0）
與提爾佛頓第一段（`0x01`）。這兩段在轉移圖上本來就沒有進入邊。

## 二、四條測試

| 測試 | 擋住什麼 |
|---|---|
| `TestRegistryMatchesSegmentLabels` | 註冊表與 `segment-labels.json` 分岔（有段沒人認領）|
| `TestEnterFromIsARealIncomingEdge` | `EnterFrom` 憑印象填——它必須是轉移圖上真的存在的一條邊 |
| `TestGameAreaFollowsTheECLMember` | 章節編號與 ECL 成員編號不一致 |
| `TestEverySegmentInTheRegistryCanBeEntered` | `SEG-04` 的驗收條件本身：25 段逐段進一次 |
| `TestSegmentEntryArtIsExported` | 段的第一個畫面是「素材尚未載入」|

段測試是 `t.Run(seg.ID, …)`，紅的時候直接看得出是哪一段——這正是分段的目的。

## 三、`ECL5/0x30` 是過場 block

進入 `0x30` 之後停在 `0x50`，不是停在自己身上。它只有 72 條可達指令、唯一的
出邊是 `0x50`，initial lifecycle 一跑就把隊伍送回世界地圖。註冊表用
`SettlesAt` 記這件事，其餘 24 段都停在自己身上。

## 四、兩個素材鍵錯誤

直接進入把「這一段的第一個畫面」單獨拉出來看，於是三段的入口是空的黃字。

**大圖的檔案選擇不該跟著章節走。** 原作只有 `BIGPIC1/2/6` 三個檔，一共四個
區塊，且彼此不重複：

| 檔案 | 區塊 |
|---|---|
| `BIGPIC1.DAX` | `0x79`、`0x7B` |
| `BIGPIC2.DAX` | `0x78` |
| `BIGPIC6.DAX` | `0x7A` |

第 3～5 章根本沒有 BIGPIC 檔，而世界地圖那張 `0x79` 是全遊戲共用的。
原本的鍵是 `bigpic{章節}-block-{區塊}`，於是 `ECL5/0x30` 去要不存在的
`bigpic5-block-79`。改成用**區塊編號定檔**（`cmd/azure-bonds-game` 的
`bigPictureSprite`）——四個區塊互不重複，這個對應是唯一的。

**世界地圖上的段也要有章節編號。** `StartStorySegment` 原本只在地城分支設
`Area.GameArea`，世界地圖段沿用上一段留下的值。開場（`ECL1/0x52`）於是拿著
章節 2 去要 `pic2-block-1D`，而實際存在的是已經匯出的 `pic1-block-1D`。
現在章節一律等於 ECL 成員編號；GEO 幾何仍然只有地城段才載。

## 五、原作序幕第一次跑出來

正常新遊戲流程走的是 `0x50 → 0x01`（`BeginAdventure` 直接 reset 到 `0x01`），
**沒有經過開場 block `0x52`**。`-segment ECL1/0x52` 是目前唯一進得去的路徑，
畫面上是完整的序幕：營火伏擊圖、三名 NPC（RUSTLE／CYNTHIA／GHENDEL）入隊、
五枚青色符印的緣起敘述。

![原作開場序幕：營火伏擊、三名 NPC 入隊與五枚青色符印的緣起](../screenshots/opening-prologue-remake.png)

「正常流程要不要經過 `0x52`」是行為問題，本輪不改；`SEG-20` 接線時再決定。

## 六、剩下的缺口

`ECL4/0x22`（散提爾堡：迪姆斯華特與脫身）的入口要
`character-area-4-head-25-body-23.png`——**頭與身體是不同區塊的組合**，
而匯出腳本目前只出 head 與 body 同號的那些。列在
`knownSegmentArtGaps` 裡，新的缺口會讓測試紅。

## 七、明確不宣稱

- 沒有宣稱 `EnterFrom` 是原作唯一的進入路徑；它是轉移圖上其中一條真實的邊，
  選的是順著劇情那一條。
- 沒有宣稱直接進入之後的隊伍狀態與正常打到那裡一樣——那是 `SEG-11` 的快照
  要解的事。本輪的隊伍是臨時建的一名角色。
- 沒有宣稱 `-segment` 與既有旗標等價（旗標多做段內定位，見 §一）。
- 沒有宣稱原作在什麼條件下會走 `DS:4FBCh` 非 0 那一支（＝ 進開場 block）。
