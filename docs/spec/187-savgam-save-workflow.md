# 第一百八十七輪：SAVGAM save workflow

狀態：`READY`

## 目的

讓玩家在載入原版 DOS slot 後，透過 remake 的 F5 或 `CAMP → SAVE` 寫回同一個 slot；一般 remake party 仍使用 versioned JSON save。

## 行為

- `-savgam-dir <dir> -savgam-slot A` 先呼叫 `LoadSAVGAMSlot`。
- 該模式建立 Ebiten save adapter；F5 與 `ConsumeSaveRequest` 都呼叫 `State.SaveSAVGAMSlot(dir, key)`。
- 未使用 SAVGAM loader 時，兩個入口仍呼叫 `SavePartyFile(party.json)`。
- 成功訊息顯示實際 prefix target；錯誤留在 State message，不終止遊戲 loop。

## 邊界

這是已載入 slot 的 workflow adapter，不是新建角色直接產生完整原版 DOS save。原版 player delete／rename、未知 sidecar、多職業欄位與跨檔案 atomic commit 仍須個別反組譯證據。
