# pc98-overlay-03 函式明細

由 `cmd/re-ledger` 產生。位址是模組內位址：overlay 為 code-local
offset（base 0），resident executable 為 IDA linear address。

| 位址 | IDA | Borland 符號 | 大小 | 指令 | 被呼叫 | 呼叫 | entry | 狀態 | 等級 | 規格／理由 | 引用 |
|---|---|---|---:|---:|---:|---:|:-:|---|---|---|---|
| `0000` | sub_0 | LOADPROTECT | 12 | 6 | 0 | 0 | ✓ | 已解讀 | exact | docs/spec/569-small-function-batch-reading.md<br>轉呼叫：整個 body 只準備參數並執行 `call far ptr 19Ah:2Ah`（body 共 12 bytes，已逐條讀完） | audit/embedded-strings.md<br>audit/function-index/dos-overlay-02.md<br>audit/function-index/dos-overlay-03.md<br>audit/overlay-init-graph.md<br>spec/1062-copy-protection-code-wheel.md<br>spec/1063-copy-protection-pc98.md |
| `000C` | sub_C | GAMECLEAN | 18 | 8 | 0 | 0 | ✓ | 已解讀 | exact | docs/spec/753-small-utility-routines.md<br>永遠回 1(retf，無參數)：配置 19Ah bytes 的 local 卻一個都沒用到，直接 mov byte [bp-1],1 後回傳 | audit/embedded-strings.md<br>audit/function-index/dos-overlay-02.md<br>audit/function-index/dos-overlay-03.md<br>spec/1062-copy-protection-code-wheel.md<br>spec/1063-copy-protection-pc98.md<br>spec/753-small-utility-routines.md |
| `0131` | sub_131 | DOPROTECT | 715 | 310 | 0 | 1 | ✓ | 已解讀 | exact | 1063<br>防拷轉輪 PC-98 側(310 條)，DOS 對側 overlay-03:0011Ah(spec 1062)。★★★ 公式逐條相同：(espruar + 22h − dethek) + 路徑×0Ch + (5−方框)×2，用 while ±24h 夾到 0..35，答案 := byte[表 + 方框×25h + 索引 + 1]；Random 上界 1Ah/16h/3/6 與第二組符文索引 +1Ah 也相同。★★★ 答案表基底是 DS:00FCh(DOS 是 DS:0000h)，6×37 的內容與 DOS 逐 byte 相同 ⇒ 轉輪防拷在兩平台是同一份資料，remake 只需實作一次。★★ PC-98 看得到重試上限：cmp var_1,2 / ja ⇒ 第 3 次錯就不再重來(DOS 的上限判斷落在未反組譯的區段)；失敗三次後叫 far 0893:0000h 兩次(參數 DS:483Ch 與 DS:4844h)、DS:7F16h:=9、顯示「目に見えぬ力が、きみを〈奈落〉に放りこんだ！」(44)、Delay(3E8h)、far 07E3:0077h。★★★ 中文化參考：PC-98 把三個路徑圖樣改成全形符號(「－・・－・・－・・」「－　－　－　－　－」「‥‥‥‥‥‥‥‥‥」各 18) —— 圖樣本身被翻譯了，因為它是文字不是圖形；語序也整個重排成數字在前(「N 番のマスには何とありますか？」) | audit/function-strings.md<br>spec/1062-copy-protection-code-wheel.md<br>spec/1063-copy-protection-pc98.md |
| `03FC` | sub_3FC | — | 7 | 5 | 0 | 0 | ✓ | 已解讀 | exact | docs/spec/752-empty-procedures.md<br>空程序：原始 bytes 就是 55 89 E5 89 EC 5D CB(push bp/mov bp,sp/mov sp,bp/pop bp/retf)，無參數無副作用。IDA 匯出漏 byte 把 89 EC 解成 in al,dx 才看起來像 I/O 讀取 | spec/752-empty-procedures.md |
