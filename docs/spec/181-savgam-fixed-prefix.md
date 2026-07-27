# 第 181 輪：SAVGAM 固定前綴 codec

狀態：`READY`（限固定前綴）

## 證據

`engine/ovr017.cs` 的 `SaveGame`／`loadSaveGame` 以固定順序處理 `SAVGAM?.DAT`：

1. `game_area` 1 byte
2. Area1 `0x800`、Area2 `0x800`、`stru_1B2CA` `0x400`、ECL memory `0x1E00`
3. map segment 5 bytes：signed `mapPosX`、signed `mapPosY`、direction、wall type、roof
4. last/current game state 各 1 byte
5. 三組 little-endian `blockId`／`setId`，共 12 bytes
6. party count 1 byte 與 8 筆固定 `0x29` CHRDAT name records，共 `0x148` bytes

`internal/save` 現在以 `SAVGAMContainer` 保留這個已證實的 binary boundary，並可 round-trip 未解碼的 Area、runtime、ECL raw bytes。`DecodeSAVGAM` 允許固定前綴後存在資料，因為 reference 之後另寫個別 CHRDAT player files；本輪不猜測那些 side-effect file 的名稱、slot 配置或內容。

## 驗證與沿用

- codec 尺寸為 `SAVGAMFixedPrefixSize`，Encode 對每個 raw segment 做嚴格尺寸驗證。
- map 座標保留 signed byte，set block 使用 little-endian，未使用的 CHRDAT record 仍保持 zero-filled。
- 後續 Gold Box 遊戲可沿用 container prefix codec，再替換作品專屬 Area／ECL schema 與 player-file adapter。

本輪沒有宣稱完成完整 DOS save/import slot、Area 欄位解碼、CHRDAT 實體檔案 side effects、atomic write 或遊戲內 CAMP SAVE 接線。
