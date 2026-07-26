# 第九十三輪：inventory quantity／cursed mutation

狀態：`READY`（限 quantity、readied、cursed 的 party mutation）

`Character.RemoveItem` 依 item record 的 quantity 規則移除一個 inventory unit：`Count == 0` 視為不可堆疊的單件物品，`Count > 1` 則遞減原 stack 並回傳一個 unit。readied item 必須先卸下，避免商店／寶物 mutation 與 fighter projection 不同步。`UnequipItem` 對已 readied 的 cursed item 回傳錯誤，模擬 cursed equipment lock。

這一輪不處理消耗品 effect、wand charges、spell scroll、彈藥扣除、金錢交易或 DOS save offset；`RemoveItem` 是後續商店／treasure service 可呼叫的安全底層 primitive。

驗證：`internal/party/character_test.go` 覆蓋 cursed lock、stack decrement、readied removal rejection 與 non-stacking removal；`go test ./...` 應通過。
