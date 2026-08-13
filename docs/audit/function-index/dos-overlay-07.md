# dos-overlay-07 函式明細

由 `cmd/re-ledger` 產生。位址是模組內位址：overlay 為 code-local
offset（base 0），resident executable 為 IDA linear address。

| 位址 | IDA | Borland 符號 | 大小 | 指令 | 被呼叫 | 呼叫 | entry | 狀態 | 等級 | 規格／理由 | 引用 |
|---|---|---|---:|---:|---:|---:|:-:|---|---|---|---|
| `0000` | sub_0 | — | 52 | 14 | 0 | 5 | ✓ | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>呼叫序列：依序執行 9 個呼叫，沒有其他動作：`call loc_EB3+3`、`call far ptr unk_17E7`、`call loc_14AD`、`call far ptr sub_177A`、`call loc_1049+1`、`call loc_15D0+1`（body 共 52 bytes，已逐條讀完） | audit/ecl-handler-operand-audit.md<br>audit/embedded-strings.md<br>audit/function-index/pc98-overlay-07.md<br>context/50-log-2026-08-09-13.md<br>knowledge/gold-box-ecl-interpreter.md<br>knowledge/golden-box-reverse-engineering-worklist.md |
| `0034` | sub_34 | — | 319 | 118 | 2 | 4 | ✓ | 已解讀 | exact | docs/spec/564-ecl-operand-decoding-and-arity-validation.md<br>READVAR(n)：operand 解碼器，從 ECL PC(DS:7F21h)解 n 個 operand 進三個平行陣列;索引從 1 起;佈局 [code][low](+[high] 當 code 為 1/2/3) | audit/ecl-handler-operand-audit.md<br>audit/embedded-strings.md<br>audit/function-index/dos-overlay-07.md<br>audit/function-index/pc98-overlay-07.md<br>context/50-log-2026-08-09-13.md<br>knowledge/gold-box-ecl-interpreter.md |
| `0173` | sub_173 | — | 137 | 58 | 0 | 3 | ✓ | 已解讀 | exact | docs/spec/563-ecl-memory-model-and-operand-resolution.md<br>ADDRESSVALUE(i)：陣列在 7685/76C5/7705，分支與 PC-98 相同 | spec/562-ecl2-helper-api-and-operand-audit.md |
| `01FC` | sub_1FC | — | 357 | 133 | 0 | 3 | ✓ | 待解讀 | — | — | — |
| `0380` | sub_380 | — | 214 | 92 | 0 | 2 | ✓ | 待解讀 | — | — | — |
| `0456` | sub_456 | — | 103 | 36 | 0 | 1 | ✓ | 已解讀 | strong inference | docs/spec/617-ecl2-move-and-map-model.md<br>與 pc98 overlay-07:0556h 助憶碼序列完全相同（36 條指令，且該序列在兩邊各自的模組內唯一），語意同該筆：GETMONSTERS:把輸出指標清 nil,New(1A7h)配置 423 bytes 記錄(與 0Bh 同一種),載入指定 id 的資料後處理其 +14Eh 物品鏈 ⚠ 運算元中的 DS／overlay-local 位址兩平台不同，引用位址前須各自確認 | — |
| `04BD` | sub_4BD | — | 149 | 53 | 0 | 2 | ✓ | 待解讀 | — | — | — |
| `0552` | sub_552 | — | 53 | 22 | 1 | 0 | ✓ | 已解讀 | strong inference | docs/spec/614-ecl2-gosub-push.md<br>與 pc98 overlay-07:064Ch 助憶碼序列完全相同（22 條指令，且該序列在兩邊各自的模組內唯一），語意同該筆：DRAWHEADBODY(a,b):清 DS:BDF5h、把兩個參數分別存進 DS:7F32h 與 DS:7F33h,再以相同兩個值呼叫 far 0176:0043,最後 far 0176:003E(1,3,3) ⚠ 運算元中的 DS／overlay-local 位址兩平台不同，引用位址前須各自確認 | — |
| `0591` | sub_591 | — | 343 | 117 | 0 | 5 | ✓ | 待解讀 | — | — | audit/embedded-strings.md |
| `06E8` | sub_6E8 | — | 37 | 16 | 4 | 0 | ✓ | 已解讀 | exact | docs/spec/563-ecl-memory-model-and-operand-resolution.md<br>ADDFNC：與 PC-98 逐指令相同 | spec/562-ecl2-helper-api-and-operand-audit.md |
| `070D` | sub_70D | — | 40 | 16 | 1 | 1 | ✓ | 已解讀 | exact | docs/spec/572-resident-service-functions.md<br>字元碼調整:arg <= 1Fh 時回傳 arg + 40h,否則原樣回傳(控制字元映射到 40h 之後) | spec/572-resident-service-functions.md |
| `0735` | sub_735 | — | 91 | 28 | 4 | 1 | ✓ | 已解讀 | exact | docs/spec/563-ecl-memory-model-and-operand-resolution.md<br>ECL 位址 bank 分類器：與 PC-98 相同，唯 bank3 上界為 9DFFh | audit/embedded-strings.md<br>knowledge/gold-box-ecl-interpreter.md<br>spec/563-ecl-memory-model-and-operand-resolution.md |
| `0790` | sub_790 | — | 97 | 35 | 1 | 1 | ✓ | 已解讀 | strong inference | docs/spec/615-ecl2-findguy-maxrange.md<br>與 pc98 overlay-07:085Ch 助憶碼序列完全相同（35 條指令，且該序列在兩邊各自的模組內唯一），語意同該筆：FINDGUY(ptr):沿 DS:9598h 角色鏈找出 ptr 的 0-based 序號。找不到不是回 FFh 而是回傳鏈長度——呼叫端不另外檢查會把「不存在」當成最後一個之後那一格。與 0Ah(由序號取指標)互為反向 ⚠ 運算元中的 DS／overlay-local 位址兩平台不同，引用位址前須各自確認 | — |
| `07F1` | sub_7F1 | — | 892 | 293 | 1 | 3 | ✓ | 待解讀 | — | — | — |
| `0B6D` | sub_B6D | — | 515 | 163 | 1 | 1 | ✓ | 待解讀 | — | — | audit/embedded-strings.md |
| `0D70` | sub_D70 | — | 302 | 107 | 0 | 3 | ✓ | 已解讀 | exact | docs/spec/563-ecl-memory-model-and-operand-resolution.md<br>STOREVALUE(addr,value)：C04B/C04C/C04D → DS:720F/7210/7211；⚠ IDA 函式邊界被切短至 0E98h，實際延伸到 0F34h 之後 | audit/embedded-strings.md<br>audit/function-index/dos-overlay-07.md<br>spec/547-normal-beholder-cave-presentation-state.md<br>spec/562-ecl2-helper-api-and-operand-audit.md<br>spec/563-ecl-memory-model-and-operand-resolution.md |
| `0E98` | sub_E98 | — | 160 | 53 | 2 | 2 |  | 邊界碎片 | — | docs/spec/587-ecl-handler-21-37-shared.md<br>邊界碎片：落在已解讀函式 0D70h 的 prologue 區間內部，自己不是 prologue。指令已隨該函式讀過。 | audit/function-index/dos-overlay-07.md<br>spec/563-ecl-memory-model-and-operand-resolution.md |
| `0F3A` | sub_F3A | — | 339 | 128 | 1 | 3 | ✓ | 待解讀 | — | — | — |
| `108D` | sub_108D | — | 545 | 190 | 0 | 2 | ✓ | 待解讀 | — | — | — |
| `12AE` | sub_12AE | — | 430 | 167 | 1 | 2 | ✓ | 待解讀 | — | — | — |
| `145C` | sub_145C | — | 508 | 194 | 1 | 2 | ✓ | 待解讀 | — | — | — |
| `1652` | sub_1652 | — | 77 | 31 | 2 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `push di`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 77 bytes，已逐條讀完） | — |
| `169F` | sub_169F | — | 9 | 3 | 2 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `call far ptr 0A54h:64Eh`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 9 bytes，已逐條讀完） | — |
| `16D3` | sub_16D3 | — | 173 | 68 | 2 | 2 | ✓ | 待解讀 | — | — | — |
| `1775` | sub_1775 | — | 5 | 2 | 2 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `add di, ax`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 5 bytes，已逐條讀完） | — |
| `177A` | sub_177A | — | 63 | 25 | 2 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `call far ptr 0A54h:64Eh`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 63 bytes，已逐條讀完） | audit/embedded-strings.md |
| `17C4` | sub_17C4 | — | 6 | 3 | 2 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：有 `pop bp` 收尾卻沒有 `push bp` 開頭；還原的是別人建立的 frame，屬被切開的後半段（body 共 6 bytes，已逐條讀完） | — |
| `17EA` | sub_17EA | — | 58 | 27 | 0 | 0 | ✓ | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `push di`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 58 bytes，已逐條讀完） | — |
| `1824` | sub_1824 | — | 13 | 6 | 3 | 1 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `lea di, [bp-12Ah]`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 13 bytes，已逐條讀完） | — |
| `18EE` | sub_18EE | — | 141 | 67 | 0 | 1 | ✓ | 待解讀 | — | — | — |
| `197B` | sub_197B | — | 214 | 86 | 0 | 1 | ✓ | 待解讀 | — | — | — |
| `1A51` | sub_1A51 | — | 104 | 37 | 0 | 1 | ✓ | 已解讀 | strong inference | docs/spec/593-ecl-comparison-flags.md<br>與 pc98 overlay-07:19FEh 助憶碼序列完全相同（37 條指令，且該序列在兩邊各自的模組內唯一），語意同該筆：ECL2 stub 0093h(數值比較):進來先 FillChar(DS:A88Ah,6,0)清空六個旗標,再用六次 cmp 分別設定 =、<>、<、>、<=、>= 六個結果到 DS:A88Ah..A88Fh。六個比較全是無號(jnb/jbe/ja/jb)。配上 16h~1Bh 的「旗標為 0 就跳過下一條」,ECL 的條件式完整 ⚠ 運算元中的 DS／overlay-local 位址兩平台不同，引用位址前須各自確認 | — |
| `1AB9` | sub_1AB9 | — | 134 | 50 | 0 | 2 | ✓ | 已解讀 | strong inference | docs/spec/614-ecl2-gosub-push.md<br>與 pc98 overlay-07:1A66h 助憶碼序列完全相同（50 條指令，且該序列在兩邊各自的模組內唯一），語意同該筆：SETUPGOSUBSTACK(n)(ECL2 stub 0098h):配置 6-byte 節點插在 DS:A882h 鏈頭(LIFO),+0 存呼叫當下的 ECL_PC(此時已被 READVAR 推過 operand,所以是下一條指令的位址)、+2 指向舊鏈頭;然後把 ECL_PC 設成 ADDFNC(DS:[A957h+n], DS:[A997h+n])。與 13h(RETURN)從鏈頭彈出完全配對。operand 索引是參數而非固定 1,代表別的指令也能借它做呼叫 ⚠ 運算元中的 DS／overlay-local 位址兩平台不同，引用位址前須各自確認 | — |
| `1B3F` | sub_1B3F | — | 151 | 53 | 0 | 2 | ✓ | 待解讀 | — | — | audit/embedded-strings.md<br>context/50-log-2026-08-09-13.md<br>knowledge/golden-box-reverse-engineering-worklist.md<br>project-status.md<br>spec/518-dos-start-ecl-call-address-space-audit.md<br>spec/519-dos-overlay-vector-to-cell-layer-accessor.md |
| `1BE0` | sub_1BE0 | — | 566 | 194 | 0 | 2 | ✓ | 待解讀 | — | — | — |
| `1E16` | sub_1E16 | — | 303 | 82 | 0 | 0 | ✓ | 待解讀 | — | — | — |
| `1F45` | sub_1F45 | — | 176 | 62 | 0 | 2 | ✓ | 已解讀 | strong inference | docs/spec/621-ecl2-robstuff.md<br>與 pc98 overlay-07:1F00h 助憶碼序列完全相同（62 條指令，且該序列在兩邊各自的模組內唯一），語意同該筆：ROBSTUFF(char,chance):走 char^[14Eh] 物品鏈,依物品 +5Fh(價值)調整機率——>255 減 90、>24 減 50,減完不會變負而是歸 0;每件各擲一次 d100(far 013E:004D(1,64h)),roll <= p 才移除。p=0 時永遠偷不走(roll 最小為 1)。先取 next 再移除,與 40h 同一寫法。這確認了物品節點 +5Fh(word)是價值 ⚠ 運算元中的 DS／overlay-local 位址兩平台不同，引用位址前須各自確認 | — |
| `1FF5` | sub_1FF5 | — | 322 | 139 | 0 | 2 | ✓ | 已解讀 | exact | docs/spec/591-skip-arity-crosscheck.md<br>ECL2 entry#29(stub 00B1h):跳過下一條 ECL 指令——從 bank3 讀下一個 opcode 存進 DS:A891h,依內建的 opcode→arity 對照表呼叫 READVAR(n),arity 0 則只 ECL_PC++。這張表是 arity 的第二個獨立來源:與 handler 自己的 READVAR 參數比對,42 個一致、20 個 arity 0 自洽;1Fh 有 arity 2 卻沒有 handler;34h/36h 兩邊不一致(SKIP 說 1、handler 說 2),是原版自己的不一致。這支確定了 16h~1Bh 是「旗標為 0 就跳過下一條」 | spec/591-skip-arity-crosscheck.md |
| `2137` | sub_2137 | — | 205 | 76 | 0 | 2 | ✓ | 已解讀 | strong inference | docs/spec/620-ecl2-charspeed.md<br>與 pc98 overlay-07:20C5h 助憶碼序列完全相同（76 條指令，且該序列在兩邊各自的模組內唯一），語意同該筆：CHARSPEED(out_min,out_max):沿角色鏈取 +1A6h(速度),effect 27h 命中則 x2(shl)、否則 2Ah 命中則 /2(有號 idiv)——兩者互斥,同時中只有加速生效;更新 min/max 後回寫。初值取鏈頭角色的未調整速度;隊伍為空時會對空指標解參照,原版沒有防護 ⚠ 運算元中的 DS／overlay-local 位址兩平台不同，引用位址前須各自確認 | — |
| `2252` | sub_2252 | — | 335 | 134 | 0 | 2 | ✓ | 待解讀 | — | — | — |
| `23A1` | sub_23A1 | — | 7 | 5 | 0 | 0 | ✓ | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>空函式：prologue／epilogue 之外沒有任何指令，呼叫即返回（body 共 7 bytes，已逐條讀完） | — |
