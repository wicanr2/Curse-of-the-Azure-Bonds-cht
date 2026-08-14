# pc98-overlay-20 函式明細

由 `cmd/re-ledger` 產生。位址是模組內位址：overlay 為 code-local
offset（base 0），resident executable 為 IDA linear address。

| 位址 | IDA | Borland 符號 | 大小 | 指令 | 被呼叫 | 呼叫 | entry | 狀態 | 等級 | 規格／理由 | 引用 |
|---|---|---|---:|---:|---:|---:|:-:|---|---|---|---|
| `0000` | sub_0 | LOADCLOCK | 32 | 10 | 0 | 0 | ✓ | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>呼叫序列：依序執行 5 個呼叫，沒有其他動作：`call far ptr 19Ah:2Ah`、`call far ptr 117h:57h`、`call far ptr 13Eh:9Dh`、`call far ptr 14Ah:101h`、`call far ptr 164h:57h`（body 共 32 bytes，已逐條讀完） | audit/embedded-strings.md<br>audit/function-index/dos-overlay-15.md<br>audit/overlay-init-graph.md<br>audit/string-pairs.md<br>spec/754-small-predicates-and-wrappers.md |
| `0020` | sub_20 | — | 696 | 239 | 1 | 1 | ✓ | 待解讀 | — | — | spec/765-bit-reverse-two-sided-wall-check-and-zero-pad.md |
| `02D8` | sub_2D8 | — | 173 | 69 | 2 | 1 | ✓ | 待解讀 | — | — | — |
| `0385` | sub_385 | — | 52 | 20 | 2 | 2 | ✓ | 已解讀 | exact | docs/spec/754-small-predicates-and-wrappers.md<br>上限 99 的累加(retf)：對應位址 0A652h/0A65Ch/680Ch/0A65Ah，與 DOS overlay-20:0385h 同 | audit/function-index/pc98-overlay-20.md<br>spec/754-small-predicates-and-wrappers.md |
| `03B9` | sub_3B9 | TICKCLOCK | 165 | 67 | 1 | 3 | ✓ | 待解讀 | — | — | audit/embedded-strings.md |
| `045E` | sub_45E | — | 233 | 92 | 2 | 2 | ✓ | 待解讀 | — | — | — |
| `0549` | sub_549 | — | 126 | 54 | 1 | 1 | ✓ | 已解讀 | strong inference | docs/spec/765-bit-reverse-two-sided-wall-check-and-zero-pad.md<br>與 DOS overlay-20:0549h（entry#9）助憶碼序列完全相同，語意同該筆：兩位數補零(retf 2)：先 14Dh:0052h 把數值轉成字串，n < 0Ah(無號)時在前面接上 CS:0547h 的 '0'，最後 0A54h:0680h(緩衝A, 緩衝B, 1, 2) 取固定寬度再寫回輸出參數。⚠ 補的是 ASCII '0'，換全形數字會讓寬度計算不一致 ⚠ 運算元中的 DS／overlay-local 位址兩平台不同，引用位址前須各自確認 | audit/function-index/pc98-overlay-20.md<br>spec/765-bit-reverse-two-sided-wall-check-and-zero-pad.md |
| `05D2` | sub_5D2 | — | 246 | 110 | 5 | 2 | ✓ | 待解讀 | — | — | — |
| `073A` | sub_73A | — | 329 | 140 | 1 | 4 | ✓ | 待解讀 | — | — | — |
| `0898` | sub_898 | — | 173 | 65 | 1 | 2 | ✓ | 已解讀 | strong inference | docs/spec/768-script-arithmetic-opcodes-and-save-size.md<br>與 DOS overlay-20:0862h（entry#12）助憶碼序列完全相同，語意同該筆：每 288 次才做一次的整隊回復(retf 2)：inc(DS:757Ch)，未達 120h(288)就離開；否則走隊伍鏈對每人叫 <呼叫>(1, 角色)，參數非 0 時再叫本模組 05D4h(0)，畫出 CS:0848h 'The Whole Party Is Healed'(25 字元，x=1、y=13h、色 0Ah)，最後計數器歸零。⚠ 訊息寫死在 x=1、沒有先清行，換短的譯文會留下殘影 ⚠ 運算元中的 DS／overlay-local 位址兩平台不同，引用位址前須各自確認 | audit/embedded-strings.md |
| `094E` | sub_94E | — | 182 | 68 | 2 | 2 | ✓ | 待解讀 | — | — | — |
| `0A11` | sub_A11 | — | 294 | 105 | 2 | 2 | ✓ | 待解讀 | — | — | — |
| `0B37` | sub_B37 | — | 177 | 67 | 1 | 3 | ✓ | 待解讀 | — | — | audit/embedded-strings.md |
| `0BE8` | sub_BE8 | — | 178 | 65 | 1 | 3 | ✓ | 待解讀 | — | — | — |
| `0CD0` | sub_CD0 | REST | 492 | 196 | 0 | 8 | ✓ | 待解讀 | — | — | — |
| `0EBC` | sub_EBC | — | 7 | 5 | 0 | 0 | ✓ | 已解讀 | exact | docs/spec/752-empty-procedures.md<br>空程序：原始 bytes 就是 55 89 E5 89 EC 5D CB(push bp/mov bp,sp/mov sp,bp/pop bp/retf)，無參數無副作用。IDA 匯出漏 byte 把 89 EC 解成 in al,dx 才看起來像 I/O 讀取 | audit/embedded-strings.md<br>spec/752-empty-procedures.md |
