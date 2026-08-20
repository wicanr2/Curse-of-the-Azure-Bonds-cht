# 驗證報告：`SEG-11` 段落快照往返

- 日期：2026-08-20
- 範圍：主線分段驗證計畫的 `SEG-11`（可重播的段落快照）
- 結論：**前半通過**。25 段的邊界狀態都存得下去、讀得回來，交接也實際跑通了；
  「每段**結束**時的快照」要等每段有自己的測試（`SEG-10`）之後才存得到。

## 一、往返閘

`TestSegmentEntrySnapshotRoundTrips` 對 25 段各做一次：進段 → `SavePartyFile`
→ 用一份新的 `State` `LoadPartyFile` → 比對。比對的欄位是段的邊界狀態：

| 欄位 | 為什麼算邊界 |
|---|---|
| `session.CurrentBlockID()` | 下一段要從哪個 block 接下去 |
| `Mode` | 玩家在地城、世界地圖還是事件裡 |
| `Area.GameArea`／`Area.InDungeon` | 章節與所在層級 |
| `GeoMapSet`／`GeoMapBlock` | 這一段站在哪張地圖上 |
| `DungeonX`／`DungeonY`／`DungeonDirection` | 站在哪一格、面向哪 |
| 隊伍人數 | 隊伍有沒有跟過來 |

## 二、GEO 檔集只能跟著章節走

三個世界地圖段一開始往返不回來：存的是章節 1、讀回來 GEO 檔集變成 1，
而原本的執行期值是 2。

原因不在存檔，在寫入端。**存檔沒有獨立的 GEO 檔集欄位**——`LoadPartyFile`
是拿 `Area.GameArea` 重建它的。所以 `GeoMapSet` 與 `Area.GameArea` 留下不同的
值，快照一定往返不回來。`StartStorySegment` 原本只在地城分支設 `GeoMapSet`，
世界地圖段沿用上一段留下的 2；現在兩個分支都設。

世界地圖段不載幾何，所以這個值不影響畫面；`cmd/azure-bonds-game` 查不到對應的
GEO 區塊時會留著旗標給的那一張。

## 三、交接實際跑過

`-segment-snapshot <path>` 進段之後把存檔寫出來就結束，下一段用 `-party-load`
接上：

```
tools/go.sh run ./cmd/azure-bonds-game -segment ECL5/0x33 \
    -segment-snapshot workplace/ecl5-33.json
tools/go.sh run ./cmd/azure-bonds-game -party-load workplace/ecl5-33.json \
    -segment ECL5/0x30 -screenshot workplace/handoff.png
```

巫師塔的快照接進黑暗精靈章尾聲，過場之後停在世界地圖 hub，畫面是
「你們來到提爾佛頓城外，要進城，還是繼續旅程？」——與 `ECL5/0x30` 唯一的出邊
`0x50` 一致。

## 四、明確不宣稱

- 沒有宣稱這些快照等於「正常打到那裡」的狀態：本輪存的是**段的入口**，
  而且隊伍是臨時建的一名角色。段的結束狀態要等 `SEG-10`。
- 沒有宣稱往返保留了選單、待處理的 ECL 執行位置或戰鬥交易；比對的欄位列在
  §一，沒列的沒有量過。
- 沒有宣稱 `-party-load` ＋ `-segment` 等於原作走過那條 `NEWECL` 邊——那是
  `SEG-12` 的 47 條交接測試要驗的事。
