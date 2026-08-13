# pc98-overlay-07 函式明細

由 `cmd/re-ledger` 產生。位址是模組內位址：overlay 為 code-local
offset（base 0），resident executable 為 IDA linear address。

| 位址 | IDA | Borland 符號 | 大小 | 指令 | 被呼叫 | 呼叫 | entry | 狀態 | 等級 | 規格／理由 | 引用 |
|---|---|---|---:|---:|---:|---:|:-:|---|---|---|---|
| `0000` | sub_0 | LOADECL2 | 52 | 14 | 0 | 8 | ✓ | 已解讀 | strong inference | docs/spec/569-small-function-batch-reading.md<br>與 dos overlay-07:0000h 助憶碼序列完全相同（14 條指令，且該序列在兩邊各自的模組內唯一），語意同該筆：呼叫序列：依序執行 9 個呼叫，沒有其他動作：`call loc_EB3+3`、`call far ptr unk_17E7`、`call loc_14AD`、`call far ptr sub_177A`、`call loc_1049+1`、`call loc_15D0+1`（body 共 52 bytes，已逐條讀完） ⚠ 運算元中的 DS／overlay-local 位址兩平台不同，引用位址前須各自確認 | audit/ecl-handler-operand-audit.md<br>audit/embedded-strings.md<br>audit/function-index/pc98-overlay-07.md<br>context/50-log-2026-08-09-13.md<br>knowledge/gold-box-ecl-interpreter.md<br>knowledge/golden-box-reverse-engineering-worklist.md |
| `008E` | sub_8E | READVAR | 520 | 206 | 2 | 4 | ✓ | 已解讀 | exact | docs/spec/564-ecl-operand-decoding-and-arity-validation.md<br>READVAR(n)：operand 解碼器，從 ECL PC(DS:7F21h)解 n 個 operand 進三個平行陣列;索引從 1 起;佈局 [code][low](+[high] 當 code 為 1/2/3) | audit/embedded-strings.md<br>audit/function-index/pc98-overlay-07.md<br>project-status.md<br>spec/562-ecl2-helper-api-and-operand-audit.md<br>spec/564-ecl-operand-decoding-and-arity-validation.md<br>spec/README.md |
| `0296` | sub_296 | ADDRESSVALUE | 129 | 57 | 0 | 3 | ✓ | 已解讀 | exact | docs/spec/563-ecl-memory-model-and-operand-resolution.md<br>ADDRESSVALUE(i)：解第 i 個 operand；三個 64 byte 平行陣列 A917/A957/A997；code 00h/01h/03h/80h/02h/81h 分支 | spec/562-ecl2-helper-api-and-operand-audit.md |
| `0317` | sub_317 | INITECL | 362 | 134 | 0 | 3 | ✓ | 待解讀 | — | — | audit/embedded-strings.md |
| `0499` | sub_499 | GETECL | 189 | 84 | 0 | 2 | ✓ | 待解讀 | — | — | — |
| `0556` | sub_556 | GETMONSTERS | 103 | 36 | 0 | 1 | ✓ | 待解讀 | — | — | knowledge/golden-box-reverse-engineering-worklist.md<br>spec/519-dos-overlay-vector-to-cell-layer-accessor.md<br>spec/543-normal-campaign-coverage-and-ida-map-cell-audit.md |
| `05BD` | sub_5BD | MAXRANGE | 143 | 52 | 0 | 2 | ✓ | 已解讀 | exact | docs/spec/615-ecl2-findguy-maxrange.md<br>MAXRANGE(dir,x,y):bank0^[1CCh]=0(非地城)時直接回 2 並把 bank1^[582h] 也設為 2;否則以 far 017C:0034 逐格檢查可通行,最多 2 格(上限寫死)。方向編碼是 0/2/4/6 間隔 2——0 與 6 動 y、2 與 4 動 x | spec/615-ecl2-findguy-maxrange.md |
| `064C` | sub_64C | DRAWHEADBODY | 53 | 22 | 1 | 2 | ✓ | 已解讀 | exact | docs/spec/614-ecl2-gosub-push.md<br>DRAWHEADBODY(a,b):清 DS:BDF5h、把兩個參數分別存進 DS:7F32h 與 DS:7F33h,再以相同兩個值呼叫 far 0176:0043,最後 far 0176:003E(1,3,3) | audit/function-index/dos-overlay-07.md |
| `068B` | sub_68B | GODRAWWINDOW | 337 | 115 | 0 | 6 | ✓ | 待解讀 | — | — | — |
| `07DC` | sub_7DC | ADDFNC | 37 | 16 | 4 | 0 | ✓ | 已解讀 | exact | docs/spec/563-ecl-memory-model-and-operand-resolution.md<br>ADDFNC：byte pair 併成 word，回傳 (b<<8)+a；不是算術加法 | spec/562-ecl2-helper-api-and-operand-audit.md |
| `0801` | sub_801 | WHICHAREA | 91 | 28 | 4 | 1 | ✓ | 已解讀 | exact | docs/spec/563-ecl-memory-model-and-operand-resolution.md<br>ECL 位址 bank 分類器：0=4B00-4EFF 1=7C00-7FFF 2=7A00-7BFF 3=8000-9E40 4=其餘 | knowledge/gold-box-ecl-interpreter.md<br>spec/563-ecl-memory-model-and-operand-resolution.md |
| `085C` | sub_85C | FINDGUY | 97 | 35 | 1 | 1 | ✓ | 已解讀 | exact | docs/spec/615-ecl2-findguy-maxrange.md<br>FINDGUY(ptr):沿 DS:9598h 角色鏈找出 ptr 的 0-based 序號。找不到不是回 FFh 而是回傳鏈長度——呼叫端不另外檢查會把「不存在」當成最後一個之後那一格。與 0Ah(由序號取指標)互為反向 | spec/615-ecl2-findguy-maxrange.md |
| `08BD` | sub_8BD | CHECKSPECIALS | 875 | 292 | 1 | 3 | ✓ | 待解讀 | — | — | audit/embedded-strings.md<br>knowledge/gold-box-ecl-interpreter.md<br>spec/565-ecl-memory-read-path-and-asymmetry.md<br>spec/567-ecl-packed-text-and-bank1-field-map.md |
| `0C28` | sub_C28 | STORESPECIALS | 515 | 163 | 1 | 2 | ✓ | 待解讀 | — | — | audit/function-triage.md |
| `0E2B` | sub_E2B | STOREVALUE | 465 | 162 | 0 | 3 | ✓ | 已解讀 | exact | docs/spec/563-ecl-memory-model-and-operand-resolution.md<br>STOREVALUE(addr,value)：依 bank 路由寫入；bank3 為 byte；bank4 具名特例含 C04B/C04C/C04D 與 C04D 的 0/2/4/6 正規化 | spec/562-ecl2-helper-api-and-operand-audit.md |
| `0FFC` | sub_FFC | GETVALUE | 30 | 12 | 1 | 2 | ✓ | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `mov ax, [bp+arg_2]`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 30 bytes，已逐條讀完） | audit/function-index/pc98-overlay-07.md<br>spec/565-ecl-memory-read-path-and-asymmetry.md |
| `101A` | sub_101A | — | 302 | 115 | 2 | 2 |  | 邊界碎片 | — | docs/spec/587-ecl-handler-21-37-shared.md<br>邊界碎片：落在 0FFCh 的 prologue 區間內部，自己不是 prologue。所屬函式尚未解讀，讀它時會一併涵蓋。 | — |
| `1148` | sub_1148 | STORESTRING | 540 | 189 | 0 | 2 | ✓ | 待解讀 | — | — | — |
| `1364` | sub_1364 | FINDSTR | 147 | 57 | 1 | 1 | ✓ | 待解讀 | — | — | spec/567-ecl-packed-text-and-bank1-field-map.md |
| `13F7` | sub_13F7 | GETSTR | 480 | 186 | 1 | 2 | ✓ | 待解讀 | — | — | audit/function-index/pc98-overlay-07.md |
| `142D` | sub_142D | — | 83 | 30 | 2 | 2 |  | 邊界碎片 | — | docs/spec/587-ecl-handler-21-37-shared.md<br>邊界碎片：落在 13F7h 的 prologue 區間內部，自己不是 prologue。所屬函式尚未解讀，讀它時會一併涵蓋。 | context/50-log-2026-08-09-13.md<br>project-status.md |
| `147D` | sub_147D | — | 31 | 13 | 2 | 1 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `jmp short loc_1423`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 31 bytes，已逐條讀完） | — |
| `1669` | sub_1669 | — | 23 | 8 | 1 | 0 | ✓ | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `mov [bp+var_2], al`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 23 bytes，已逐條讀完） | audit/function-index/pc98-overlay-07.md |
| `1697` | sub_1697 | — | 174 | 70 | 2 | 2 |  | 邊界碎片 | — | docs/spec/587-ecl-handler-21-37-shared.md<br>邊界碎片：落在 1669h 的 prologue 區間內部，自己不是 prologue。所屬函式尚未解讀，讀它時會一併涵蓋。 | — |
| `1745` | sub_1745 | — | 5 | 3 | 2 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `push ax`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 5 bytes，已逐條讀完） | — |
| `174A` | sub_174A | — | 30 | 11 | 2 | 1 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：有 `pop bp` 收尾卻沒有 `push bp` 開頭；還原的是別人建立的 frame，屬被切開的後半段（body 共 30 bytes，已逐條讀完） | — |
| `1780` | sub_1780 | ECLMENUH | 20 | 10 | 0 | 0 | ✓ | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `mov ax, 28h`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 20 bytes，已逐條讀完） | audit/function-index/pc98-overlay-07.md |
| `1794` | sub_1794 | — | 10 | 4 | 2 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `push es`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 10 bytes，已逐條讀完） | — |
| `179E` | sub_179E | — | 5 | 3 | 2 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `push ss`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 5 bytes，已逐條讀完） | — |
| `17A3` | sub_17A3 | — | 5 | 3 | 2 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `push ax`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 5 bytes，已逐條讀完） | — |
| `17A8` | sub_17A8 | — | 72 | 31 | 2 | 2 |  | 邊界碎片 | — | docs/spec/587-ecl-handler-21-37-shared.md<br>邊界碎片：落在 1780h 的 prologue 區間內部，自己不是 prologue。所屬函式尚未解讀，讀它時會一併涵蓋。 | — |
| `17F4` | sub_17F4 | — | 11 | 6 | 3 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `mov al, 0`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 11 bytes，已逐條讀完） | audit/embedded-strings.md |
| `1808` | sub_1808 | — | 130 | 43 | 2 | 5 |  | 邊界碎片 | — | docs/spec/587-ecl-handler-21-37-shared.md<br>邊界碎片：落在 1780h 的 prologue 區間內部，自己不是 prologue。所屬函式尚未解讀，讀它時會一併涵蓋。 | — |
| `188A` | sub_188A | ECLMENUV | 158 | 75 | 0 | 1 | ✓ | 待解讀 | — | — | — |
| `1928` | sub_1928 | CHECKSTRING | 81 | 36 | 0 | 0 | ✓ | 已解讀 | exact | docs/spec/593-ecl-comparison-flags.md<br>ECL2 stub 008Eh(字串比較):把兩個字串各複製一份到區域緩衝區,FillChar(DS:A88Ah,6,0),然後呼叫六次 RTL 字串比較(0A65:0734),用與數值比較完全相同的 jnz/jz/jnb/jbe/ja/jb 分別設定六個旗標。所以 16h~1Bh 對字串與數值的語意一致 | audit/function-index/pc98-overlay-07.md<br>spec/593-ecl-comparison-flags.md |
| `1979` | sub_1979 | — | 133 | 50 | 2 | 1 |  | 邊界碎片 | — | docs/spec/587-ecl-handler-21-37-shared.md<br>邊界碎片：落在已解讀函式 1928h 的 prologue 區間內部，自己不是 prologue。指令已隨該函式讀過。 | — |
| `19FE` | sub_19FE | CHECKSTATUS | 104 | 37 | 0 | 1 | ✓ | 已解讀 | exact | docs/spec/593-ecl-comparison-flags.md<br>ECL2 stub 0093h(數值比較):進來先 FillChar(DS:A88Ah,6,0)清空六個旗標,再用六次 cmp 分別設定 =、<>、<、>、<=、>= 六個結果到 DS:A88Ah..A88Fh。六個比較全是無號(jnb/jbe/ja/jb)。配上 16h~1Bh 的「旗標為 0 就跳過下一條」,ECL 的條件式完整 | audit/function-index/dos-overlay-07.md<br>spec/593-ecl-comparison-flags.md |
| `1A66` | sub_1A66 | SETUPGOSUBSTACK | 134 | 50 | 0 | 2 | ✓ | 已解讀 | exact | docs/spec/614-ecl2-gosub-push.md<br>SETUPGOSUBSTACK(n)(ECL2 stub 0098h):配置 6-byte 節點插在 DS:A882h 鏈頭(LIFO),+0 存呼叫當下的 ECL_PC(此時已被 READVAR 推過 operand,所以是下一條指令的位址)、+2 指向舊鏈頭;然後把 ECL_PC 設成 ADDFNC(DS:[A957h+n], DS:[A997h+n])。與 13h(RETURN)從鏈頭彈出完全配對。operand 索引是參數而非固定 1,代表別的指令也能借它做呼叫 | audit/function-index/dos-overlay-07.md<br>spec/614-ecl2-gosub-push.md |
| `1AEC` | sub_1AEC | MOVEFORWARD | 145 | 52 | 0 | 2 | ✓ | 待解讀 | — | — | — |
| `1B89` | sub_1B89 | GODUEL | 584 | 201 | 0 | 2 | ✓ | 待解讀 | — | — | — |
| `1DD1` | sub_1DD1 | ROBDOUGH | 303 | 82 | 0 | 0 | ✓ | 待解讀 | — | — | — |
| `1F00` | sub_1F00 | ROBSTUFF | 176 | 62 | 0 | 3 | ✓ | 待解讀 | — | — | spec/520-dos-movement-to-overlay-cell-layer-bridge.md |
| `1FB0` | sub_1FB0 | NONEXT | 277 | 138 | 0 | 2 | ✓ | 已解讀 | exact | docs/spec/591-skip-arity-crosscheck.md<br>ECL2 entry#29(stub 00B1h):跳過下一條 ECL 指令——從 bank3 讀下一個 opcode 存進 DS:A891h,依內建的 opcode→arity 對照表呼叫 READVAR(n),arity 0 則只 ECL_PC++。這張表是 arity 的第二個獨立來源:與 handler 自己的 READVAR 參數比對,42 個一致、20 個 arity 0 自洽;1Fh 有 arity 2 卻沒有 handler;34h/36h 兩邊不一致(SKIP 說 1、handler 說 2),是原版自己的不一致。這支確定了 16h~1Bh 是「旗標為 0 就跳過下一條」 | spec/591-skip-arity-crosscheck.md |
| `20C5` | sub_20C5 | CHARSPEED | 205 | 76 | 0 | 2 | ✓ | 待解讀 | — | — | — |
| `2235` | sub_2235 | KILLTHEDUDE | 516 | 208 | 0 | 2 | ✓ | 待解讀 | — | — | — |
| `2439` | sub_2439 | — | 7 | 5 | 0 | 0 | ✓ | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>空函式：prologue／epilogue 之外沒有任何指令，呼叫即返回（body 共 7 bytes，已逐條讀完） | — |
