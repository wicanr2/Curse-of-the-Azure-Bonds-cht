# 第二十五輪：real NEWECL regression

狀態：`READY`（限已定位 transition entry）

以 `-all-entries -graph` 掃描 ECL1–ECL6 後，定位到第一條可重現 real transition：

```text
member: ECL4.DAX
block: 0x25 (decimal 37)
payload entry: +0x022B (decimal 555)
opcode: 0x20 NEWECL
target: 0x50
```

驗收命令：

```sh
go run ./cmd/azure-bonds \
  -image curseoftheazurebonds.zip -member ECL4.DAX \
  -block-id 37 -run-subset -run-start 555
```

預期輸出包含 `subset steps=1` 與 `subset requested ECL block 0x50`。這證明 real ECL transition signal 與 BlockSession target contract 相容；尚未宣稱從完整玩家流程抵達此 entry，也尚未完成跨 block memory／call stack。
