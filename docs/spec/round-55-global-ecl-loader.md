# 第五十五輪：全 ECL global block namespace

狀態：`READY`（限原始 ECL DAX loader 與 real `NEWECL` entry regression）

## 已確認行為

原始 `NEWECL` operand 使用跨資料檔的 global block ID；例如 ECL4 block `0x25` 的 `NEWECL 0x50`，其 target 實際位於 ECL1，而不是 ECL4 內。ECL5 block `0x30` 也有同樣的 `NEWECL 0x50`。

`cmd/azure-bonds-game` 現在啟動時載入 `ECL1.DAX` 至 `ECL6.DAX`，檢查 global block ID 不重複，並以 ECL1 的第一個 block 作為 opening entry。這讓 `BlockSession` 能找到真實 target，而不會在 loader boundary 停止。

## 驗證

將 ECL1／ECL4／ECL5 合併成同一 namespace 後，從 ECL4 block `0x25:+0x022B` 執行：

- session 真實切換到 ECL1 block `0x50`；
- target initial entry 解出 `YOU ARE AT THE EDGE OF`、`TILVERTON` 與後續詢問文字；
- runner 正確停在 `ENTER CITY／JOURNEY ON／CAMP` menu。

## 邊界與未完成項目

本輪完成的是 DAX block namespace 與 entry-level real transition，不是完整劇情流程。ECL6 全部 event entry 的 real reachability、未支援 opcode、外部 `PROGRAM` routine、原始地圖／場所／戰鬥後 continuation 仍待完成。
