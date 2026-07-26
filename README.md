# Curse of the Azure Bonds 中文化／Remake

這是 SSI《Curse of the Azure Bonds》（青色枷的詛咒）的反組譯、繁體中文化與 remake 研究專案。目前是**可執行的初步 prototype**，不是完整重製版；GitHub 上的每輪提交都保留可測試的成果與驗證邊界。

## 目前成果

以下圖片由原始 `curseoftheazurebonds.zip`，透過專案目前的 DAX／GFX／GEO parser 離線產生，證明圖像資料管線已經接通：

![TILES.DAX 原始圖塊 gallery](docs/screenshots/tiles-gallery.png)

![GEO2 原始 16×16 wall geometry](docs/screenshots/geo-geometry.png)

![原版規則生成的 50×25 wilderness floor 局部](docs/screenshots/wilderness-floor.png)

![GEO2 wall/door 組合出的 dungeon floor slice](docs/screenshots/dungeon-floor.png)

目前已完成的垂直切片包括：

- DAX 容器／RLE、ECL bounded VM trace 與跨 ECL1–ECL6 block context。
- 繁中開場、荒野／場所狀態、角色建立、party JSON 存檔，以及可操作戰鬥 prototype。
- `TILES.DAX`／`8X8D*.DAX` indexed pictures、`WALLDEF*.DAX`、EGA16 palette 與 `GEO2–GEO6` geometry parser。
- 原版 50×25 wilderness floor 生成規則、background entry → tile index mapping，以及依 movement cost 的荒野移動。
- GEO2 wall／door fields → dungeon background composition → TILES pixel art 的可見 slice（`D` 預覽）。
- dungeon table／chair decoration 已依 GEO `terrain & 0x40` 與原版 seeded dice pass 接入。
- Ebiten 原始 tile gallery、GEO wall viewport 與依 GEO wall bytes 驗證的游標移動。

執行遊戲需要原始素材與可顯示繁中的 TTF／OTF 字型：

```sh
go test ./...
go run ./cmd/azure-bonds-game -font /path/to/chinese-font.ttf
# 例：選擇原始 GEO3 block 0x10 作為目前 map preview
go run ./cmd/azure-bonds-game -font /path/to/chinese-font.ttf -geo-set 3 -geo-block 0x10
```

遊戲內快捷鍵：`Enter` 開始、`C` 建立角色、`J` 冒險手札、`T` 圖塊預覽、`G` GEO 預覽、`D` dungeon floor 預覽、`F5/F9` 儲存／載入 party。

## 尚未完成

完整 ECL opcode／routine、副本與城鎮 floor／tile mapping、完整場所與劇情、AD&D 全規則、音效音樂，以及原版 DOS save/import 仍在反組譯與實作中。GEO catalog 已保留原始 set／block IDs，但 area pointer 尚未自動驅動完整遊戲地圖切換。

更多證據與規格請見 [`CONTEXT.md`](CONTEXT.md)、[`docs/spec/`](docs/spec/)、[`docs/manual/`](docs/manual/) 與 [`docs/history.md`](docs/history.md)。
