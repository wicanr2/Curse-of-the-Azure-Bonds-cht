# pc98-overlay-01 函式明細

由 `cmd/re-ledger` 產生。位址是模組內位址：overlay 為 code-local
offset（base 0），resident executable 為 IDA linear address。

| 位址 | IDA | Borland 符號 | 大小 | 指令 | 被呼叫 | 呼叫 | entry | 狀態 | 等級 | 規格／理由 | 引用 |
|---|---|---|---:|---:|---:|---:|:-:|---|---|---|---|
| `0000` | sub_0 | — | 81 | 28 | 1 | 1 | ✓ | 已解讀 | exact | docs/spec/756-map-fill-confirms-grid-and-assorted-routines.md<br>可中斷的計時等待(retf 2)：07E3h:02C7h／0418h:133Dh／08EEh:03B3h，其餘與 DOS overlay-01:0000h 同 | audit/embedded-strings.md<br>audit/far-call-map-dos.md<br>audit/far-call-map-pc98.md<br>audit/function-index/pc98-overlay-01.md<br>audit/overlay-init-graph.md<br>audit/pc98-audio-state-writers.md |
| `0321` | sub_321 | — | 1551 | 775 | 1 | 0 | ✓ | 已解讀 | exact | 916<br>同 DOS overlay-01:001DFh，48 行(48×16+7 = 775 逐條對上)。★80 欄把同一份名單從兩欄改排四欄(欄 2/13/23/32)，列數 22 壓到 14，下半頁多 14 行 PonyCanyon 移植組名單 + 'ＶＥＲＳＩＯＮ１．２'(全支唯一全形字)。另：全小寫改成標題全大寫/人名首字大寫、'tsr, inc.'→'TSR,INC.'(逗號無空格)、DOS 分兩列的 'jeff grubb'+'george mac donald' 併成一個字串 'Jeff Grubb,George Mac Donald'(版面因素，非文法)。用兩支印字助手 418h:0D17h(26 次)與 418h:0DA5h(70 次)，分流原因無結論 | audit/function-strings.md<br>audit/function-triage.md<br>spec/915-title-sequence.md<br>spec/916-credits-screen.md |
| `0936` | sub_936 | DOINTRO | 364 | 162 | 0 | 3 | ✓ | 已解讀 | exact | 915<br>同 DOS overlay-01:00634h，但★在旗標檢查後補了 釋放(@圖)——修掉 DOS 漏放第一張圖的洩漏(PC-98 修 DOS 缺陷的第三例，前有 spec 883)。另：進入時 DS:8BF3h := 1 並在第一張圖後叫 893h:0114h(該值)(形狀上是片頭音樂開關)、載入呼叫多一個 byte 參數(retf 0Ah→0Ch)、圖 2 之後多 延遲(5)、圖 3 的列 6→5、音效改 893h:0 配 word[4854h]、清畫面改 0418h:13B6h | audit/function-strings.md<br>spec/915-title-sequence.md |
| `0AA2` | sub_AA2 | — | 6 | 4 | 0 | 0 | ✓ | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `pop bp`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 6 bytes，已逐條讀完） | — |
