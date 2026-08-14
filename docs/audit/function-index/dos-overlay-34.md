# dos-overlay-34 函式明細

由 `cmd/re-ledger` 產生。位址是模組內位址：overlay 為 code-local
offset（base 0），resident executable 為 IDA linear address。

| 位址 | IDA | Borland 符號 | 大小 | 指令 | 被呼叫 | 呼叫 | entry | 狀態 | 等級 | 規格／理由 | 引用 |
|---|---|---|---:|---:|---:|---:|:-:|---|---|---|---|
| `0000` | sub_0 | — | 7 | 5 | 0 | 0 | ✓ | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>空函式：prologue／epilogue 之外沒有任何指令，呼叫即返回（body 共 7 bytes，已逐條讀完） | audit/embedded-strings.md<br>audit/function-index/dos-overlay-04.md<br>audit/function-index/dos-overlay-15.md<br>audit/function-index/dos-overlay-16.md<br>audit/function-index/dos-overlay-17.md<br>audit/function-index/dos-overlay-25.md |
| `0240` | sub_240 | — | 1735 | 774 | 0 | 1 | ✓ | 已解讀 | exact | 1083<br>ECL 助憶碼表(retf 2)：整支是 65 路 case，把 DS:75FFh(指令碼) 換成字串常數寫進 arg_2(上限 0FFh)。★★★原作自己留下的 ECL 指令名稱，一次把幾十份 ECL 規格的名字釘到確切指令碼上。★★★29h = ENCOUNTER MENU 不是 PARLAY(PARLAY 是 2Ch) ⇒ 更正 spec 611/1041 的名字(行為描述不受影響)；2Bh = HORIZONTAL MENU 與 spec 1039 吻合。★1Fh 完全沒有分支 ⇒ 未使用的指令碼，arg_2 不被寫入。★字串在資料段的排列順序與指令碼不一致(LOAD PIECES 排在 21h 與 22h 之間但對應 37h) ⇒ 不能靠排列推碼。SUBTRAT 少一個 C、IF = 尾端多空白，都是原作瑕疵 | audit/embedded-strings.md<br>audit/function-strings.md<br>audit/function-triage.md<br>spec/1083-ecl-opcode-mnemonics.md |
| `0907` | sub_907 | — | 7 | 5 | 0 | 0 | ✓ | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>空函式：prologue／epilogue 之外沒有任何指令，呼叫即返回（body 共 7 bytes，已逐條讀完） | — |
