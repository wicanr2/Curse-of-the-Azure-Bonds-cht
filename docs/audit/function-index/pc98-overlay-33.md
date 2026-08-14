# pc98-overlay-33 函式明細

由 `cmd/re-ledger` 產生。位址是模組內位址：overlay 為 code-local
offset（base 0），resident executable 為 IDA linear address。

| 位址 | IDA | Borland 符號 | 大小 | 指令 | 被呼叫 | 呼叫 | entry | 狀態 | 等級 | 規格／理由 | 引用 |
|---|---|---|---:|---:|---:|---:|:-:|---|---|---|---|
| `0000` | sub_0 | LOADSQRPAK24 | 12 | 6 | 0 | 0 | ✓ | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>轉呼叫：整個 body 只準備參數並執行 `call far ptr 19Ah:2Ah`（body 共 12 bytes，已逐條讀完） | audit/embedded-strings.md<br>audit/function-index/dos-overlay-16.md<br>audit/function-index/dos-overlay-17.md<br>audit/overlay-init-graph.md<br>spec/756-map-fill-confirms-grid-and-assorted-routines.md |
| `004C` | sub_4C | LOAD24X24SET | 539 | 206 | 0 | 1 | ✓ | 待解讀 | — | — | audit/function-strings.md |
| `0267` | sub_267 | PUT24X24SYMBOL | 97 | 39 | 0 | 1 | ✓ | 已解讀 | exact | docs/spec/759-combatant-xy-lookup-and-bank-select-bit.md<br>用 bit 7 選資源(retf 8)：02A8h:0712h、DS:9668h/DS:9664h、DS:7F56h，其餘同 DOS overlay-33:00F9h | spec/759-combatant-xy-lookup-and-bank-select-bit.md |
| `02E8` | sub_2E8 | DISPOSEFIGURE | 116 | 48 | 0 | 1 | ✓ | 已解讀 | exact | docs/spec/816-camp-menu-and-pointer-pair-free.md<br>同 DOS，但兩處不同：開頭多一道守衛 <0A65h:08E4h>(索引, @loc_2C8) 回 0 就整支跳過；兩個欄位的相對位置相反——第一個在 DS:9670h、第二個在 DS:966Ch（DOS 是 +0 與 +4）。釋放常式是 <02A8h:10D5h>。spec 816 | — |
| `038F` | sub_38F | LOADFIGURE | 1050 | 451 | 0 | 1 | ✓ | 待解讀 | — | — | audit/function-strings.md |
| `07A9` | sub_7A9 | PUTFIGURE | 285 | 114 | 0 | 1 | ✓ | 已解讀 | exact | 880<br>同 DOS overlay-33:00508h，但★第二次 Move 的長度改用 p^[0] × p^[2](寬高相乘)，DOS 兩次都用 p^[11h]——兩台機器圖形平面配置不同(PC-98 planar / DOS packed)，第二塊大小關係不一樣。其餘 14 條是位址 | spec/880-sprite-copy-and-blit.md |
| `08C6` | sub_8C6 | — | 7 | 5 | 0 | 0 | ✓ | 已解讀 | exact | docs/spec/752-empty-procedures.md<br>空程序：原始 bytes 就是 55 89 E5 89 EC 5D CB(push bp/mov bp,sp/mov sp,bp/pop bp/retf)，無參數無副作用。IDA 匯出漏 byte 把 89 EC 解成 in al,dx 才看起來像 I/O 讀取 | spec/752-empty-procedures.md |
