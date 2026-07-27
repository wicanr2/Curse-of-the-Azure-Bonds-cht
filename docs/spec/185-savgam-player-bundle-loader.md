# 第 185 輪：SAVGAM player bundle loader

狀態：`READY`（限 reference save load）

`ovr017.SaveGame/loadSaveGame` 證實：slot 主檔使用 `savgamA.dat` 到 `savgamJ.dat`；prefix 的 party count 後，每名角色使用 `CHRDAT{slot}{1..6}.sav`，並可伴隨同 basename 的 `.swg` inventory 與 `.fx` effects。`game.State.LoadSAVGAMSlot` 現在依此命名載入：

1. 讀取並套用第 182 輪 SAVGAM prefix adapter。
2. 依 party count 讀取必要 `.sav`，optional sidecars 缺少時保留 nil。
3. 重用既有 DOS player／SWG／FX parser，建立 party roster、fighter projection 與繁中啟動狀態。
4. `cmd/azure-bonds-game -savgam-dir DIR -savgam-slot A` 可直接啟動該 slot。

本輪只實作有證據支持的 load path；不會由 name record 猜測玩家資料，也沒有把 remake 角色改寫成原版 `Player.StructSize`、刪除原始檔、重建 `.swg/.fx` 或完成 CAMP 多檔案 save transaction。
