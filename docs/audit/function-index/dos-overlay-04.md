# dos-overlay-04 函式明細

由 `cmd/re-ledger` 產生。位址是模組內位址：overlay 為 code-local
offset（base 0），resident executable 為 IDA linear address。

| 位址 | IDA | Borland 符號 | 大小 | 指令 | 被呼叫 | 呼叫 | entry | 狀態 | 等級 | 規格／理由 | 引用 |
|---|---|---|---:|---:|---:|---:|:-:|---|---|---|---|
| `0000` | sub_0 | — | 71 | 16 | 0 | 0 | ✓ | 已解讀 | exact | docs/spec/751-overlay-init-chain-dependency-graph.md<br>unit 初始化段(retf，不收參數)：依序呼叫 9 個相依 unit 的 0000h — overlay-21、overlay-19、overlay-26、overlay-24、overlay-07、overlay-34、overlay-22、overlay-23、overlay-25。由原始 bytes 解出(IDA 匯出在此漏 byte) | audit/embedded-strings.md<br>audit/function-index/dos-overlay-02.md<br>audit/function-index/dos-overlay-04.md<br>audit/function-index/dos-overlay-23.md<br>audit/function-index/dos-overlay-24.md<br>audit/function-index/dos-overlay-25.md |
| `0047` | sub_47 | — | 172 | 50 | 6 | 0 | ✓ | 待解讀 | — | — | audit/embedded-strings.md<br>audit/function-index/dos-overlay-04.md<br>audit/function-index/dos-overlay-19.md<br>audit/function-index/pc98-overlay-04.md<br>spec/764-fsplit-dbcs-and-eight-slot-longint-table.md<br>spec/785-cross-platform-pairs-third-batch.md |
| `00F3` | sub_F3 | — | 397 | 143 | 7 | 1 | ✓ | 待解讀 | — | — | audit/function-index/dos-overlay-04.md<br>audit/function-index/pc98-overlay-04.md<br>spec/785-cross-platform-pairs-third-batch.md<br>spec/794-remove-curse-and-party-cycle-keys.md<br>spec/796-cure-disease-neutralize-poison.md |
| `0280` | sub_280 | — | 134 | 53 | 1 | 3 | ✓ | 已解讀 | exact | docs/spec/785-cross-platform-pairs-third-batch.md<br>Cure Blindness(retf，無參數)：<014Dh:00A7h>(@暫存, 21h, DS:6506h) 為假時顯示 CS:0263h 'is not blind.' 並用 本模組 0047h 問 Y/N；答 Y 才顯示 CS:0271h 'Cure Blindness' 並叫 本模組 00F3h(3E8h=1000)，再答 Y 就 <0141h:002Fh>(0, 0, 21h, DS:6506h) 移除效果 21h。與 spec 763 石化解除同一個模子(那支用 7D0h) | audit/function-index/pc98-overlay-04.md<br>spec/785-cross-platform-pairs-third-batch.md<br>spec/796-cure-disease-neutralize-poison.md |
| `0324` | sub_324 | — | 217 | 75 | 1 | 3 | ✓ | 已解讀 | exact | docs/spec/796-cure-disease-neutralize-poison.md<br>Cure Disease：for i := 1 to 6 查 <014Dh:00A7h>(@暫存, byte[0DDh+i], DS:6506h)（疾病碼表 DS:0DDh[1..6] = 1Fh,22h,2Bh,2Ch,32h,39h，兩平台相同），任一命中即算生病；否則顯示 名字+'is not Diseased.'(16) 並用 0047h 再確認。答 Y 顯示 'Cure Disease'(12)、00F3h(3E8h=1000) 再問；答 Y 則 DS:6F9Ch:=1、for i := 1 to 6 呼叫 <0141h:002Fh>(DS:6506h, byte[0DDh+i], longint 0)、DS:6F9Ch:=0。retf。spec 796 | spec/796-cure-disease-neutralize-poison.md |
| `043D` | sub_43D | — | 513 | 177 | 1 | 2 | ✓ | 待解讀 | — | — | — |
| `063E` | sub_63E | — | 251 | 81 | 1 | 3 | ✓ | 待解讀 | — | — | — |
| `0892` | sub_892 | — | 205 | 77 | 1 | 3 | ✓ | 已解讀 | exact | docs/spec/796-cure-disease-neutralize-poison.md<br>Neutralize Poison：只查效果碼 37h（<014Dh:00A7h>）；沒有就顯示 名字+'is not poisoned.'(16) 並用 0047h 再確認。答 Y 顯示 'Neutralize Poison'(17)、00F3h(3E8h=1000) 再問；答 Y 則 DS:6F9Ch:=1，依序 <0141h:002Fh>(DS:6506h, 37h/16h/0Fh, longint 0)，DS:6F9Ch:=0。檢查與移除不對稱（查一個移三個）。retf。spec 796 | spec/796-cure-disease-neutralize-poison.md |
| `097B` | sub_97B | — | 202 | 72 | 1 | 3 | ✓ | 已解讀 | exact | docs/spec/794-remove-curse-and-party-cycle-keys.md<br>Remove Curse：先掃 DS:6506h^[14Dh] 物品鏈（next +2Ah）找 +36h <> 0（詛咒旗標）；沒有就查效果碼 24h（<014Dh:00A7h>），兩者都沒有才顯示 名字+'is not cursed.'(14) 並用 本模組 0047h 再確認；答 Y 則顯示 'Remove Curse'(12)、本模組 00F3h(0DACh=3500) 再問一次，答 Y 則 DS:7435h := DS:6506h 後呼叫 <011Ah:004Dh>。與 spec 763/785 的神殿服務同一個模子，同樣沒有扣款動作。retf。spec 794 | spec/794-remove-curse-and-party-cycle-keys.md |
| `0A63` | sub_A63 | — | 175 | 48 | 1 | 3 | ✓ | 待解讀 | — | — | — |
| `0B12` | sub_B12 | — | 506 | 220 | 1 | 8 | ✓ | 待解讀 | — | — | — |
| `0DD6` | sub_DD6 | — | 563 | 243 | 0 | 5 | ✓ | 待解讀 | — | — | — |
| `1009` | sub_1009 | — | 22 | 8 | 1 | 2 |  | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：有 `pop bp` 收尾卻沒有 `push bp` 開頭；還原的是別人建立的 frame，屬被切開的後半段（body 共 22 bytes，已逐條讀完） | — |
| `101F` | sub_101F | — | 7 | 5 | 0 | 0 | ✓ | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>空函式：prologue／epilogue 之外沒有任何指令，呼叫即返回（body 共 7 bytes，已逐條讀完） | audit/embedded-strings.md<br>audit/string-pairs.md |
