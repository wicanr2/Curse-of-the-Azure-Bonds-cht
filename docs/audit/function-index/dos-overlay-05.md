# dos-overlay-05 函式明細

由 `cmd/re-ledger` 產生。位址是模組內位址：overlay 為 code-local
offset（base 0），resident executable 為 IDA linear address。

| 位址 | IDA | Borland 符號 | 大小 | 指令 | 被呼叫 | 呼叫 | entry | 狀態 | 等級 | 規格／理由 | 引用 |
|---|---|---|---:|---:|---:|---:|:-:|---|---|---|---|
| `0000` | sub_0 | — | 57 | 15 | 0 | 4 | ✓ | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>呼叫序列：依序執行 10 個呼叫，沒有其他動作：`call far ptr unk_104A`、`call loc_770`、`call far ptr sub_1184`、`call far ptr loc_16BC+1`、`call loc_69D+2`、`call far ptr loc_15CF+2`（body 共 57 bytes，已逐條讀完） | audit/embedded-strings.md<br>audit/function-index/dos-overlay-02.md<br>audit/function-index/dos-overlay-05.md<br>audit/function-index/pc98-overlay-02.md<br>audit/function-index/pc98-overlay-05.md<br>audit/overlay-init-graph.md |
| `0039` | sub_39 | — | 792 | 278 | 1 | 2 | ✓ | 待解讀 | — | — | audit/embedded-strings.md<br>audit/function-index/pc98-overlay-02.md |
| `0351` | sub_351 | — | 491 | 166 | 1 | 1 | ✓ | 待解讀 | — | — | — |
| `053C` | sub_53C | — | 1128 | 343 | 1 | 5 | ✓ | 待解讀 | — | — | — |
| `0A7E` | sub_A7E | — | 603 | 250 | 1 | 2 | ✓ | 待解讀 | — | — | — |
| `0CE6` | sub_CE6 | — | 203 | 90 | 1 | 3 | ✓ | 已解讀 | exact | docs/spec/817-opportunity-attack-and-take-menu.md<br>Take 清單：retf 10h = 四個遠指標，宣告順序 (參數C, 參數8, 參數4, 參數0)。旗標 := 1；<far 01A0h:0000h>（DOS 才有）；掃 DS:6F8Ch 物品鏈（next +2Ah）對每個節點呼叫 <far 14F3h+2>(0,0,0,0,節點,longint 0)；輸出 := <far 16A8h+1>('Items: '(7，結尾空白), 'Take'(4), 0Dh, 0Ah, 0Fh, 1, 1, 26h, 16h, DS:6F8Ch, 1, @旗標, 參數0, 參數4)；參數C^ := al、參數8^ := 參數4^。spec 817 | audit/function-index/dos-overlay-05.md<br>audit/function-index/pc98-overlay-05.md<br>spec/817-opportunity-attack-and-take-menu.md<br>spec/818-take-loop-and-adjacent-lookup.md |
| `0DB1` | sub_DB1 | — | 270 | 101 | 2 | 4 | ✓ | 已解讀 | exact | docs/spec/818-take-loop-and-adjacent-lookup.md<br>Take 主迴圈：retf。repeat：選中 := NIL、清單 := DS:6F8Ch、旗標 := 0；本模組 0CE6h(@鍵, @選中, @清單, @旗標)（spec 817）；鍵 <> 54h('T') 且 <> 0Dh(Enter) 就結束；否則 <far 069Ah>(@失敗, 選中)，失敗 = 0 時把 選中 從 DS:6F8Ch 鏈（next +2Ah）摘掉（頭節點直接改 DS:6F8Ch，否則 while q^[2Ah] <> 選中 走鏈，無 NIL 守衛）、選中^[2Ah] := NIL、FreeMem(選中, 3Fh)（63 bytes，與 spec 789 的 GetMem 相同；前面的 <> NIL 判斷到不了）；DS:6F8Ch 變 NIL 就自動結束。迴圈後 <far 15A9h>。spec 818 | spec/775-take-menu-and-palette-writer.md<br>spec/818-take-loop-and-adjacent-lookup.md |
| `0ED7` | sub_ED7 | — | 79 | 31 | 1 | 4 | ✓ | 已解讀 | exact | docs/spec/775-take-menu-and-palette-writer.md<br>戰利品的 Take: 選單(retf 8，兩個遠指標都指向計數 byte)：任一為 0 就走各自的離開路徑；否則迴圈組出 CS:0EBFh 'Take: '(6 字元含尾隨空白) 與 CS:0EC6h 'Money Items Exit'(單一 16 字元字串)，依按鍵 'M'/'I'/'E' 或 #0 分派，'G'(47h)/'O'(4Fh) 走同一支；每輪結尾更新後，兩個計數任一歸零就結束。⚠ 熱鍵是寫死的 ASCII 大寫字母，選項字面換成中文後會與熱鍵脫節。補洞前這支匯出位移不可靠(spec 756 曾列為未判讀) | audit/function-index/pc98-overlay-05.md<br>spec/756-map-fill-confirms-grid-and-assorted-routines.md<br>spec/775-take-menu-and-palette-writer.md<br>spec/778-map-four-planes-and-pc98-take-menu.md |
| `106C` | sub_106C | — | 250 | 86 | 1 | 4 | ✓ | 待解讀 | — | — | — |
| `1166` | sub_1166 | — | 30 | 12 | 3 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `mov ax, 0FFh`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 30 bytes，已逐條讀完） | — |
| `1184` | sub_1184 | — | 87 | 39 | 2 | 1 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `lea di, [bp-106h]`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 87 bytes，已逐條讀完） | — |
| `1370` | sub_1370 | — | 246 | 80 | 1 | 2 | ✓ | 已解讀 | exact | docs/spec/801-take-from-pool-and-combat-state-freed.md<br>戰鬥結束的收尾：word(DS:4F9Dh^[590h]) := 0；掃 DS:650Ah 鏈——p^[18Dh]^[13h] = 1 或 p^[197h] = 1（陣營 1）者：DS:47ECh := 1、p^[196h] <> 1 時計數 +1、先抓 後 := p^[189h] 再把 DS:6506h := p 並呼叫 <far 0F0Dh+2>(1, ord(p^[18Dh]^[13h] = 1))；其餘成員：FreeMem(p^[18Dh], 16h) 並把 +18Dh 清成 NIL——**戰鬥狀態記錄是 22 bytes**。最後 DS:6506h := DS:650Ah。retf。spec 801 | spec/801-take-from-pool-and-combat-state-freed.md |
| `14AC` | sub_14AC | — | 242 | 87 | 1 | 1 | ✓ | 待解讀 | — | — | audit/function-index/dos-overlay-05.md |
| `15A9` | sub_15A9 | — | 112 | 45 | 5 | 2 |  | 邊界碎片 | — | docs/spec/587-ecl-handler-21-37-shared.md<br>邊界碎片：落在 14ACh 的 prologue 區間內部，自己不是 prologue。所屬函式尚未解讀，讀它時會一併涵蓋。 | audit/function-index/dos-overlay-05.md<br>audit/function-index/pc98-overlay-05.md<br>spec/775-take-menu-and-palette-writer.md<br>spec/818-take-loop-and-adjacent-lookup.md |
| `169F` | sub_169F | — | 9 | 3 | 4 | 1 |  | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>尾呼叫：最後一條是 `jmp loc_15B3`，控制權轉交後不返回；先設定 `mov [bp-4], ax`、`mov [bp-2], dx`（body 共 9 bytes，已逐條讀完） | — |
| `1736` | sub_1736 | — | 451 | 161 | 0 | 6 | ✓ | 待解讀 | — | — | — |
| `18F9` | sub_18F9 | — | 7 | 5 | 0 | 0 | ✓ | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>空函式：prologue／epilogue 之外沒有任何指令，呼叫即返回（body 共 7 bytes，已逐條讀完） | — |
