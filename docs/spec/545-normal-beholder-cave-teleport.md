# 第五百四十五輪：眼魔洞穴正常傳送與位置交易

狀態：`READY`（限本輪已驗證的正常玩家路徑；不是完整洞穴或整作通關）

## 結論

第 543／544 輪的同一新遊戲 session 現在能從散提爾堡正常進入眼魔洞穴，
再由洞穴出生點 `(4,5,N)` 逐格走到 GEO4 block `0x25` 的 `(5,9)`、terrain
`0xA2`。原始 ECL4 block `0x22` 的同區塊事件會把位置交易寫成另一個洞穴座標；
CoAB game-pack 現在以 `set_map_position` 宣告這個已驗證的玩家可見結果，正常
路徑最後落在 `(13,1,W)`。

這項修正不是 `0x4C00` 規則解碼，也沒有把 `0x4C00` 命名成 D&D 欄位。只要
某個 raw work address 不改變玩家可見位置、D&D 數值、戰鬥、存檔相容或其他規則
結果，就維持 `unknown`，不列為 remake blocker。

## 第五百四十六輪勘誤：Dexam 聚焦夾具不需要 `0x4C00`

原本的 `TestRealBeholderCaveDexamAndZhentilBattles` 曾為了模擬較早的散提爾堡
狀態，直接寫入 `0x4C00=1`。本輪移除這個前置值後，在 Docker 中重新驗證同一
個聚焦玩家結果：Dexam 揭露、兩場戰鬥、戰利品、戰後 ECL continuation、洞穴
出口與回到世界地圖仍通過；測試只保留該夾具真正需要的 `0x4C01` escort state。

這項結果只證明 `0x4C00` 不是 Dexam 這段 D&D／玩家流程的必要輸入，不是把它
全作所有事件的 raw writer 刪除。其他已證明會影響劇情靜默、儀式完成或路由的
事件，仍可保留未命名的 raw write；只要沒有玩家結果或 D&D 規則證據，就不再為
`0x4C00` 追查完整 consumer，也不把它命名為年齡、能力值或其他規則欄位。

## 證據與位址空間

| 項目 | 結果 | 推論等級 |
|---|---|---|
| 原始輸入 | `curseoftheazurebonds.zip`，SHA-256 `c98698a6271c17177dfdb27f34b0389b7d34f58ef206e92575393f4655f5b26d` | `exact`（輸入識別） |
| ECL 資源 | `ECL4.DAX`，script block `0x22`；這是 DAX block／ECL 位址空間，不是檔案 offset | `exact`（資源選擇） |
| GEO 資源 | `GEO4.DAX`，geometry block `0x25`；出生 `(4,5,N)` 與觸發候選 `(5,9)` terrain `0xA2` | `exact`（解碼資料／重製地圖） |
| ECL work position | VM 事件後觀察到 `C04B=13`、`C04C=1`、`C04D=3`；`C04D` 的 raw half-direction 由 CoAB adapter 投影成畫面方向 `6`（西） | `strong inference`（位置交易）；不可外推為 D&D 欄位 |
| player-visible result | 同一正常 session 的 `State` 最終為 area `4`、GEO `0x25`、`(13,1,W)`，且 JSON event exactly-once 已套用 | `exact`（remake contract） |

本輪使用 `coab-go-test:20260729` Docker image、CoAB CLI 的 GEO decoder／path
diagnostic 與 Go normal-path test。ECL decoded payload offset、ECL code address、
ECL work address、GEO cell 與檔案 offset 維持不同位址空間；本文件不以相同的
十六進位數字合併它們。

## 實作邊界

新增的 JSON event：

`zhentil-keep.beholder-cave.same-block-launch`

以 `ecl_block=34` 與 `C04B/C04C/C04D` 三欄 predicate 觸發，呼叫既有 engine
的中立 `set_map_position`；Go State 沒有新增洞穴名稱、劇情旗標或 D&D 規則
分支，也沒有讓所有地城生命週期盲目同步 `C04B..C04D`。因此下水道等其他
地圖不會因共享 raw 暫存器而被錯誤移動。

驗收測試為
`TestRealNewGameContinuesFromHapToBeholderCaveEntrance`，沒有直接設定洞穴
座標、沒有 direct-entry 戰鬥，也沒有寫入 `0x4C00`。測試固定斷言：

- mode 仍為 `ModeDungeon`；
- area `4`／GEO block `0x25`；
- 位置為 `(13,1,6)`，即 `(13,1,W)`；
- 上述 game-pack event 已套用一次。

## 尚未完成

從 `(13,1)` 到 Dexam 入口 `(15,1)` 在目前解碼的 GEO4 block `0x25` 不是一般
可走路徑；本輪不以攻略座標或 generic edge 假造連線。Dexam 雙戰、洞穴其他
傳送／隨機遭遇、出口、重訪持久化與完整戰後世界路由仍是後續玩家可見工作，
必須各自有正常輸入與 ECL continuation 證據。
