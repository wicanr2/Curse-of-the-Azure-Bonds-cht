# 第 182 輪：SAVGAM prefix 與 State adapter

狀態：`READY`（限已解碼 state 欄位）

`State.LoadSAVGAMPrefix` 現在可讀取第 181 輪的固定前綴，並套用已經有 Area codec 證據的欄位：

- Area1／Area2 的 `GameArea`、GEO block、dungeon flag、city、sky、head block 等已知欄位。
- 5-byte map segment 的 signed `mapPosX/mapPosY`、direction、wall type、roof。
- 原始 Area1／Area2、runtime state、ECL memory、set blocks 與 CHRDAT name records 會保留在 State 的 raw prefix 中。
- Area1／Area2／runtime 三塊同時會與 ECL session 記憶體雙向同步：那三塊就是 ECL 位址空間的區 0／1／2（spec 1163），寫入只動 VM 碰過的位址，讀入不收 0 值。

`State.SaveSAVGAMPrefix` 會先以 Area codec 更新已知欄位，再輸出固定前綴；新建 State 沒有原始 raw segment 時，會建立正確尺寸的 zero-filled records。角色姓名只寫入固定 CHRDAT name record，不能由此推導或生成完整 `.SAV/.GUY` player file。

本輪沒有把 prefix export 接替既有 remake JSON F5 save，也沒有宣稱完成 DOS slot 選擇、individual CHRDAT files、pointer resolution 或 atomic multi-file side effects。ECL 記憶體的雙向同步在 spec 1163 補上（區 3 ＝ ECL 程式碼本身仍不寫回）。
