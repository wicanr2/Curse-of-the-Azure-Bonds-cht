# 第一百八十六輪：SAVGAM player writeback

狀態：`READY`

## 目的

把第 185 輪已完成的 `LoadSAVGAMSlot` 延伸成可測試的原版 slot 回寫入口，讓 remake 對已載入角色的 HP、能力、財寶、icon、法術與 thief skill 變更可以寫回 DOS player bundle。

## 證據與邊界

- `ovr017.SaveGame/loadSaveGame` 證實 slot prefix、`CHRDAT{slot}{1..6}.sav`、`.swg` 的 `0x3F` item records 與 `.fx` 的 9-byte effects records。
- `PatchDOSPlayerRecord` 以既有 parser 的 offset 為準，只覆寫已證實欄位；原始 `.sav` 其餘 bytes 維持原值。
- `.swg`／`.fx` 由目前的 `EncodeItems`／`EncodeAffects` 產生；沒有把未知 sidecar bytes 假裝成已知格式。
- `SaveSAVGAMSlot` 先把所有輸出寫入 slot 目錄下的 staging directory，再逐檔 replace；這是失敗時不留下半成品 staging 的 staged replacement，不等同原版刪檔或跨檔案 atomic commit。

## 驗證

`internal/game` regression 會載入 synthetic slot、修改 HP／金幣／memorized spells、保存後重新 parse，並檢查 `.sav` 未知 byte 未被改寫；`internal/monster` regression 覆蓋 `.swg`／`.fx` known-field round-trip。

## 後續

仍需反組譯確認原版刪除舊 player files、rename failure recovery、多職業欄位、未知 sidecar records，以及 CAMP SAVE 是否應直接選用此 slot writer。
