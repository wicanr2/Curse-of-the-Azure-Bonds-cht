# 第二百七十七輪：ADD NPC 與青色枷 demo 展示序列

狀態：READY

## 推翻的舊 framing

`CMD_AddNPC` 呼叫 `vm_LoadCmdSets(2)`：

1. NPC／monster ID；
2. morale。

真實 ECL1 block `0x52:+0x001F` bytes 是 `ADD NPC 0x55,100`，其後
`+0x0024/+0x0029` 再加入 `0x58/0x5A`。舊 parser 只消耗第一 operand，將第二
operand 的 `0x00` code 誤判成 EXIT，因此舊「五步抵達 EXIT」證據無效。

## Reference party transaction

`load_npc` 在 party size ≤ 7 時：

- 從目前 `MON{area}CHA` 載入共用 Player record；
- 從同 ID `MON{area}SPC/ITM` 載入 effects／items；
- 將 `mod_id` 設成 NPC ID；
- 加入 TeamList、選為 SelectedPlayer；
- 指派最低未使用的 0..7 icon slot；
- `control_morale = (morale >> 1) + 0x80`；
- 重算 player values 與 party summary。

State 現依 ECL chapter 使用相同三份 DAX table，建立 persistent Character 與 party
Fighter。MON NPC record 可有 stale `class_id`；若八個 ClassLevel 中恰有一個非零，
NPC parser依該 slot 推導 class，普通玩家 save parser仍維持嚴格驗證。

## Real opening evidence

後續 reference `sub_29758` 證明 block `0x52` 僅在 `gbl.inDemo == true` 載入，
執行後會清空 TeamList；它是 attract/demo sequence，不是玩家建立隊伍後的正式 new game。
正式流程已於第 278 輪接到 global block `0x01`。本節保留 demo sequence 的真實證據。

block `0x52:+0x0014` 現執行 53 steps：

- 加入 RUSTLE（fighter 9）、CYNTHIA（magic-user 9）、GRENDEL（cleric 9）；
- 三人的 morale operand 100 轉成 control morale `0xB2`；
- 輸出 11 段原始序幕文字；
- 聚合 11 次 `CALL 0x6803` picture-frame routine 與 12 次 `DELAY`；
- 最後在 `COMBAT` 停止，保留 monster setup／spawn。

State 若同一 result 同時包含 PICTURE 與 COMBAT，會先顯示事件圖片並保存 combat；
demo adapter／direct regression Continue 後會用剛加入的隊伍建立 Battle，不再遺失 signal。
11 段原文逐行保留在 VM evidence，State／zh-TW catalog 則組成繁中序幕，涵蓋伏擊、
五個青色符印、金屬枷鎖、聯盟成員與重新掌握命運。

## Boundary

`CALL 0x6803` 對應 reference `DrawMaybeOverlayed → NextFrame → GameDelay`；目前 Ebiten
PIC renderer 已有時間式多 frame animation，但尚未逐 pulse 重播 ECL 的 11 次 timing。
NPC 的完整 AI morale 行為、treasure share、demo 結束清隊 UI 與 SAVGAM 新增 sidecar
serialization 仍待驗證。正常玩家流程不可自動加入這三名 demo NPC。

```text
go test ./internal/ecl ./internal/monster ./internal/party ./internal/game ./internal/save
```
