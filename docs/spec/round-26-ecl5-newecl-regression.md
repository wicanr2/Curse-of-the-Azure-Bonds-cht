# 第二十六輪：ECL5 real NEWECL regression

狀態：`READY`（限已定位 transition entry）

掃描全部 ECL 初始化 entries 後，第二個可重現的 real transition 位於：

```text
member: ECL5.DAX
block: 0x30 (decimal 48)
payload entry: +0x0098 (decimal 152)
opcode: 0x20 NEWECL
target: 0x50
```

驗收命令：

```sh
go run ./cmd/azure-bonds \
  -image curseoftheazurebonds.zip -member ECL5.DAX \
  -block-id 48 -run-subset -run-start 152
```

預期輸出包含 `subset steps=1` 與 `subset requested ECL block 0x50`。
此回歸只證明 ECL5 的 entry-level transition signal 與 BlockSession target contract 相容；尚未宣稱從完整玩家流程抵達，也未完成跨 block memory／call stack。

- [x] 以原始映像解碼 ECL5 block 0x30。
- [x] 從 payload `+0x0098` 執行單一 NEWECL 指令。
- [x] 驗證 target block `0x50`。
- [ ] 從完整玩家流程抵達此 entry，保存跨 block memory／call stack。
