# 第一百五十七輪：ECL ADD NPC signal

## 反組譯／實際資料證據

ECL command table 將 opcode `0x36` 定義為 `ADD NPC`，arity 為 1。實際 `ECL1.DAX` block `0x52` 的 prefix 是 `LOAD FILES → CLEAR BOX → PICTURE 0x50 → ADD NPC 0x55 → EXIT`；目前 bounded runner 在 `ADD NPC` 停止，無法證明該 block 的完整 prefix。

## 實作結果

- `RunResult.NPCIDs` 保存 `ADD NPC` operand 經 numeric operand resolver 得到的 ID。
- runner 消耗正確 operand 後繼續執行；synthetic regression 確認 `ADD NPC 0x55 → EXIT` 兩步完成。
- 若本地存在原始映像，real regression 讀取 `ECL1.DAX` block `0x52`，確認 NPC ID `0x55` 並抵達 EXIT；缺少映像時測試明確 skip。
- `BlockSession` 與 CLI 都保留／顯示 NPC ID signal。

## 明確 boundary

本輪不虛構 NPC record、姓名、對話、加入 party、商店或 ECL external routine side effect；這些需要 NPC data member、`WHO`／`ADD NPC` reference routine 與完整劇情 continuation 證據。
