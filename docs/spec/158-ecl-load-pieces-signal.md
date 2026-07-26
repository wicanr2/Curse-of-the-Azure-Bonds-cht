# 第一百五十八輪：ECL LOAD PIECES signal

## 反組譯／實際資料證據

ECL command table 將 opcode `0x37` 定義為 `LOAD PIECES`，arity 為 3。原始映像掃描顯示它在 ECL2 block `0x01`、ECL3、ECL5、ECL6 多個實際 entry 反覆出現；ECL2 block `0x01:+0x0014` 的前綴明確為 `LOAD FILES 1,2,0xFF → LOAD PIECES 1,2,3 → SAVE ...`。

## 實作結果

- `RunResult.LoadPiecesRequested` 與 `[3]uint16 LoadPieces` 保存三個已解析 selector。
- runner 消耗 `LOAD PIECES` 的三個 operand 後繼續執行，不再以 unsupported opcode 停止。
- `BlockSession` 聚合 signal，CLI 會顯示 `LOAD PIECES=[...]`；synthetic regression 覆蓋三個 selector 與後續 EXIT。

## 明確 boundary

本輪不猜測三個值各自對應的 DUNGCOM／8X8D／WALLDEF／TILES 檔案與拼圖寫入規則，也不把 signal 當成完成地城 renderer；實際 floor construction、wall side effects、碰撞與 camera 仍需 file loader／反組譯證據。
