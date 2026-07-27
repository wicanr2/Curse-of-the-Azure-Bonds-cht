# 第一百九十九輪：真實 ECL1→ECL2 transition

狀態：`READY`

## Evidence

原始 ECL1 block `0x50` payload `+0x5B5` 的指令是 `NEWECL 0x03`。global ECL namespace 中 block `3` 位於 ECL2.DAX；不是 ECL1 的 local block。

## Regression

`TestRealECL1ToECL2NEWECLSwitch` 載入原始 `ECL1.DAX` 與 `ECL2.DAX`，建立只含 source `0x50`／target `3` 的 `BlockSession`，從 `+0x5B5` 執行 bounded run，驗證 current block 已切換到 `3`。target entry 若在後續 unsupported routine 停止，不會否定已完成的 switch；測試要求 transition 已套用且仍有 bounded result。

## Interaction with monster tables

切換後 `State.monsterRecordsForCurrentECL()` 會選 ECL2 的 `MON2CHA` table，避免與 ECL1 相同 numeric ID 碰撞。這條 regression 只驗證 script/session transition，不宣稱一般玩家從 ECL1 opening 已能自然抵達此 entry。

## Reuse boundary

後續 Golden Box 遊戲可沿用 global block loader、shared runtime context 與 target-first stop semantics，但要以各作品 DAX namespace 證明 target file；不能將 block ID 當作檔案內 local index。
