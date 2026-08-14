# dos-overlay-09 函式明細

由 `cmd/re-ledger` 產生。位址是模組內位址：overlay 為 code-local
offset（base 0），resident executable 為 IDA linear address。

| 位址 | IDA | Borland 符號 | 大小 | 指令 | 被呼叫 | 呼叫 | entry | 狀態 | 等級 | 規格／理由 | 引用 |
|---|---|---|---:|---:|---:|---:|:-:|---|---|---|---|
| `0000` | sub_0 | — | 62 | 16 | 0 | 8 | ✓ | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>呼叫序列：依序執行 11 個呼叫，沒有其他動作：`call loc_1046+4`、`call loc_18B2+1`、`call far ptr loc_1951+2`、`call far ptr sub_1847`、`call far ptr loc_16BC+1`、`call far ptr sub_15D1`（body 共 62 bytes，已逐條讀完） | audit/embedded-strings.md<br>audit/function-index/dos-overlay-02.md<br>audit/overlay-init-graph.md<br>context/50-log-2026-08-09-13.md<br>spec/508-pc98-general-target-scan-producer.md<br>spec/583-ledger-denominator-repair.md |
| `004D` | sub_4D | — | 512 | 186 | 0 | 14 | ✓ | 待解讀 | — | — | audit/function-index/dos-overlay-14.md<br>spec/777-item-append-and-area-save-sweep.md |
| `024D` | sub_24D | — | 100 | 33 | 1 | 1 | ✓ | 已解讀 | exact | docs/spec/757-vroomm-stub-rebuild-image-blit-and-class-slots.md<br>retf 4：p^[18Dh]^[11h] 非 0 回 false；否則 (p^[109h] <= 0) and (p^[111h] <= p^[0E6h]) 也回 false；再叫一支帶輸出參數的 overlay 函式，回 0 就 false；都通過才叫另一支並回 true。比較都是有號 | audit/function-index/dos-overlay-14.md<br>audit/function-index/pc98-overlay-09.md<br>spec/757-vroomm-stub-rebuild-image-blit-and-class-slots.md<br>spec/783-cross-platform-pairs-first-batch.md |
| `02B1` | sub_2B1 | — | 256 | 96 | 1 | 4 | ✓ | 已解讀 | exact | docs/spec/777-item-append-and-area-save-sweep.md<br>範圍法術的豁免掃描(retf 8)：施法者陣營(DS:6506h^[197h])為 0 時修正 −2、否則 +8；先叫掃描程序(地圖 DS:6E92h, 1, 0FFh, byte[37E9h + 法術*10h], y, x)，再對 DS:6E96h 筆結果逐一取 byte[6E94h + i*3] 當索引、由 DS:6D35h 取戰鬥員；只處理不同陣營且 byte[37E2h + 法術*10h] <> 1 的目標，<豁免>(修正, byte[37E3h + 法術*10h], 目標) 回 0 就把結果設 true。⚠ 豁免修正取決於施法者陣營，兩邊不對稱 | audit/embedded-strings.md<br>audit/function-index/dos-overlay-13.md<br>spec/777-item-append-and-area-save-sweep.md |
| `03B1` | sub_3B1 | — | 249 | 97 | 2 | 5 | ✓ | 待解讀 | — | — | spec/687-far-call-flattening-and-stack-leftover.md |
| `04AA` | sub_4AA | — | 347 | 122 | 1 | 4 | ✓ | 待解讀 | — | — | — |
| `0605` | sub_605 | — | 304 | 111 | 1 | 3 | ✓ | 待解讀 | — | — | — |
| `0735` | sub_735 | — | 630 | 246 | 1 | 4 | ✓ | 待解讀 | — | — | audit/embedded-strings.md |
| `09C7` | sub_9C7 | — | 671 | 242 | 2 | 8 | ✓ | 待解讀 | — | — | audit/function-index/dos-overlay-09.md |
| `0C53` | sub_C53 | — | 5 | 2 | 1 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `idiv cx`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 5 bytes，已逐條讀完） | — |
| `0C58` | sub_C58 | — | 10 | 4 | 2 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `les di, es:[di+18Dh]`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 10 bytes，已逐條讀完） | — |
| `0C62` | sub_C62 | — | 9 | 2 | 2 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `cmp byte ptr ds:47F1h, 1`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 9 bytes，已逐條讀完） | — |
| `0C7B` | sub_C7B | — | 41 | 16 | 2 | 2 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `push word ptr [bp+6]`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 41 bytes，已逐條讀完） | — |
| `0CB7` | sub_CB7 | — | 228 | 75 | 3 | 7 |  | 邊界碎片 | — | docs/spec/587-ecl-handler-21-37-shared.md<br>邊界碎片：落在 09C7h 的 prologue 區間內部，自己不是 prologue。所屬函式尚未解讀，讀它時會一併涵蓋。 | — |
| `0DB1` | sub_DB1 | — | 1043 | 349 | 1 | 15 | ✓ | 待解讀 | — | — | — |
| `11CF` | sub_11CF | — | 38 | 15 | 2 | 2 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：有 `pop bp` 收尾卻沒有 `push bp` 開頭；還原的是別人建立的 frame，屬被切開的後半段（body 共 38 bytes，已逐條讀完） | — |
| `1201` | sub_1201 | — | 95 | 31 | 5 | 4 | ✓ | 已解讀 | exact | docs/spec/756-map-fill-confirms-grid-and-assorted-routines.md<br>兩條路的選擇(retf 4)：先無條件叫一支程序；若 f1(p) 或 f2(p) 或 p^[18Dh]^[3] = 0 則回 h1(p)，否則回 h2(p) | audit/function-index/pc98-overlay-09.md<br>spec/756-map-fill-confirms-grid-and-assorted-routines.md |
| `1273` | sub_1273 | — | 248 | 84 | 2 | 3 | ✓ | 已解讀 | exact | docs/spec/774-packbits-rle-and-combat-hotkeys.md<br>戰鬥中的三個熱鍵(retf 4)：KeyPressed 為假就離開；ReadKey(回 0 就再讀一次擴充碼)。'2' 切換 DS:75DAh 並顯示 CS:1260h 'Magic On' / CS:1269h 'Magic Off'；空白把 +0F7h < 80h(士氣旗標未設)且 +195h <> 1 的成員 +198h 清 0，若參數角色的 +198h 也是 0 就把 +18Dh^[3] 設 14h 並回 true；'-' 走另一支。⚠ 擴充鍵的第二個位元組會直接跟 32h/20h/2Dh 比對，掃描碼相同就會誤判 | audit/embedded-strings.md<br>spec/774-packbits-rle-and-combat-hotkeys.md |
| `1388` | sub_1388 | — | 197 | 67 | 1 | 3 | ✓ | 已解讀 | strong inference | docs/spec/758-morale-field-0f7h-round-trip.md<br>與 PC98 overlay-09:13A5h（entry#9）助憶碼序列完全相同，語意同該筆：士氣判定(retf 4)：兩段訊息 CS:1385h 'は退散させられた。' 與 CS:1398h 'は降伏した。'。p^[18Eh]^[10h] 非 0 走第一條；否則要 p^[0F7h] > 7Fh，取 (and 7Fh)*2(>66h 歸 0)，再與受傷百分比 100−(p^[1A5h]*100 div p^[78h]) 及 100−DS:7F09h^[58Ch] 兩次比較，最後 a <= b div 2 或 p^[13h] > 5 決定結局。兩條路都設 p^[18Eh]^[14h]:=1。⚠ 舊 DS:0A03Ch 存進 [bp-3] 後整支沒再讀(死存放，不會還原) ⚠ 運算元中的 DS／overlay-local 位址兩平台不同，引用位址前須各自確認 | — |
| `1444` | sub_1444 | — | 7 | 3 | 3 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `mov al, ds:6FA2h`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 7 bytes，已逐條讀完） | — |
| `1458` | sub_1458 | — | 3 | 1 | 3 | 1 |  | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>尾呼叫：最後一條是 `jmp loc_1539`，控制權轉交後不返回（body 共 3 bytes，已逐條讀完） | — |
| `150E` | sub_150E | — | 6 | 2 | 2 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `push word ptr [bp+6]`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 6 bytes，已逐條讀完） | — |
| `1542` | sub_1542 | — | 13 | 5 | 1 | 0 | ✓ | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `mov al, es:[di+2Eh]`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 13 bytes，已逐條讀完） | audit/embedded-strings.md<br>audit/function-index/dos-overlay-09.md |
| `154F` | sub_154F | — | 6 | 3 | 2 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `mov cl, 4`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 6 bytes，已逐條讀完） | — |
| `1577` | sub_1577 | — | 15 | 5 | 2 | 1 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `jle short sub_159F`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 15 bytes，已逐條讀完） | — |
| `1586` | sub_1586 | — | 11 | 4 | 4 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `mov cx, 3`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 11 bytes，已逐條讀完） | — |
| `1595` | sub_1595 | — | 5 | 2 | 2 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `xor ah, ah`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 5 bytes，已逐條讀完） | — |
| `159A` | sub_159A | — | 5 | 2 | 7 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `mov [bp-2], al`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 5 bytes，已逐條讀完） | — |
| `159F` | sub_159F | — | 32 | 12 | 3 | 1 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `cmp byte ptr es:[di+2Eh], 55h`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 32 bytes，已逐條讀完） | — |
| `15D1` | sub_15D1 | — | 10 | 3 | 2 | 1 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `les di, es:[di+18Dh]`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 10 bytes，已逐條讀完） | — |
| `15DB` | sub_15DB | — | 126 | 44 | 2 | 1 |  | 邊界碎片 | — | docs/spec/587-ecl-handler-21-37-shared.md<br>邊界碎片：落在 1542h 的 prologue 區間內部，自己不是 prologue。所屬函式尚未解讀，讀它時會一併涵蓋。 | — |
| `1681` | sub_1681 | — | 1129 | 371 | 2 | 6 | ✓ | 待解讀 | — | — | audit/function-index/dos-overlay-09.md |
| `16A4` | sub_16A4 | — | 26 | 10 | 2 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `sub ax, dx`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 26 bytes，已逐條讀完） | — |
| `1847` | sub_1847 | — | 98 | 35 | 2 | 1 |  | 邊界碎片 | — | docs/spec/587-ecl-handler-21-37-shared.md<br>邊界碎片：落在 1681h 的 prologue 區間內部，自己不是 prologue。所屬函式尚未解讀，讀它時會一併涵蓋。 | — |
| `18A9` | sub_18A9 | — | 6 | 2 | 2 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `mov [bp-0Eh], dx`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 6 bytes，已逐條讀完） | — |
| `1921` | sub_1921 | — | 5 | 2 | 2 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `and al, 10h`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 5 bytes，已逐條讀完） | — |
| `1926` | sub_1926 | — | 664 | 206 | 2 | 6 |  | 邊界碎片 | — | docs/spec/587-ecl-handler-21-37-shared.md<br>邊界碎片：落在 1681h 的 prologue 區間內部，自己不是 prologue。所屬函式尚未解讀，讀它時會一併涵蓋。 | — |
| `1BC6` | sub_1BC6 | — | 7 | 5 | 0 | 0 | ✓ | 已解讀 | exact | docs/spec/752-empty-procedures.md<br>空程序：原始 bytes 就是 55 89 E5 89 EC 5D CB(push bp/mov bp,sp/mov sp,bp/pop bp/retf)，無參數無副作用。IDA 匯出漏 byte 把 89 EC 解成 in al,dx 才看起來像 I/O 讀取 | spec/752-empty-procedures.md |
