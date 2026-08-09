# SSI Gold Box 正常玩家路徑知識

本頁記錄如何把反組譯／原始資料解碼成果接成可重播的正常玩家路徑，供青色枷
與後續 SSI Gold Box game pack 共用。這不是攻略，也不是把測試 shortcut 包裝成
通關證明。

## 路徑的四層

每一段玩家路徑都要分開保存：

1. **來源層**：原始 executable、ECL／GEO／DAX bytes、符號、手冊與 runtime
   observation。記錄檔名、SHA-256、工具版本、位址空間。
2. **狀態層**：作品中立 engine 的 typed state、ECL session、座標／方向、
   pending continuation 與 save snapshot。
3. **資料層**：CoAB game-pack 的 block、map selector、事件 ID、stable message
   ID、人物／怪物／戰利品與翻譯；不得把劇情常數塞進共用 runtime。
4. **驗收層**：由新遊戲或正常地圖輸入抵達，保存每個 boundary 的 state、畫面、
   音效 intent、Journal／獎勵與離開後重訪結果。

## 地城移動交易

第 510 輪的 `State.MoveDungeon` 是一個可重用的範例。前端不應自行做一半：

```text
player input
  → cardinal delta／direction validation
  → decoded GEO wall／door legality
  → geometry coordinate wrap
  → ECL-local projection（若有 map adapter）
  → wall／roof register projection
  → transient per-step guard clear
  → ECL per-turn
  → ECL search-location／event boundary
  → renderer refresh after success
```

`GEO` 的 `(x,y)`、ECL 的 `C04B／C04C` 與檔案 offset 不可混稱。先以
`DungeonGeometryView` 取得 geometry 空間，再由 `SetDungeonGeometryView` 投影回
script 空間；日後其他 Gold Box 作品若有不同鏡像／局部地圖，只新增 typed adapter，
不要在劇情 State 內寫一個座標特例。

## 測試規則

- 最強的路徑測試從 `ActionStart`、角色建立或版本化 save 開始，使用正式 game pack
  與原始 bytes；不要先設定目標座標、旗標或 ECL PC。
- 若為了鎖定單一 opcode 而直接設定座標／flag，名稱必須包含
  `coordinate-assisted` 或 `direct-entry`，並且不得列入完整正常玩家路徑證據。
- 事件輸入、選項、文字、裝備與法術期待值要以 stable ID／JSON resolver 取得；
  測試不能複製目前的繁中顯示字串。
- 每段路徑都要驗證 continuation：戰鬥後、圖片後、手札後、離開後重訪，而不是
  只驗證第一個畫面。
- 通過一段路徑只提升該段的 evidence level，不會自動提升整作完成度。

## 提爾佛頓設施路徑與 one-shot 群組

第 511 輪把第 510 輪的一格交易延伸到提爾佛頓設施之間。從開場中斷休息後的
`(4,13)`，正式測試透過 `State.MoveDungeon` 逐步抵達 Filani、Weaponers、Gond
altar、Training Hall、Tavern 與高階祭司所在格；途中招牌與下水道傳聞的繁中內容
由 locale JSON 提供，未知 pause 不可被測試忽略。

高階祭司是正常路徑驗證中特別容易誤判的例子。原始 ECL2 block 1 的 SearchLocation
先由 terrain 得到 `7F7Ah`，高階祭司入口在 payload `+1104h` 以
`AND 4C03h,7F7Ah,7F79h` 後於非零時 `EXIT`。因此 terrain `0x0D` 招牌先把
`4C03h` 設為 `0x80` 後，之後抵達 terrain `0x8F` 可能按同一 one-shot 群組保持
靜默；這是原始控制流，不是 remake 應清掉的 stale flag。若要驗證祭司本身，應以
未先消耗該群組的 fresh ECL session 進行 branch regression，再另外驗證正常重訪
不重播。

這裡的「共用事件群組」是 `strong inference`：AND／EXIT bytes 與 runtime 前後
狀態已閉合，但 `4C03h` 在所有 SSI 作品的泛用名稱、存檔格式與所有城市事件仍未
完整解出。不要把 `0x80` 直接寫成跨作品固定劇情旗標。

## 目前 CoAB checkpoint

已由新遊戲進入提爾佛頓 GEO2 block 1，使用原始 west step 抵達 Windlord’s Inn，
並回歸圖片、HEAD／BODY 舞台、繁中訊息與 Journal 31。第 511 輪再從正常地城
狀態逐步走過提爾佛頓多個設施，記錄招牌／祭司共享 one-shot 群組，並以 fresh
session 回歸高階祭司完整分支。這段 map／ECL integration 是 `exact`（remake
對原始資料）；State movement transaction 與 DOS 逐幀 loop 的對應仍是
`strong inference`。其餘由 Tilverton 到結局的路徑仍必須逐段建立，不能把本頁的
checkpoint 寫成完整通關。
