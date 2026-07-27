# 第一百五十七輪：ECL ADD NPC signal

## 反組譯／實際資料證據

ECL command table 將 opcode `0x36` 定義為 `ADD NPC`。本輪早期依 command metadata
誤判為一個 operand；第 277 輪依 `CMD_AddNPC.vm_LoadCmdSets(2)` 與 raw bytes 修正為
兩個 operand：NPC ID 與 morale。原先被判成 `EXIT` 的 `0x00` 其實是 morale operand code。

## 實作結果

- `RunResult.NPCIDs` 保留相容 ID view；第 277 輪新增 `NPCRequests{ID,Morale}`。
- synthetic regression 現確認 `ADD NPC 0x55,100 → EXIT` 的正確 framing。
- real regression 已更新：block `0x52` 連續加入 `0x55/0x58/0x5A`、morale 均為 100，
  播放完整 demo 展示序列，最後在 COMBAT 停止；不存在先前宣稱的早期 EXIT。
- `BlockSession` 與 CLI 都保留／顯示 NPC ID signal。

## 明確 boundary

第 277 輪已依 `load_npc/load_mob/AssignPlayerIconId` 接入 MON*CHA／SPC／ITM、
control morale、最低空 icon slot、selected player 與 party/fighter insertion。其他作品
NPC table、八人上限外的劇情替代流程與完整 NPC AI 仍需各自驗證。
