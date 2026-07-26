# 專案現況

更新日期：2026-07-26

## 目前輪次

第 16 輪：VERTICAL MENU。

第 1 輪 commit：`d87b8c3`，已推送至 GitHub `main`。
第 2 輪 commit：`f46bb3d`，已推送至 GitHub `main`。
第 3 輪 commit：`b2631a9`，已推送至 GitHub `main`。
第 4 輪 commit：`beb9de1`，已推送至 GitHub `main`。
第 5 輪 commit：`9162462`，已推送至 GitHub `main`。
第 6 輪 commit：`26b889c`，已推送至 GitHub `main`。
第 7 輪 commit：`002acab`，已推送至 GitHub `main`。
第 8 輪 commit：`c079c50`，已推送至 GitHub `main`。
第 9 輪 commit：`ec85c89`，已推送至 GitHub `main`。
第 10 輪 commits：`d1aea12`、`a4df55b`，已推送至 GitHub `main`。
第 11–12 輪 commits：`a798531`、`07c9309`、`697ff34`，已推送至 GitHub `main`。
第 13 輪 commits：`a6529c8`、`2489507`，已推送至 GitHub `main`。
第 14 輪 commits：`77375cd`、`b3aa461`，已推送至 GitHub `main`。
第 15 輪 commit：`2d50510`，已推送至 GitHub `main`。

## 已確認

- `curseoftheazurebonds.zip` 是 94 個檔案、約 1.2 MiB 的 DOS 遊戲映像，包含 `START.EXE`、`GAME.OVR`、六組 `ECL*.DAX`、圖像、精靈、地城與怪物資料。
- `START.EXE` 為 16-bit DOS MZ executable；`GAME.OVR` 是約 272 KiB 的 overlay／資料檔候選。
- `ECL1.DAX` 至 `ECL6.DAX` 的大小約 16–27 KiB，應是按章節分組的腳本資料；尚未宣稱其 opcode 或 header 已確定。
- DAX 容器的 2-byte header offset、9-byte block index 與 signed-byte RLE 已由第二輪工具在 ECL/GEO 樣本上驗證。
- `scripts/dax_dump.py` 已能輸出 block metadata 與 ECL 6-bit packed text 取樣；`tests/test_dax_dump.py` 覆蓋正常 RLE 與截斷錯誤。
- `START.EXE` 的 MZ header 與 `GAME.OVR` 的 `TPOV` 前綴已完成位元組級盤點；loader／overlay 的真正載入邊界仍未宣稱確定。
- `internal/dax` 已實作已確認的 DAX container parser；`internal/locale` 與 `assets/locale/zh-TW.json` 建立第一版繁中資源層。
- Go 1.22 容器驗證通過：`go test ./...`，CLI 能解析 `ECL1.DAX` 的 3 blocks。
- `internal/ecl.ParseOperands` 已實作並測試 ECL 的局部 cursor／word framing；尚未執行 opcode。
- `internal/ecl.Trace` 已加入 command metadata 與安全停止行為；可追蹤已知命令，但尚未實作分支或副作用。
- `internal/ecl.DecodePackedText`／`FindPackedTextCandidates` 已能抽取 ECL 英文候選；真實 `YOU ARE AT THE EDGE OF` payload regression test 通過。
- `internal/game.State` 已將 `zh-TW` catalog 接入 Title → Wilderness → Event 的 opening flow，並有狀態轉移測試。
- `cmd/azure-bonds-game` 已接入 Ebiten、鍵盤輸入與外部 TTF/OTF 字型；仍是 prototype，未宣稱完整遊戲。
- Ebiten opening 已改為讀取 `ECL1.DAX` 第一個 block；`game.State.OriginalOpening` 會記錄從原始 payload 辨識出的 opening marker。
- 第十輪驗證通過：internal tests 與 `cmd/azure-bonds-game` Ebiten compile；GUI 仍需有 display 才能實跑。
- `internal/ecl.TraceGraph` 已能追蹤 code-segment `GOTO/GOSUB` targets；不執行條件或副作用。
- 參考公開 CoAB 重寫程式後，確認 ECL 初始化會先連續讀五組 word-valued command-set；`internal/ecl.EntryPoints` 已加入此安全解析器與 regression test。
- 實際 `ECL1.DAX` 三個 block 的 initial entry 都解析為 `0x8014`；CLI `-graph` 已優先使用該入口對應的 payload offset `+0x0014`，不再盲目從 `+0x0000` 開始。
- `ParseOperands` 已正確消耗 `0x80 length payload` compressed-string operand；`TraceAt` 可從 initial entry 開始，block 81 已回歸解出 `AS YOU DEPART...`、`AS YOU LEAVE...` 等原始事件文字。
- `internal/ecl.RunSubset` 已支援 bounded `SAVE/COMPARE/IF/GOTO/GOSUB/RETURN/PRINT`；實際 ECL1 從 initial entry 執行時會在尚未支援的 `0x25 ON GOTO` 或其他副作用命令安全停止。
- `RunSubset` 已加入 `0x14 COMPARE AND`、`0x2A GETTABLE`、`0x2B HORIZONTAL MENU`；實際 ECL1 已讀出 TILVERTON／SHADOWDALE 開場與 `ENTER CITY/JOURNEY ON/CAMP` menu。
- `game.NewStateFromECL` 與 Ebiten opening 已接上原始 menu 的繁中 locale 映射；runner 仍以 deterministic index 0，尚未接完整玩家 menu input。
- `RunSubsetWithSelections`、`game.State.Select` 與 Ebiten cursor 已接上 menu index；實際 selection 1 會走到不同 ECL branch，block 80 停在未支援 `0x15 VERTICAL MENU`。
- `0x15 VERTICAL MENU` 已加入 prompt／options／selection parser；實際 ECL1 可讀到 `INN/STORE/BAR/LEAVE`，後續多層 menu sequence 尚未完整接入 UI state。
- 原始映像與 PDF／RAR 手冊是本地研究素材，第一輪不直接納入 Git 追蹤。

