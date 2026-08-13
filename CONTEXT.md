# 專案現況

更新日期：2026-08-13（第 559 輪：轉向全模組反組譯盤點）

本檔只保留**目前有效的現況與入口**。歷史敘述已分冊到
[`docs/context/`](docs/context/)，逐行保留原文，不再放在這裡。

## 真相來源優先序

同一件事若多份文件說法不同，依這個順序判定，並主動修掉被推翻的舊斷言：

1. 目前 worktree 的程式碼、JSON 與測試。
2. 原始 bytes、IDA database 與可重現的 runtime trace。
3. 標為 `READY` 且限定範圍明確的 `docs/spec/`。
4. [`docs/knowledge/coab-re-coverage-matrix.md`](docs/knowledge/coab-re-coverage-matrix.md)（系統層完整度）
   與 [`docs/audit/`](docs/audit/) 的可重生清冊（模組／函式層完整度）。
5. [`WORKLIST.md`](WORKLIST.md)（執行順序）。
6. `docs/context/` 分冊與 `docs/project-status.md`（歷史，只作追溯用）。

工作規則的單一權威入口是 [`AGENTS.md`](AGENTS.md)。

## 目前階段：先把反組譯盤點做完，再談語意與實作

2026-08-13 使用者指示停止「邊做邊分析」：先徹底完成 CoAB 的反組譯分析，
再繼續 remake。同日的盤點證實這個轉向有具體理由——

**先前的 RE 是問答式的**：`scripts/ida/` 的 64 支腳本每支只回答當時那一題，
沒有任何一份全模組函式清冊。因此「還剩多少沒看」在此之前無法回答。

**overlay 程式碼幾乎沒被分析過**：DOS `GAME.OVR` 直接載入 IDA 只得到
**1 個函式**；PC-98 同樣。原因是 raw TPOV overlay 沒有 MZ header 也沒有
entry point，自動分析無從開始。真正的 entry point 全在 resident executable 的
`CD 3F` five-byte stub 裡（第 412 輪已證明格式）。

本輪把那份格式知識接成工具鏈後，DOS 側從 1 個函式變成 **1,320 個**
（36 段 overlay ＋ resident，871 個 entry stub 當種子）。這不是新的語意成果，
是把先前看不見的程式碼變成可盤點的對象。

新的驗收口徑（使用者 2026-08-13 決定）：

- **全函式覆蓋台帳**：兩平台每個函式都要在台帳裡有狀態
  （`已解讀`＋推論等級／`不阻塞`＋理由／`待解讀`），不再只有系統層百分比。
- **兩平台都全掃**：DOS 是行為 oracle，PC-98 有 Borland symbols 可當語意
  交叉驗證；同一機制在兩邊的證據要分開記錄，不得因數字相同就合併。

方法、產物與批次順序見
[`docs/spec/559-full-module-re-sweep.md`](docs/spec/559-full-module-re-sweep.md)。

## 現有能力（可驗證，不等於整作完成）

正常新遊戲 session 已可從開場走到散提爾堡世界選單：下水道 → 火刀據點入口至
首領 → 戰後世界 → 阿沙本福德 → 立石群 → 艾森布拉城外 → Hap／熔岩洞／
巫師塔 → 散提爾堡 → 眼魔洞穴（手札 59、Dexam 雙戰、東門離場）。

其餘系統的逐項狀態一律查完整度矩陣，不在本檔重述。

## 尚未完成（摘要）

完整 ECL 有序副作用與 external routine、全城市／全房間 coverage、完整戰鬥
規則與敵方 AI、原機音訊、全量繁中校對、完整存檔相容、開場到結局的同一
session，以及三平台發行。

## 歷史分冊

| 分冊 | 範圍 |
|---|---|
| [`10-architecture-and-boundaries.md`](docs/context/10-architecture-and-boundaries.md) | 獨立 engine／JSON game pack 的分工與早期確認事項 |
| [`20-rounds-through-347.md`](docs/context/20-rounds-through-347.md) | 第 1–347 輪工作序列與成果 |
| [`30-log-2026-07.md`](docs/context/30-log-2026-07.md) | 2026-07-28 至 07-31 |
| [`40-log-2026-08-01-03.md`](docs/context/40-log-2026-08-01-03.md) | 2026-08-01 至 08-03 |
| [`50-log-2026-08-09-13.md`](docs/context/50-log-2026-08-09-13.md) | 2026-08-09 至 08-13 |

分冊只作追溯：裡面的「目前」「下一步」都是當時的狀態，不得當成現況引用。
新的輪次紀錄一律寫進當月分冊，本檔只更新現況段落。
