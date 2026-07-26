# 第二十三輪：game ECL block session integration

狀態：`READY`（限 block ownership／selection offset integration）

`cmd/azure-bonds-game` 現在載入 ECL1.DAX 的全部 decoded blocks，透過 `game.NewStateFromECLBlocks` 建立 `BlockSession`。每次 `State.Select` 會：

1. 將新選項附加至 global selection sequence。
2. 從目前 block initial entry 執行 interactive subset。
3. 在 sequence 用完時停在下一個 menu，或透過 `NEWECL` switch 到 session target block。

`BlockSession` 保存 current block 與 selection offset；目前 memory、call stack、party/map state 尚未完整跨 block 持久化，因此仍是 bounded integration，不是完整 VM。

## 驗收

```sh
go test ./...
go test ./cmd/azure-bonds-game
```
