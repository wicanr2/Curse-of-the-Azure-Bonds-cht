# dos-overlay-01 函式明細

由 `cmd/re-ledger` 產生。位址是模組內位址：overlay 為 code-local
offset（base 0），resident executable 為 IDA linear address。

| 位址 | IDA | Borland 符號 | 大小 | 指令 | 被呼叫 | 呼叫 | entry | 狀態 | 等級 | 規格／理由 | 引用 |
|---|---|---|---:|---:|---:|---:|:-:|---|---|---|---|
| `0000` | sub_0 | — | 81 | 28 | 1 | 1 | ✓ | 已解讀 | exact | docs/spec/756-map-fill-confirms-grid-and-assorted-routines.md<br>可中斷的計時等待(retf 2)：06EAh:024Ch() 開頭結尾成對；起點 := 0542h:0994h()(32-bit 計數)，終點 := 起點 + 參數*100；迴圈中 09ABh:02FAh() 非 0 就提早離開，否則等到計數 >= 終點(無號 32-bit 比較) | audit/embedded-strings.md<br>audit/function-index/pc98-overlay-01.md<br>audit/overlay-init-graph.md<br>spec/751-overlay-init-chain-dependency-graph.md<br>spec/756-map-fill-confirms-grid-and-assorted-routines.md |
| `01DF` | sub_1DF | — | 1103 | 551 | 1 | 0 | ✓ | 已解讀 | exact | 916<br>製作群畫面(retf，無參數)，由 spec 915 片頭近呼叫。整支就是 34 行寫死的 印字(欄, 列, 色, 0, 字串)，每行固定 16 條指令(34×16+7 = 551 逐條對上，無漏讀)。色碼 10=說明文字 11=人名 14=分類標題；下半段排兩欄(欄 2 與欄 22)。★決定性證據：建字串助手 0A54h:0634h 只彈掉 @來源，把 @目的 留在堆疊上當下一個呼叫的引數——印字呼叫自己一條參數都沒推。完整版面表見 spec 916 | audit/function-index/dos-overlay-01.md<br>audit/function-strings.md<br>spec/915-title-sequence.md<br>spec/916-credits-screen.md |
| `0634` | sub_634 | — | 320 | 142 | 0 | 3 | ✓ | 已解讀 | exact | 915<br>片頭四張圖(retf，無參數)：資源檔名 'title'，依序載入圖 1..4——圖1(0,0) 停 5，接著 DS:4F98h = 0 就直接離開；圖2(0,0) 無延遲、圖3(列6,欄0Bh) 停 0Ah、圖4(列0,欄0Bh) 配音效 word[25C6h] 停 0Ah；最後清畫面(0542h:0A0Eh)、呼叫本模組 001DFh 製作群名單、再清畫面。⚠DOS 圖 1 從未釋放就被圖 2 蓋掉(記憶體洩漏)，PC-98 有補 | audit/function-index/pc98-overlay-01.md<br>audit/function-strings.md<br>spec/749-combat-teardown-and-battlefield-grid.md<br>spec/915-title-sequence.md<br>spec/916-credits-screen.md |
| `0774` | sub_774 | — | 7 | 5 | 0 | 0 | ✓ | 已解讀 | exact | docs/spec/752-empty-procedures.md<br>空程序：原始 bytes 就是 55 89 E5 89 EC 5D CB(push bp/mov bp,sp/mov sp,bp/pop bp/retf)，無參數無副作用。IDA 匯出漏 byte 把 89 EC 解成 in al,dx 才看起來像 I/O 讀取 | spec/752-empty-procedures.md |
