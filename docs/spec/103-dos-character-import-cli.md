# 第一百零三輪：DOS character import CLI

狀態：`READY`（限 sidecar bundle → versioned remake party JSON）

`cmd/azure-bonds` 現在提供可重現的角色匯入命令：

```sh
go run ./cmd/azure-bonds \
  -import-character \
  -character-id hero-1 \
  -character-record SAVE/CHRDATB1.SAV \
  -character-effects SAVE/CHRDATB1.FX \
  -character-inventory SAVE/CHRDATB1.SWG \
  -out-party imported-party.json
```

`-character-record` 必須是已解壓的 `.SAV`／`.GUY`；`.FX`／`.SWG` 可省略。輸出使用目前 versioned remake party JSON，輸入檔案只讀不改。若不提供 `-out-party`，JSON 會輸出到 stdout，方便 pipe 到其他工具。

這不是 `SAVGAM?.DAT` party/area save importer；CLI 刻意只暴露已證實的三檔角色 bundle boundary。
