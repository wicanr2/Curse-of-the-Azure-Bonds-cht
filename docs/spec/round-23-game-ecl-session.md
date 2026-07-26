# 第二十三輪：game ECL block session integration

狀態：`READY`（限 block ownership／selection offset integration）

`cmd/azure-bonds-game` 現在載入 ECL1.DAX 的全部 decoded blocks，透過 `game.NewStateFromECLBlocks` 建立 `BlockSession`。每次 `State.Select` 會：

1. 將新選項附加至 global selection sequence。
2. 從目前 block initial entry 執行 interactive subset。
3. 在 sequence 用完時停在下一個 menu，或透過 `NEWECL` switch 到 session target block。

`BlockSession` 保存 current block 與 selection offset，並在 `NEWECL` signal 時切換；第 54 輪已加入 bounded memory／call stack 的共享 context，第 55 輪已加入 ECL1–ECL6 global namespace。party/map state 與完整 VM semantics 仍未完成。

## 驗收

```sh
go test ./...
go test ./cmd/azure-bonds-game
```
