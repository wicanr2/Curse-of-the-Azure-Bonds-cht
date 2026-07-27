# 291：提爾佛頓高階祭司與手札 19

狀態：READY

## 證據

- `GEO2.DAX` block 1 的 `(1,10)` terrain selector 是 `0x8F`。
- 真實 ECL2 SearchLocation 顯示 `PICTURE 6`、`HEAD6`、`BODY6`。高階祭司詢問隊伍
  是否願意說明遭遇，YES 分支施展 `REMOVE CURSE SPELL` 並要求記錄
  `JOURNAL ENTRY 19`。
- 使用者提供的 Adventure Journal 掃描 PDF 第 15 頁印刷頁碼包含 Entry 19 全文：
  施法令青色枷發光、射出藍焰並使眾人劇痛；祭司停止施法，承認自己的神力無法解除。

## 實作契約

- 高階祭司問答、施法結果與離場文字由作品 State adapter 翻成繁中；VM 仍使用原始
  `YES/NO` 選項索引。
- 手札 19 只在玩家選 YES 且真實 ECL 發出 Entry 19 後解鎖；文字忠實翻譯使用者
  提供的 PDF，不提前揭露。
- 兩個 press-button pause 都必須保留：先閱讀施法／手札訊息，再閱讀離場訊息；
  最後回到 `(1,10)` 並清除 stale choices。
- 事件畫面沿用 640×480 獨立 layout：HEAD／BODY nearest-neighbor 3×，繁中 24px
  直接在放大後畫布 rasterize。

## 驗證

- real-image regression 從正式 new-game session 進入 selector `0x8F`，鎖定
  PICTURE／HEAD／BODY、YES/NO、手札 19 內容、兩段 continuation 與同格返回。
- `-high-priest` 提供可重現的 640×480 畫面入口。