## 尚未確定

- DAX 的 header、壓縮方式、索引與圖像／腳本子格式。
- `GAME.OVR` 與 `START.EXE` 的載入關係。
- ECL opcode、字串編碼、分支／呼叫慣例。
- unknown opcode `0x85` 的完整語意與 IF／menu 的 runtime state 仍未完成。
- 完整 menu rendering／input semantics、CAMP 行為與後續事件仍未完成。
- 中文化的字型格式與字串長度限制。

## 下一步

1. 建立可重現的映像 manifest／樣本檢查工具。
2. 對 `ECL*.DAX`、`GAME.OVR` 做十六進位與反組譯取樣，找出共通 header／索引。
3. 依證據更新 `docs/spec/`，未驗證內容維持 `DRAFT`。
4. 建立 ECL block 的欄位／事件邊界測試，才評估是否能把 ECL 規格標為 `READY`。
5. 將 ECL decoded payload 分成事件 header、控制資料與文字，建立最小場景 trace。
6. 以 operand framing 為基礎加入未知 opcode 安全停止的 trace walker。
7. 對齊 ECL branch targets 與原版場景入口，建立最小可執行事件狀態。
8. 對全部 ECL 建立去重原文 catalog，開始逐項接入繁中翻譯。
9. 建立 Ebiten input／render adapter，先視覺化 opening state。
10. 將 ECL event trace 與 Ebiten event screen 接起來，加入第一個可驗證劇情場景。
11. 建立 ECL branch target graph，將 opening marker 後的選項與事件序列接入 state。
12. 對 ECL graph 的 entry points 做原版事件文字對齊，建立第一個完整 event screen。
13. 用 `EntryPoints` 對實際 ECL1–ECL6 做入口 regression，再逐步加入可執行 VM command subset。
14. 將 `ON GOTO/GOSUB`、選單與 memory model 以 regression 驗證後接入遊戲 state。
15. 將 menu selection 變成 Ebiten input／runner action，完成 Enter City／Journey On 第一個事件分支。
16. 實作 `VERTICAL MENU` 的可觀測選項與 input，再擴展第一個城市事件。
17. 將 successive menu sequence 保存為 game event state，完成城市場所選擇與離開分支。
