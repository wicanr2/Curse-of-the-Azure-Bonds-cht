# remake 目前量得到什麼

由 `cmd/remake-status` 從 `docs/audit/` 既有的報表併出來，不要手改。

⚠ **這不是完成度百分比。** 每個分母的意義都不同，平均起來只會得到一個沒有意義而且偏樂觀的數字。每一列的「證明不了什麼」跟數字一樣重要。

⚠ **「沒有數字」的那幾列才是重點。** 量得到的部分已經接近收斂，剩下的工作大多落在還沒有分母的類別上——那也是為什麼手寫的「尚未完成」清單會同時漏掉已完成的項目、又講不清楚剩下的工作有多大。

| 類別 | 目前量到 | 出處 | 這個數字證明不了什麼 |
|---|---|---|---|
| ECL 有序副作用與 external routine | 可達指令 14177 條：`done` 13690／`partial` 487；opcode `done` 58／`partial` 3 | `ecl-effect-coverage.json`（`cmd/ecl-effect-coverage`） | `done` 是「這個 opcode 的原作語意讀完並接上」，**不是**「玩家走得到的每一條都驗過」。可達性來自控制流走訪，跟不到的碼不在分母裡。 |
| 原作玩家可見文字接線 | 1022 頁：matched 1002、**unmatched 0** | `ecl-text-coverage.json`（`cmd/ecl-text-coverage`） | 「有一條 `text_rule` 命中」不等於譯文正確，也不等於那一頁的事件副作用已還原。 |
| 存檔相容（角色記錄） | 422 bytes：decoded 299／documented 123／unknown 0；remake 實際消費 291 | `dos-save-field-coverage.json`（`cmd/save-field-coverage`） | 欄位讀得懂不等於**往返**得回去。`MOVEPARTY` 跨遊戲轉移、角色刪除／改名尚未 round-trip。 |
| 反組譯覆蓋（兩平台全模組） | 2874 個函式：已解讀 2137／不阻塞 162／邊界碎片 575／**待解讀 0**；未被 IDA 認成程式碼的位元組 36386 | `coab-function-index.json`（`cmd/re-ledger`） | 狀態是**人工判定**寫在 `re-function-ledger.json` 裡的，工具只讀不寫。「已解讀」是有人讀過並留下 spec，不是自動驗證。「未被 IDA 認成程式碼」的那些位元組多半是資料，但**沒有被逐段判過**。 |
| 譯名一致性 | 詞條 138 條，未解決的不一致 **0** 筆 | `glossary.json`（`cmd/glossary-audit`） | 閘擋不住「誤用的寫法剛好是別的詞條的正確譯名」。那一半只能逐句對照原文。 |
| 玩家文字資料分離（Go 漢字 literal 債） | 合計 862 次（`localization_debt` 0／`frontend_ui_debt` 15／`runtime_ui_debt` 847） | `go-han-literals-baseline.json`（`cmd/coab-audit`） | 這是**債務**不是進度：數字要往下降。分類只是遷移排序用的 heuristic。 |
| 戰鬥中可用物品（充能效果） | 充能物品 18 件，效果**沒接 0** 個 | `combat-item-use.md`（`cmd/item-use-audit`） | 「已接」是 remake 有那一支 handler 的行為讀法，不是逐件實機驗過。 |
| 原機音訊（音樂／音效生命週期） | **沒有數字** | — | PC-98 音樂與音效的 producer 已有規格，但**播放生命週期與戰鬥 phase 同步沒有覆蓋率報表**。 |
| UI fidelity（逐張與原版比對） | **沒有數字** | — | 本機沒有原版逐張畫面，所以「像不像」目前**不是一個數字**。要先建原版截圖 oracle。 |
| 開場到結局的同一 session | **沒有數字** | — | 端到端測試跑得到結局，但那是**測試路徑**；真人從開場玩到結局沒有自動化的證明。 |
| 三平台發行 | **沒有數字** | — | 專案明文決定：遊戲完整可玩之前不花時間做（`WORKLIST.md` P2）。 |
| 全城市／全房間 coverage | **沒有數字** | — | 文字接線有數字（見上），但「每個可選房間、全部寶物、失敗／重訪分支都走過」**沒有**分母。 |
| 敵方 AI 與完整戰鬥規則 | **沒有數字** | — | 個別規則有規格與測試，整體行為**沒有**對照原版的覆蓋率報表。 |
