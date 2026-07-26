# 專案現況

更新日期：2026-07-26

## 目前輪次

第 1 輪：基線與原始素材盤點。

本輪 commit：`dd625c9`（已在本地獨立 Git metadata 建立）。指定 GitHub remote 的 `ls-remote` 可連線但沒有列出分支；push 則被環境的外部資料輸出審核擋下，待使用者明確核准後重試。

## 已確認

- `curseoftheazurebonds.zip` 是 94 個檔案、約 1.2 MiB 的 DOS 遊戲映像，包含 `START.EXE`、`GAME.OVR`、六組 `ECL*.DAX`、圖像、精靈、地城與怪物資料。
- `START.EXE` 為 16-bit DOS MZ executable；`GAME.OVR` 是約 272 KiB 的 overlay／資料檔候選。
- `ECL1.DAX` 至 `ECL6.DAX` 的大小約 16–27 KiB，應是按章節分組的腳本資料；尚未宣稱其 opcode 或 header 已確定。
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
