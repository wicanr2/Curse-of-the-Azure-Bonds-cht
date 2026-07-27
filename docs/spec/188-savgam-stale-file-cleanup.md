# 第一百八十八輪：SAVGAM stale-file cleanup

狀態：`READY`

## 證據

`ovr017.SaveGame` 在寫入 party records 後會處理舊 player files；因此 party 從六人縮編時，slot 目錄不能保留會被下一次 loader 誤認的舊 `CHRDAT{slot}N` sidecars。

## 實作邊界

`SaveSAVGAMSlot` 只處理目前 slot key 的明確集合：prefix 與 `CHRDAT{key}{1..6}` 的 `.sav`、`.swg`、`.fx`。既有檔案先移到受限 backup directory，新的 prefix／player sidecars 完成後才替換；替換中發生錯誤時移除已安裝的新檔並還原 backup。其他檔名不會被掃描或刪除。

## 驗證

synthetic slot regression 先建立 stale `CHRDATC2.sav`，以一人隊伍保存後確認它被移除；既有 raw-byte regression 同時確認 `.sav` 未知欄位仍保留。

## 尚未完成

多職業 record、未解碼 player bytes、原版 rename／delete 的完整錯誤語意，以及跨作品的 player serializer 仍需更多反組譯證據。
