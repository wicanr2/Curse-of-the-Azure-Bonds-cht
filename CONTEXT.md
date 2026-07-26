# 專案現況

更新日期：2026-07-26

## 目前輪次

第 8 輪：繁中開場狀態核心。

第 1 輪 commit：`d87b8c3`，已推送至 GitHub `main`。
第 2 輪 commit：`f46bb3d`，已推送至 GitHub `main`。
第 3 輪 commit：`b2631a9`，已推送至 GitHub `main`。
第 4 輪 commit：`beb9de1`，已推送至 GitHub `main`。
第 5 輪 commit：`9162462`，已推送至 GitHub `main`。
第 6 輪 commit：`26b889c`，已推送至 GitHub `main`。
第 7 輪 commit：`002acab`，已推送至 GitHub `main`。

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
- 原始映像與 PDF／RAR 手冊是本地研究素材，第一輪不直接納入 Git 追蹤。

## 尚未確定

- DAX 的 header、壓縮方式、索引與圖像／腳本子格式。
- `GAME.OVR` 與 `START.EXE` 的載入關係。
- ECL opcode、字串編碼、分支／呼叫慣例。
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
