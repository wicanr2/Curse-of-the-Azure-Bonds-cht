# 第一百零四輪：remake startup DOS character import

狀態：`READY`（限單一 DOS character bundle 的直接啟動 bridge）

`cmd/azure-bonds-game` 新增：

```sh
go run ./cmd/azure-bonds-game \
  -font /path/to/cjk.ttf \
  -dos-character-record SAVE/CHRDATB1.SAV \
  -dos-character-effects SAVE/CHRDATB1.FX \
  -dos-character-inventory SAVE/CHRDATB1.SWG \
  -dos-character-id hero-1
```

啟動時會以 `game.State.LoadDOSCharacterFiles` 解析 sidecar、套用原始 HP／icon／equipment／effects，並以 imported character 建立目前 remake party。`-party-load` 與 direct DOS import 互斥；若兩者都未提供，維持既有角色建立／debug encounter 行為。

這是單一角色的可玩 startup bridge，不是完整六人 `SAVGAM?.DAT` party／area loader；後者仍需原始 container bytes 與欄位證據。
