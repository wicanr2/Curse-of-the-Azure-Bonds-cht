# 292：皇家馬車、Fire Knives 強制攻擊與盜賊公會轉場

狀態：READY

## 證據

- 使用者提供的 clue／journal PDF、GameFAQs 地圖與真實 `GEO2.DAX`／ECL2 block 1
  一致指出：Tilverton `(1,0)` 是城門事件位置。
- 先造訪 Weaponers `(2,12)` 與 Filani `(6,5)`，且曾被城門衛兵驅回一次後，再次進入
  `(1,0)` 會顯示 `PICTURE 11`。國王聲音令青色枷發光，強迫隊伍攻擊皇家馬車。
- 原 ECL 接著載入五名 `ROYAL GUARD`，戰勝後顯示紅袍人劫走假國王，並詢問
  `DO YOU SURRENDER?`。
- YES 分支將隊伍投入牢房；`PICTURE 2／HEAD2／BODY2` 的盜賊歸還裝備並帶路，
  最後 `NEWECL` 至 block `0x02`，map registers 為 `(1,12,0)`。

## 實作契約

- 第一次城門警告與第二次皇家馬車事件必須由同一 resumable ECL memory 決定，
  不可由 renderer 以「已看過畫面」布林值猜測。
- PICTURE 11、四段 carriage／bond 敘事、五名真實 MON2 Royal Guard 戰鬥、戰後
  投降選單、牢房、盜賊救援與 block 2 transition 都必須保留原始順序。
- COMBAT 必須載入 `MON2CHA.DAX` 並建立 active battle；缺資料時顯示 placeholder
  不算通過主線 gate。
- 繁中翻譯由作品 State adapter 處理；ECL 原始 YES／NO 與 PC/memory continuation
  不變。
- `-carriage` 可在同一正式 new-game session 依序跑過 Weaponers、Filani 與第一次
  gate warning，再停在第二次 PICTURE 11，供 640×480 畫面重現。

## 驗證

- real-image regression：正式建立角色 → Weaponers → Filani → 第一次城門 →
  PICTURE 11 → 青色枷強制攻擊 → 五名 Royal Guard active combat → 勝利 →
  surrender YES → jail → PICTURE 2 thief → ECL block 2／`(1,12,0)`。
- 戰鬥 gate 同時鎖定六名 fighters（單一 test hero＋五名衛兵）與五個 enemy targets。
- 結尾不得保留上一個 press-button choice。
