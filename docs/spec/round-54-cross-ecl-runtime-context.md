# 第五十四輪：跨 ECL block 的 runtime context

狀態：`READY`（限 `NEWECL` context transfer）

## 已確認行為

`NEWECL` 是 ECL VM 的控制轉移，不是重新建立空白事件。當 bounded runner 在 source block 執行 `NEWECL` 時，會保存目前的：

- numeric／string memory；
- `COMPARE` flags；
- `GOSUB`／`ON GOSUB` return stack。

`BlockSession` 切換至 target block 後，從該 block 的 initial entry 繼續，並使用同一份 runtime context。target block 可以讀取 source block 在轉移前寫入的 memory；selection offset 也沿用同一個 session。

## 驗證

- synthetic source block 以 `SAVE` 寫入 `memory[0x9000]`，再 `NEWECL 0x51`。
- target block initial entry 以 `PRINT memory[0x9000]` 讀取，結果為 `7`。
- `go test -vet=off ./...` 通過。

## 邊界與未完成項目

這不是完整 DOS VM：原始 64 KiB memory 的地址初始化、所有 opcode、外部 `PROGRAM` routine、block loader 的真實 call/return continuation，以及完整劇情仍待反組與 real-entry regression。跨 block context 目前只在已支援的 bounded command subset 內有效。
