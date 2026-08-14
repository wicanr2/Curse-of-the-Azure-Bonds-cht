# dos-overlay-11 函式明細

由 `cmd/re-ledger` 產生。位址是模組內位址：overlay 為 code-local
offset（base 0），resident executable 為 IDA linear address。

| 位址 | IDA | Borland 符號 | 大小 | 指令 | 被呼叫 | 呼叫 | entry | 狀態 | 等級 | 規格／理由 | 引用 |
|---|---|---|---:|---:|---:|---:|:-:|---|---|---|---|
| `0000` | sub_0 | — | 47 | 13 | 0 | 0 | ✓ | 已解讀 | exact | docs/spec/572-resident-service-functions.md<br>unit 初始化(LOADINIT):依序 far call 八個依賴單元的初始化 entry,最後呼叫 19Ch:2Ah 的 overlay 載入器 | audit/embedded-strings.md<br>audit/overlay-init-graph.md<br>context/50-log-2026-08-09-13.md<br>knowledge/golden-box-reverse-engineering-worklist.md<br>spec/1074-global-init-pc98.md<br>spec/520-dos-movement-to-overlay-cell-layer-bridge.md |
| `0056` | sub_56 | — | 1666 | 403 | 0 | 1 | ✓ | 邊界碎片 | — | docs/spec/569-small-function-batch-reading.md<br>邊界碎片：body 內沒有 `ret` 也沒有尾跳躍，最後一條是 `mov ds:7214h, ax`；這是 IDA 建錯的函式邊界，真正的函式體要以位址範圍重讀（body 共 1666 bytes，已逐條讀完） | audit/embedded-strings.md<br>audit/function-strings.md |
| `06D8` | sub_6D8 | — | 516 | 164 | 0 | 1 | ✓ | 已解讀 | exact | 892<br>新遊戲的全域初始化(retf，無參數)：清 DS:5000h(0A8h)/50A8h(150h)/51F8h(150h,填 0FFh)/5348h(150h) 四塊與 遠指標 4F99h(800h)/4F9Dh(800h)/4FA1h(400h)/4FA5h(1E00h)，DS:720Fh 5 bytes 清零後設 7/0Dh/2，DS:7588h 3 bytes 填 1，for i:=2 to 3 把 word[7210h+i*4] 與 word[7212h+i*4] 設 0FFFFh，之後一長串常數。★五個先前規格的預設值在這裡定案：DS:4FA9h=4(遊戲速度)、DS:5BEEh=1(磁碟編號)、DS:4FBAh=4(畫面模式)、DS:728Ah/728Bh=0FFh、4F9Dh^[67Ch]=0(人數)；DS:650Ah 與 DS:6506h 一起清成 NIL | audit/function-index/pc98-overlay-11.md<br>spec/892-new-game-globals.md |
| `08DC` | sub_8DC | — | 7 | 5 | 0 | 0 | ✓ | 已解讀 | exact | docs/spec/752-empty-procedures.md<br>空程序：原始 bytes 就是 55 89 E5 89 EC 5D CB(push bp/mov bp,sp/mov sp,bp/pop bp/retf)，無參數無副作用。IDA 匯出漏 byte 把 89 EC 解成 in al,dx 才看起來像 I/O 讀取 | spec/752-empty-procedures.md |
