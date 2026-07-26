# 第十八輪：城市選擇場景

狀態：`READY`（限 ECL city-menu extraction／locale mapping）

新增 `-interactive` CLI，可用 selection sequence 重現原始 menu session。ECL1 block 80 的實際序列已確認：

```text
0 ENTER CITY
0 PRESS BUTTON OR RETURN TO CONTINUE.
1 JOURNEY ON
→ SHADOWDALE / ASHABENFORD / DAGGER FALLS
```

第三個 menu 已接入 game state 與繁中 locale：`暗影谷`、`阿沙本福德`、`匕首瀑布`。選擇 Shadowdale 後的下一個 `WILDERNESS / EXIT` map-entry menu 已在下一輪接入；場所功能與戰鬥尚未完成。

## 驗收

```sh
go test ./...
go run ./cmd/azure-bonds -image curseoftheazurebonds.zip -member ECL1.DAX -run-subset -interactive -select 0,0,1
```
