# 第二百零四輪：ECL continuous event localization

狀態：`READY`

## 原始 image 證據

real-image bounded runs 取得以下 ECL3／ECL4 event segments：

- ECL3：三名邪教徒倒臥、受傷牧師喘息／狂熱咆哮；
- ECL3：城市遭戰火摧毀的區域；
- ECL4：發現小型魔法商店；
- ECL3 cultist segment 後接 `PRESS BUTTON OR RETURN TO CONTINUE.` menu。

## Contract

- 已驗證 segment 依原始順序用 zh-TW catalog 合併成 State `Message`；
- message 在 ECL menu pause 前先保存，因此按鍵等待時仍看得到事件文字；
- raw `RunResult.Text`／原文 `OriginalEvent` 不變；
- 未知句子仍 fallback 原文。

這輪完成的是連續事件文字顯示，不代表已完成事件後的 `CALL` routine、完整對話
分支或所有 ECL text translation。
