# pc98-overlay-22 函式明細

由 `cmd/re-ledger` 產生。位址是模組內位址：overlay 為 code-local
offset（base 0），resident executable 為 IDA linear address。

| 位址 | IDA | Borland 符號 | 大小 | 指令 | 被呼叫 | 呼叫 | entry | 狀態 | 等級 | 規格／理由 | 引用 |
|---|---|---|---:|---:|---:|---:|:-:|---|---|---|---|
| `0000` | sub_0 | LOADSPELLS | 32 | 10 | 0 | 3 | ✓ | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>呼叫序列：依序執行 5 個呼叫，沒有其他動作：`call sub_15A1`、`call far ptr loc_1625+2`、`call far ptr loc_147C+1`、`call far ptr loc_1696+1`、`call loc_19CA`（body 共 32 bytes，已逐條讀完） | audit/embedded-strings.md<br>audit/string-pairs.md<br>knowledge/golden-box-reverse-engineering-worklist.md<br>spec/543-normal-campaign-coverage-and-ida-map-cell-audit.md |
| `0020` | sub_20 | — | 367 | 125 | 1 | 2 | ✓ | 待解讀 | — | — | knowledge/golden-box-reverse-engineering-worklist.md<br>spec/543-normal-campaign-coverage-and-ida-map-cell-audit.md |
| `0227` | sub_227 | SPELLMENU | 565 | 233 | 0 | 2 | ✓ | 待解讀 | — | — | — |
| `0462` | sub_462 | — | 436 | 168 | 1 | 1 | ✓ | 待解讀 | — | — | audit/embedded-strings.md<br>audit/string-pairs.md |
| `0621` | sub_621 | — | 749 | 291 | 1 | 1 | ✓ | 待解讀 | — | — | — |
| `090E` | sub_90E | — | 224 | 76 | 2 | 3 | ✓ | 待解讀 | — | — | — |
| `09EE` | sub_9EE | — | 123 | 44 | 1 | 3 | ✓ | 已解讀 | exact | docs/spec/680-overlay22-clamp-and-cure1d8.md<br>清 DS:7CCEh 起每筆 4 bytes 的 far pointer 陣列(後測迴圈,i 走 0..30h 共 **49 筆**不是 48),再把 DS:7D92h 清 0;之後走 DS:9594h^[14Eh] 物品鏈(next 在 +52h),<sub_64E8>(p) 非 0 時呼叫 <sub_90E>(arg_0)——注意那支只吃 arg_0 不吃目前這個物品,作用對象另有來源 | spec/680-overlay22-clamp-and-cure1d8.md |
| `0A69` | sub_A69 | BUILDSPELLLIST | 866 | 325 | 0 | 5 | ✓ | 待解讀 | — | — | — |
| `0DCB` | sub_DCB | FIGSPELLRANGE | 170 | 67 | 0 | 2 | ✓ | 待解讀 | — | — | — |
| `0E75` | sub_E75 | — | 237 | 99 | 10 | 2 | ✓ | 待解讀 | — | — | audit/embedded-strings.md<br>audit/string-pairs.md |
| `0F62` | sub_F62 | — | 441 | 156 | 47 | 5 | ✓ | 待解讀 | — | — | — |
| `112C` | sub_112C | GETSPELLTARGETS | 283 | 103 | 0 | 1 | ✓ | 待解讀 | — | — | — |
| `1247` | sub_1247 | TARGETDIR | 333 | 117 | 1 | 1 | ✓ | 待解讀 | — | — | audit/embedded-strings.md |
| `140A` | sub_140A | — | 1 | 1 | 8 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `hlt`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 1 bytes，已逐條讀完） | — |
| `140F` | sub_140F | — | 5 | 1 | 7 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `sub byte ptr [bx+di-507Eh], 82h`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 5 bytes，已逐條讀完） | — |
| `1414` | sub_1414 | — | 5 | 2 | 4 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `adc al, 8Eh`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 5 bytes，已逐條讀完） | — |
| `1419` | sub_1419 | — | 1 | 1 | 3 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `hlt`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 1 bytes，已逐條讀完） | — |
| `1423` | sub_1423 | — | 6 | 2 | 0 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `fadd qword ptr [bp+si-7D49h]`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 6 bytes，已逐條讀完） | — |
| `1432` | sub_1432 | — | 6 | 2 | 13 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `xchg dl, [bp+si-7D9Ah]`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 6 bytes，已逐條讀完） | — |
| `143D` | sub_143D | CASTSPELL | 39 | 14 | 0 | 2 | ✓ | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `mov di, ax`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 39 bytes，已逐條讀完） | audit/function-index/pc98-overlay-22.md |
| `1510` | sub_1510 | — | 6 | 4 | 4 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `mov al, 0Ah`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 6 bytes，已逐條讀完） | — |
| `151F` | sub_151F | — | 5 | 3 | 2 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `push di`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 5 bytes，已逐條讀完） | — |
| `1524` | sub_1524 | — | 5 | 1 | 19 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `call far ptr 0A65h:62Fh`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 5 bytes，已逐條讀完） | — |
| `1529` | sub_1529 | — | 5 | 1 | 5 | 1 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `call loc_4E94+3`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 5 bytes，已逐條讀完） | — |
| `152E` | sub_152E | — | 10 | 6 | 2 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `push di`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 10 bytes，已逐條讀完） | — |
| `1538` | sub_1538 | — | 5 | 1 | 12 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `call far ptr 0A65h:62Fh`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 5 bytes，已逐條讀完） | audit/embedded-strings.md |
| `153D` | sub_153D | — | 5 | 3 | 11 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `mov al, 0Ah`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 5 bytes，已逐條讀完） | — |
| `1542` | sub_1542 | — | 9 | 4 | 6 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `call far ptr loc_1687+1`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 9 bytes，已逐條讀完） | audit/embedded-strings.md |
| `1556` | sub_1556 | — | 11 | 3 | 6 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `push word ptr [bp-8]`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 11 bytes，已逐條讀完） | — |
| `156A` | sub_156A | — | 11 | 5 | 6 | 2 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `jz short loc_15AE`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 11 bytes，已逐條讀完） | — |
| `15A1` | sub_15A1 | — | 121 | 49 | 2 | 4 |  | 邊界碎片 | — | docs/spec/587-ecl-handler-21-37-shared.md<br>邊界碎片：落在 143Dh 的 prologue 區間內部，自己不是 prologue。所屬函式尚未解讀，讀它時會一併涵蓋。 | — |
| `1622` | sub_1622 | — | 6 | 2 | 3 | 0 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `push word ptr [bp-8]`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 6 bytes，已逐條讀完） | — |
| `1679` | sub_1679 | — | 11 | 3 | 2 | 1 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `mov [bp-4], al`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 11 bytes，已逐條讀完） | — |
| `1854` | sub_1854 | — | 28 | 11 | 1 | 0 | ✓ | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `mov al, [bp+arg_6]`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 28 bytes，已逐條讀完） | — |
| `1897` | sub_1897 | — | 77 | 27 | 1 | 1 | ✓ | 已解讀 | exact | docs/spec/680-overlay22-clamp-and-cure1d8.md<br>把 arg_8^ 夾在 0..31h(49)、arg_4^ 夾在 0..18h(24),即 50 × 25。用 jle/jge(**有號**)所以負值真的會被夾到 0——與 spec 671 的 15731h 對照,那支用無號 jb 所以下限判斷是死碼,同一專案裡兩種寫法都有。參數是指標,夾限結果寫回呼叫端 | spec/680-overlay22-clamp-and-cure1d8.md |
| `18E4` | sub_18E4 | — | 208 | 74 | 1 | 3 | ✓ | 待解讀 | — | — | — |
| `19B4` | sub_19B4 | — | 855 | 358 | 5 | 4 | ✓ | 待解讀 | — | — | audit/embedded-strings.md |
| `1D0B` | sub_1D0B | — | 386 | 146 | 3 | 7 | ✓ | 待解讀 | — | — | — |
| `1E8D` | sub_1E8D | — | 181 | 75 | 2 | 3 | ✓ | 待解讀 | — | — | — |
| `1F53` | sub_1F53 | — | 39 | 18 | 0 | 1 | ✓ | 已解讀 | exact | docs/spec/628-spell-effect-wrappers.md<br><sub_1E8D>(DS:9594h^[198h], 'は祝福を受けた。')——薄包裝,除了取欄位與備妥字串外沒有其他動作 | audit/function-index/dos-overlay-22.md<br>spec/628-spell-effect-wrappers.md |
| `1F8B` | sub_1F8B | — | 43 | 19 | 0 | 2 | ✓ | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>常數字串：以固定字串「は呪いを受けた。」（unk_1F7A，長度 16）呼叫訊息 routine（body 共 43 bytes，已逐條讀完） | audit/spell-status-message-strings.md |
| `1FB6` | sub_1FB6 | — | 61 | 24 | 0 | 1 | ✓ | 已解讀 | exact | docs/spec/680-overlay22-clamp-and-cure1d8.md<br>DS:0A520h(選定目標數)為 0 就直接返回,連骰都不擲;否則 <far 013Eh:008Eh>(t, ROLLDICE(1,8), 0) 非 0 時再呼叫 <far 014Ah:054Fh>(t)。1d8 正是 AD&D 的 Cure Light Wounds——加上 4416h 的 2d8+1 與 46CAh 的 3d8+3,cure 系列三個標準級距都齊了 | spec/680-overlay22-clamp-and-cure1d8.md |
| `1FF4` | sub_1FF4 | — | 54 | 29 | 0 | 2 | ✓ | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>常數字串：以固定字串「」（unk_1FF3，長度 0）呼叫訊息 routine（body 共 54 bytes，已逐條讀完） | — |
| `203D` | sub_203D | — | 45 | 25 | 0 | 1 | ✓ | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>常數字串：以固定字串「は魔法にかかった。」（unk_202A，長度 18）呼叫訊息 routine（body 共 45 bytes，已逐條讀完） | audit/spell-status-message-strings.md |
| `2077` | sub_2077 | — | 30 | 17 | 0 | 0 | ✓ | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `mov di, offset unk_206A`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 30 bytes，已逐條讀完） | — |
| `2095` | sub_2095 | — | 15 | 8 | 3 | 1 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：有 `pop bp` 收尾卻沒有 `push bp` 開頭；還原的是別人建立的 frame，屬被切開的後半段（body 共 15 bytes，已逐條讀完） | — |
| `20BF` | sub_20BF | — | 45 | 25 | 0 | 1 | ✓ | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>常數字串：以固定字串「は冷気に対する防護を得た。」（unk_20A4，長度 26）呼叫訊息 routine（body 共 45 bytes，已逐條讀完） | audit/spell-status-message-strings.md |
| `20ED` | sub_20ED | — | 52 | 27 | 0 | 2 | ✓ | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>常數字串：以固定字串「」（unk_20EC，長度 0）呼叫訊息 routine（body 共 52 bytes，已逐條讀完） | — |
| `2147` | sub_2147 | — | 189 | 79 | 0 | 6 | ✓ | 已解讀 | exact | docs/spec/630-spell-target-array.md<br>t := DS:0A521h;免疫條件是 t^[11Ah](RACETYPE)>1 或 t^[0DEh]>1(是大於 1 不是等於某值),命中就備妥「は影響を受けなかった。」呼叫 <far 014A:0084>(1,0Ah);否則 n := (signext(DS:9594h^[198h]) shl 7) + <far 014A:00D4>(DS:0A031h,DS:0A031h),備妥「は魅了された。」呼叫 <sub_F62>,再 <far 014A:00A7>(t,0Bh,@var_4) 非 0 時 <far 013E:002A>(0Bh,t,var_4,0)。移位量是 7,與 29A3h/3804h 的 4 不同 | audit/function-index/dos-overlay-22.md<br>audit/function-index/pc98-overlay-22.md<br>spec/629-spell-pack-idiom-and-uninit.md<br>spec/630-spell-target-array.md<br>spec/632-spell-target-array-loop.md |
| `222A` | sub_222A | — | 291 | 116 | 0 | 6 | ✓ | 已解讀 | exact | docs/spec/631-strength-spell-exceptional.md<br>力量法術:先把 DS:7D93h 設 12h、DS:7D94h 設 0,再依 <far 014A:00D4>(DS:0A031h) 的回傳 1..0Bh 改其一——case 1..6 改百分比(0/1/33h/4Ch/5Bh/64h = 0/1/51/76/91/100)、case 7..0Bh 改力量值(13h..16h = 19..22)。換成十進位就是 AD&D 的特例力量 18/xx,級距切在 01/51/76/91/00。兩組互斥一次只改一個,且 case 有預設值所以沒有未初始化問題。之後 <far 013E:007F> 非 0 才印「は強くなった。」否則印「は影響を受けなかった。」 | audit/embedded-strings.md<br>audit/function-index/dos-overlay-22.md<br>audit/function-index/pc98-overlay-22.md<br>spec/631-strength-spell-exceptional.md |
| `235E` | sub_235E | — | 147 | 43 | 0 | 4 | ✓ | 待解讀 | — | — | — |
| `2404` | sub_2404 | — | 70 | 34 | 0 | 1 | ✓ | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>常數字串：以固定字串「は友好的になった。」（unk_23F1，長度 18）呼叫訊息 routine（body 共 70 bytes，已逐條讀完） | — |
| `244B` | sub_244B | — | 95 | 47 | 0 | 3 | ✓ | 待解讀 | — | — | — |
| `24C1` | sub_24C1 | — | 45 | 25 | 0 | 1 | ✓ | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>常數字串：以固定字串「は魔法の楯に守られた。」（unk_24AA，長度 22）呼叫訊息 routine（body 共 45 bytes，已逐條讀完） | audit/embedded-strings.md<br>audit/spell-status-message-strings.md |
| `24EF` | sub_24EF | — | 71 | 37 | 0 | 3 | ✓ | 待解讀 | — | — | — |
| `2547` | sub_2547 | — | 310 | 122 | 0 | 3 | ✓ | 待解讀 | — | — | — |
| `2690` | sub_2690 | — | 81 | 35 | 0 | 2 | ✓ | 已解讀 | exact | docs/spec/629-spell-pack-idiom-and-uninit.md<br>case DS:0A520h of 1:(DS:0A031h=17h ? 0FEh : 0FDh) 2:0FFh 3,4:0——**沒有 else**,落在 1..4 以外時傳給 <sub_1D0B> 的 v 是未初始化的堆疊殘值(與 spec 624 的 7D0Dh 同一類原作行為)。訊息「は金縛りにあった。」 | spec/629-spell-pack-idiom-and-uninit.md<br>spec/631-strength-spell-exceptional.md |
| `26FC` | sub_26FC | — | 45 | 25 | 0 | 1 | ✓ | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>常數字串：以固定字串「は火炎に対する防護を得た。」（unk_26E1，長度 26）呼叫訊息 routine（body 共 45 bytes，已逐條讀完） | audit/spell-status-message-strings.md |
| `2736` | sub_2736 | — | 45 | 25 | 0 | 1 | ✓ | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>常數字串：以固定字串「は沈黙した。」（unk_2729，長度 12）呼叫訊息 routine（body 共 45 bytes，已逐條讀完） | audit/spell-status-message-strings.md |
| `2776` | sub_2776 | — | 168 | 62 | 0 | 4 | ✓ | 已解讀 | exact | docs/spec/629-spell-pack-idiom-and-uninit.md<br>t := DS:0A521h;若 t^[196h]=1 就把 DS:0A521h 清成 nil 且不印訊息;否則 <far 014A:00A7>(t,37h,@var_4) 非 0 時,若 t^[1A5h]=0 才拉到 1(HP 為 0 才救回,活著的人不會加血),備妥「は魔法にかかった。」呼叫 <sub_F62>,最後 <far 013E:002A>(4Eh,t,0,1) | audit/function-index/dos-overlay-22.md<br>audit/function-index/pc98-overlay-22.md<br>spec/629-spell-pack-idiom-and-uninit.md<br>spec/630-spell-target-array.md<br>spec/632-spell-target-array-loop.md |
| `282D` | sub_282D | — | 174 | 67 | 0 | 2 | ✓ | 已解讀 | exact | docs/spec/630-spell-target-array.md<br>挑目標填進陣列:budget := DS:9594h^[1A5h](存 DS:7D93h),掃 DS:9598h 角色鏈,+11Ah(RACETYPE)=0Eh 且 budget >= p^[1A5h] 時扣掉該值、inc DS:0A520h、把 far pointer 存進 [0A51Dh + 4×索引]。索引由 1 開始,所以 DS:0A521h 就是第 1 筆——這解釋了其他法術函式都在讀的那個位址。逐個扣預算不是固定選 N 個,扣不動的整個跳過(不會部分影響),而且預算歸零後仍走完整條鏈。訊息「は魅了された。」 | audit/function-index/dos-overlay-22.md<br>spec/630-spell-target-array.md |
| `28DC` | sub_28DC | — | 66 | 33 | 0 | 2 | ✓ | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `call far ptr sub_140A`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 66 bytes，已逐條讀完） | — |
| `2933` | sub_2933 | — | 45 | 25 | 0 | 1 | ✓ | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>常數字串：以固定字串「は透明になった。」（unk_2922，長度 16）呼叫訊息 routine（body 共 45 bytes，已逐條讀完） | audit/spell-status-message-strings.md |
| `2969` | sub_2969 | — | 45 | 25 | 0 | 1 | ✓ | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>常數字串：以固定字串「こんこん」（unk_2960，長度 8）呼叫訊息 routine（body 共 45 bytes，已逐條讀完） | audit/spell-status-message-strings.md |
| `29A3` | sub_29A3 | — | 82 | 39 | 0 | 2 | ✓ | 已解讀 | exact | docs/spec/629-spell-pack-idiom-and-uninit.md<br>v := (ROLLDICE(1,4) shl 4) or <far 014A:00D4>(DS:0A031h)——高 4 bit 放數量、低 4 bit 放另一個值,是打包不是乘法;再推入 DS:0A031h、v、三個 0,備妥「は分身した。」呼叫 <sub_F62>。1d4 個分身與 AD&D Mirror Image 一致(推論,不寫進結論)。與 3804h 用同一套打包寫法 | audit/function-index/dos-overlay-22.md<br>audit/function-index/pc98-overlay-22.md<br>spec/629-spell-pack-idiom-and-uninit.md |
| `2A04` | sub_2A04 | — | 45 | 25 | 0 | 1 | ✓ | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>常數字串：以固定字串「は弱くなった。」（unk_29F5，長度 14）呼叫訊息 routine（body 共 45 bytes，已逐條讀完） | audit/spell-status-message-strings.md |
| `2A52` | sub_2A52 | — | 1120 | 433 | 0 | 9 | ✓ | 待解讀 | — | — | — |
| `2EB2` | sub_2EB2 | — | 433 | 148 | 0 | 4 | ✓ | 待解讀 | — | — | — |
| `3072` | sub_3072 | — | 416 | 137 | 0 | 3 | ✓ | 待解讀 | — | — | — |
| `3223` | sub_3223 | — | 61 | 25 | 0 | 3 | ✓ | 已解讀 | exact | docs/spec/628-spell-effect-wrappers.md<br>if <far 013E:0070>(t, 21h) <> 0 then <far 014A:00A2>(t,1,「の視力は戻った。」),t 是 DS:0A521h。與 41AFh 分支方向相反,交叉確定 013E:0070 是「目標身上有沒有這個效果」,21h 是失明 | audit/function-index/dos-overlay-22.md<br>spec/628-spell-effect-wrappers.md |
| `3273` | sub_3273 | — | 45 | 25 | 0 | 1 | ✓ | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>常數字串：以固定字串「は視力を奪われた。」（unk_3260，長度 18）呼叫訊息 routine（body 共 45 bytes，已逐條讀完） | audit/spell-status-message-strings.md |
| `32A0` | sub_32A0 | — | 165 | 52 | 1 | 3 | ✓ | 待解讀 | — | — | — |
| `3345` | sub_3345 | — | 17 | 9 | 0 | 1 | ✓ | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>轉呼叫：整個 body 只準備參數並執行 `call near ptr sub_32A0`（body 共 17 bytes，已逐條讀完） | — |
| `3369` | sub_3369 | — | 45 | 25 | 0 | 1 | ✓ | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>常數字串：以固定字串「は病いに冒された。」（unk_3356，長度 18）呼叫訊息 routine（body 共 45 bytes，已逐條讀完） | audit/spell-status-message-strings.md |
| `3396` | sub_3396 | — | 178 | 67 | 1 | 2 | ✓ | 待解讀 | — | — | — |
| `345B` | sub_345B | — | 922 | 346 | 0 | 6 | ✓ | 待解讀 | — | — | — |
| `3804` | sub_3804 | — | 73 | 35 | 0 | 2 | ✓ | 已解讀 | exact | docs/spec/628-spell-effect-wrappers.md<br>x := <far 014A:00D4>(DS:0A031h, DS:0A031h);n := (signext(DS:9594h^[198h]) shl 4) + x——把 [198h] 放高 4 bit、x 放低位打包成一個 word(不是相加),再推入 n 與三個 0、備妥 'は祈っている。' 呼叫 <sub_F62> | audit/function-index/dos-overlay-22.md<br>audit/function-index/pc98-overlay-22.md<br>spec/628-spell-effect-wrappers.md<br>spec/629-spell-pack-idiom-and-uninit.md<br>spec/631-strength-spell-exceptional.md |
| `3879` | sub_3879 | SPELL43 | 253 | 85 | 0 | 4 | ✓ | 待解讀 | — | — | — |
| `3983` | sub_3983 | — | 45 | 25 | 0 | 1 | ✓ | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>常數字串：以固定字串「は呪われた！」（unk_3976，長度 12）呼叫訊息 routine（body 共 45 bytes，已逐條讀完） | audit/spell-status-message-strings.md |
| `39C1` | sub_39C1 | — | 45 | 25 | 0 | 1 | ✓ | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>常數字串：以固定字串「は点滅している。」（unk_39B0，長度 16）呼叫訊息 routine（body 共 45 bytes，已逐條讀完） | audit/spell-status-message-strings.md |
| `39EF` | sub_39EF | — | 246 | 103 | 0 | 5 | ✓ | 待解讀 | — | — | — |
| `3AE5` | sub_3AE5 | — | 215 | 88 | 2 | 4 | ✓ | 待解讀 | — | — | — |
| `3BCB` | sub_3BCB | — | 42 | 20 | 0 | 1 | ✓ | 已解讀 | exact | docs/spec/628-spell-effect-wrappers.md<br><sub_3AE5>(2Ah, DS:9594h^[198h], 'は加速された。')。效果 id 2Ah 與 41AFh 指同一個(加速) | audit/function-index/dos-overlay-22.md<br>spec/628-spell-effect-wrappers.md |
| `3BF5` | sub_3BF5 | — | 188 | 71 | 5 | 4 | ✓ | 待解讀 | — | — | — |
| `3CB1` | sub_3CB1 | — | 669 | 266 | 4 | 7 | ✓ | 待解讀 | — | — | — |
| `3F4E` | sub_3F4E | — | 75 | 39 | 0 | 3 | ✓ | 待解讀 | — | — | — |
| `3FA8` | sub_3FA8 | — | 46 | 21 | 0 | 2 | ✓ | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>常數字串：以固定字串「は減速された。」（unk_3F99，長度 14）呼叫訊息 routine（body 共 46 bytes，已逐條讀完） | audit/spell-status-message-strings.md |
| `3FEB` | sub_3FEB | — | 433 | 158 | 0 | 3 | ✓ | 待解讀 | — | — | — |
| `41AF` | sub_41AF | — | 65 | 32 | 0 | 3 | ✓ | 已解讀 | exact | docs/spec/628-spell-effect-wrappers.md<br>if <far 013E:0070>(t, 2Ah) = 0 then 推入 DS:0A031h 與四個 0、備妥「はすばやくなった。」呼叫 <sub_F62>。沒有該效果才施放,故 2Ah 是加速 | audit/function-index/dos-overlay-22.md<br>audit/function-index/pc98-overlay-22.md<br>spec/628-spell-effect-wrappers.md |
| `41F0` | sub_41F0 | — | 64 | 26 | 0 | 1 | ✓ | 待解讀 | — | — | — |
| `423F` | sub_423F | — | 133 | 56 | 0 | 4 | ✓ | 已解讀 | exact | docs/spec/631-strength-spell-exceptional.md<br>與 222Ah 同一族的力量法術入口:if <far 013E:007F>(t,0,15h,@var_1) <> 0 則備妥「は強くなった。」呼叫 <far 014A:0084>(1,0Ah);接著 n := ROLLDICE(1,4)*0Ah + 28h(擲骰乘 10 再加 40),<far 013E:0057>(t,92h,n,var_1,1),最後 <far 013E:0098>(t,0)。⚠ 92h 是外層 013E:0057 的參數不是 ROLLDICE 的 | audit/function-index/dos-overlay-22.md |
| `42C4` | sub_42C4 | — | 66 | 37 | 0 | 2 | ✓ | 已解讀 | exact | docs/spec/628-spell-effect-wrappers.md<br>n := ROLLDICE(1,6) + 14h (1d6+20),接著 <sub_3BF5>(n,4,0,…) 與 <sub_3CB1>(0,4,14h,3)。整支不印任何訊息 | audit/function-index/dos-overlay-22.md<br>spec/628-spell-effect-wrappers.md |
| `4313` | sub_4313 | — | 45 | 25 | 0 | 1 | ✓ | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>常數字串：以固定字串「は麻痺した。」（unk_4306，長度 12）呼叫訊息 routine（body 共 45 bytes，已逐條讀完） | audit/spell-status-message-strings.md |
| `434D` | sub_434D | — | 78 | 33 | 0 | 2 | ✓ | 已解讀 | exact | docs/spec/627-spell-cure-family.md<br>n := ROLLDICE(2,4)+2 (2d4+2);if <far 013E:008E>(t, n, 0) <> 0 則 <far 014A:00A2>(t,1,「は回復した。」),t 是 DS:0A521h。⚠ 參數交錯推入:t 先推(外層的第一個參數),2/4 是內層 ROLLDICE 的。與 5B2Bh 逐指令相同,只差字串常數位址而兩處內容相同——同一段程式碼編譯了兩次 | audit/function-index/pc98-overlay-22.md<br>spec/627-spell-cure-family.md<br>spec/680-overlay22-clamp-and-cure1d8.md |
| `43AC` | sub_43AC | — | 45 | 25 | 0 | 1 | ✓ | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>常數字串：以固定字串「は透明になった。」（unk_439B，長度 16）呼叫訊息 routine（body 共 45 bytes，已逐條讀完） | audit/spell-status-message-strings.md |
| `43DA` | sub_43DA | — | 59 | 31 | 0 | 2 | ✓ | 已解讀 | exact | docs/spec/627-spell-cure-family.md<br>n := <far 013E:0052>(DS:0A031h,0,0,2,4)+2 → 2d4+2,再 <sub_F62>(8,n,空字串)。與 46CAh 除了三個常數(2/4/+2 對 3/8/+3)以外逐指令相同;訊息由 sub_F62 自己組,呼叫端給的字串長度為 0 | audit/function-index/pc98-overlay-22.md<br>spec/627-spell-cure-family.md<br>spec/680-overlay22-clamp-and-cure1d8.md |
| `4416` | sub_4416 | — | 57 | 31 | 0 | 2 | ✓ | 待解讀 | — | — | spec/627-spell-cure-family.md<br>spec/680-overlay22-clamp-and-cure1d8.md |
| `4477` | sub_4477 | — | 244 | 57 | 0 | 4 | ✓ | 待解讀 | — | — | — |
| `456C` | sub_456C | — | 81 | 34 | 0 | 2 | ✓ | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `call far ptr sub_1414`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 81 bytes，已逐條讀完） | — |
| `45F1` | sub_45F1 | — | 150 | 64 | 0 | 6 | ✓ | 待解讀 | — | — | — |
| `4687` | sub_4687 | — | 66 | 26 | 0 | 1 | ✓ | 待解讀 | — | — | — |
| `46CA` | sub_46CA | — | 59 | 31 | 0 | 2 | ✓ | 已解讀 | exact | docs/spec/627-spell-cure-family.md<br>n := <far 013E:0052>(DS:0A031h,0,0,3,8)+3 → 3d8+3(正好是 AD&D 的 Cure Critical Wounds,但對應法術名稱不寫進結論),再 <sub_F62>(8,n,空字串)。與 43DAh 只差三個常數 | audit/function-index/pc98-overlay-22.md<br>spec/627-spell-cure-family.md<br>spec/680-overlay22-clamp-and-cure1d8.md |
| `4719` | sub_4719 | — | 96 | 50 | 0 | 2 | ✓ | 待解讀 | — | — | — |
| `477A` | sub_477A | — | 54 | 29 | 0 | 2 | ✓ | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>常數字串：以固定字串「」（unk_4779，長度 0）呼叫訊息 routine（body 共 54 bytes，已逐條讀完） | — |
| `47BD` | sub_47BD | — | 203 | 38 | 0 | 2 | ✓ | 待解讀 | — | — | — |
| `48AA` | sub_48AA | — | 42 | 14 | 0 | 1 | ✓ | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `call far ptr sub_1414`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 42 bytes，已逐條讀完） | — |
| `496B` | sub_496B | — | 171 | 72 | 0 | 3 | ✓ | 已解讀 | exact | docs/spec/632-spell-target-array-loop.md<br>bank0^[1CCh] 非 0 就直接結束;否則 for i := 1 to DS:0A520h,**逐筆檢查 nil** 後 v := <far 013E:0048>(0,4,t)、x := <sub_E75>(t,88h,88h),備妥「は絡みつかれた。」呼叫 <far 013E:0089>(v,1,0,0,x,訊息)。與 4A9Ah 相反,這支會檢查 nil——而 2776h 確實會把陣列第 1 筆清成 nil | audit/function-index/dos-overlay-22.md<br>spec/632-spell-target-array-loop.md |
| `4A2D` | sub_4A2D | — | 32 | 17 | 0 | 1 | ✓ | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>常數字串：以固定字串「に光がまとわりついた。」（unk_4A16，長度 22）呼叫訊息 routine（body 共 32 bytes，已逐條讀完） | audit/spell-status-message-strings.md |
| `4A5E` | sub_4A5E | — | 45 | 25 | 0 | 1 | ✓ | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>常數字串：以固定字串「は透明になった。」（unk_4A4D，長度 16）呼叫訊息 routine（body 共 45 bytes，已逐條讀完） | audit/spell-status-message-strings.md |
| `4A9A` | sub_4A9A | — | 133 | 57 | 0 | 4 | ✓ | 已解讀 | exact | docs/spec/632-spell-target-array-loop.md<br>對整個目標陣列施放:訊息「は魅了された。」在迴圈外只印一次,迴圈 for i := 1 to DS:0A520h 內對每筆<far 014A:00A7>(陣列[i],0Bh,@var) 非 0 時接 <far 013E:002A>(0Bh,陣列[i],var,0)。數量為 0 時整段跳過。與 2147h 同一套效果接法,差別只在那支處理單一目標。不檢查陣列元素是否為 nil | audit/function-index/dos-overlay-22.md<br>audit/function-index/pc98-overlay-22.md<br>spec/632-spell-target-array-loop.md |
| `4B2C` | sub_4B2C | — | 184 | 79 | 0 | 3 | ✓ | 待解讀 | — | — | — |
| `4BF7` | sub_4BF7 | — | 398 | 152 | 0 | 6 | ✓ | 待解讀 | — | — | — |
| `4DB9` | sub_4DB9 | — | 292 | 115 | 0 | 6 | ✓ | 待解讀 | — | — | — |
| `4F5F` | sub_4F5F | — | 454 | 212 | 0 | 3 | ✓ | 待解讀 | — | — | — |
| `5147` | sub_5147 | — | 275 | 132 | 0 | 6 | ✓ | 待解讀 | — | — | — |
| `525B` | sub_525B | — | 54 | 29 | 0 | 2 | ✓ | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>常數字串：以固定字串「」（unk_525A，長度 0）呼叫訊息 routine（body 共 54 bytes，已逐條讀完） | — |
| `529E` | sub_529E | — | 45 | 25 | 0 | 1 | ✓ | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>常數字串：以固定字串「は守られた。」（unk_5291，長度 12）呼叫訊息 routine（body 共 45 bytes，已逐條讀完） | audit/spell-status-message-strings.md |
| `52E0` | sub_52E0 | — | 1312 | 508 | 0 | 9 | ✓ | 待解讀 | — | — | audit/function-index/pc98-overlay-22.md<br>audit/function-triage.md |
| `562A` | sub_562A | — | 135 | 51 | 3 | 2 |  | 邊界碎片 | — | docs/spec/587-ecl-handler-21-37-shared.md<br>邊界碎片：落在 52E0h 的 prologue 區間內部，自己不是 prologue。所屬函式尚未解讀，讀它時會一併涵蓋。 | — |
| `5888` | sub_5888 | — | 164 | 72 | 0 | 6 | ✓ | 待解讀 | — | — | — |
| `592D` | sub_592D | — | 170 | 70 | 0 | 4 | ✓ | 待解讀 | — | — | — |
| `59D8` | sub_59D8 | — | 45 | 25 | 0 | 1 | ✓ | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>常數字串：以固定字串「」（unk_59D7，長度 0）呼叫訊息 routine（body 共 45 bytes，已逐條讀完） | audit/spell-status-message-strings.md |
| `5A06` | sub_5A06 | — | 45 | 25 | 0 | 1 | ✓ | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>常數字串：以固定字串「」（unk_5A05，長度 0）呼叫訊息 routine（body 共 45 bytes，已逐條讀完） | audit/spell-status-message-strings.md |
| `5A34` | sub_5A34 | — | 45 | 25 | 0 | 1 | ✓ | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>常數字串：以固定字串「」（unk_5A33，長度 0）呼叫訊息 routine（body 共 45 bytes，已逐條讀完） | audit/spell-status-message-strings.md |
| `5A61` | sub_5A61 | — | 189 | 75 | 0 | 4 | ✓ | 待解讀 | — | — | — |
| `5B2B` | sub_5B2B | — | 78 | 33 | 0 | 2 | ✓ | 已解讀 | exact | docs/spec/627-spell-cure-family.md<br>與 434Dh 逐指令相同的另一份複本:if <far 013E:008E>(t, ROLLDICE(2,4)+2, 0) <> 0 則印「は回復した。」。字串常數在 5B1Eh(434Dh 用 4340h),內容相同 | audit/function-index/pc98-overlay-22.md<br>spec/627-spell-cure-family.md<br>spec/680-overlay22-clamp-and-cure1d8.md |
| `5B90` | sub_5B90 | — | 376 | 161 | 0 | 11 | ✓ | 待解讀 | — | — | — |
| `5D17` | sub_5D17 | — | 268 | 102 | 0 | 6 | ✓ | 待解讀 | — | — | — |
| `5E32` | sub_5E32 | — | 472 | 176 | 0 | 9 | ✓ | 待解讀 | — | — | — |
| `6019` | sub_6019 | — | 389 | 148 | 0 | 8 | ✓ | 待解讀 | — | — | — |
| `61AD` | sub_61AD | — | 277 | 105 | 0 | 7 | ✓ | 待解讀 | — | — | — |
| `62D7` | sub_62D7 | — | 229 | 102 | 0 | 10 | ✓ | 待解讀 | — | — | — |
| `63D4` | sub_63D4 | — | 276 | 108 | 0 | 7 | ✓ | 待解讀 | — | — | — |
| `64E8` | sub_64E8 | ISASCROLL | 77 | 29 | 1 | 1 | ✓ | 待解讀 | — | — | — |
| `6535` | sub_6535 | LOSESCROLLSPELL | 121 | 42 | 0 | 2 | ✓ | 待解讀 | — | — | — |
| `65C9` | sub_65C9 | CHARSPELLMSG | 290 | 131 | 1 | 7 | ✓ | 待解讀 | — | — | — |
| `66EB` | sub_66EB | INITSPELLS | 1434 | 443 | 0 | 0 | ✓ | 待解讀 | — | — | audit/function-triage.md |
| `6C85` | sub_6C85 | — | 7 | 5 | 0 | 0 | ✓ | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>空函式：prologue／epilogue 之外沒有任何指令，呼叫即返回（body 共 7 bytes，已逐條讀完） | — |
