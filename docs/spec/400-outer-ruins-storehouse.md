# 400：外圍遺跡倉庫入口、主動搜索與財寶

狀態：`READY`

## 問題

第 399 輪只完成提爾雪雅結盟戰，尚未證明戰後倉庫如何進入、入口守軍是否
重播、普通踏入與玩家主動 `SEARCH` 有何差別，以及倉庫財寶的 exact 內容。
此外，Burial Glen 邊界流程留下的 `7ED5h` 若滲入新 ECL block，會讓外圍
遺跡每一步重播來源地圖的邊界問題，阻斷目的格事件。

## 原始資料證據

- 來源：`curseoftheazurebonds.zip` 中 `GEO6.DAX／ECL6.DAX／ITEM6.DAX`。
- `GEO6.DAX` block `42h`：
  - 提爾雪雅所在 terrain `01h` 是 `(1,12)`。
  - 倉庫入口 terrain `02h` 是 `(3,14)`。
  - 倉庫內部 terrain `83h` 是 `(2,14)`。
  - `(1,12)→(2,12,E)→(3,12,E)→(3,13,S)→(3,14,S)→(2,14,W)`
    每一步均通過 `CanMoveDungeonWrapped`；測試沒有瞬移或忽略牆面。
- terrain `02h` 在 `4CD1h==0` 時詢問是否逃走：
  - `YES` 直接離開且不寫完成旗標。
  - `NO` 建立 `MON6CHA 44h` HELL HOUND ×6 與 `45h` MARGOYLE ×6。
  - 勝利後 `SAVE 1,[4CD1h]` 位於 payload `+0C71h`。
  - `4CD1h==1` 時入口安靜；因此第 399 輪與提爾雪雅並肩作戰以及直接
    擊敗倉庫入口守軍，會匯合到同一入口完成狀態。
- terrain `83h` 的普通 SearchLocation 只顯示食物、衣物與小物件堆積；
  不寫 `4CD2h`，也不發出財寶。
- 玩家主動 `SEARCH` 會以 work byte `7ECAh=1` 進入另一分支：
  - 顯示最終找到值錢物品。
  - exact `TREASURE` operands 是
    `[0,0,0,2000,1500,8,8], ItemBlock=82h`。
  - 依既有貨幣換算為 9,500 gold、8 gems、8 jewelry；真實
    `ITEM6.DAX/82h` 解析為兩件裝備。
  - 財寶服務返回後，payload `+0D00h` 才執行
    `SAVE 1,[4CD2h]`；第二次主動搜索直接離開，不再複製財寶。
- `7ED5h` 是來源地圖的 boundary-attempt work signal。正常
  `RunDungeonExitLifecycle` 設定它後，若 ECL 選單分支已 `NEWECL` 或直接
  `EXIT`，該次邊界 transaction 即已消費完成；目的 block 的下一次 per-turn
  entry 不得再看到它。

以上字串、怪物、旗標與財寶均由 real-image `BlockSession` 重現。地址掃描
另由 `-find-save-destination 4CD1／4CD2` 交叉驗證 writer；結論標為
`exact`。貨幣顯示與 ITEM block 投影由現有 typed adapter 驗證。

## 實作

1. game-pack 新增入口守衛、物資堆與搜索發現三組英文／繁中 stable ID；
   Go runtime 與測試不複製中文正文。
2. `State` 的地城財寶路徑在建立通用財寶選單後，保留本次 ECL 的資料包
   翻譯文字，不再被「發現財寶」通用提示覆蓋。
3. `State` 額外追蹤正在進行的 boundary transaction。只有該 transaction
   的選單分支完成 `NEWECL` 或 `EXIT` 時才清除 `7ED5h`；不能在每次普通
   移動時無條件清除，因為那會改變同一 block 內其他已驗證流程。
4. `TestRealOuterRuinsStorehouse` 鎖定逃跑、直接戰鬥、重訪、普通踏入、
   主動搜索、exact 財寶及一次性旗標。
5. Standing Stone 起始的長玩家路徑接續第 399 輪，沿合法 GEO 路線進入
   倉庫、閱讀繁中物資描述、執行真正搜索、驗證財寶增量，離開財寶選單後
   再次搜索並證明沒有重複取得。

## 完成邊界

本輪完成 terrain `02h／83h`，不代表 block `42h` 已完成。terrain
`04h／05h` 的逃亡男子、`08h／09h` 的灌木事件、`0Bh／0Ch`、rakshasa
居所、下水道入口、神殿與結局仍須依同一證據流程逐項完成。地獄犬與石像鬼
的完整 AD&D 特殊能力及 DOS 戰鬥動態演出也不在本輪完成範圍。
