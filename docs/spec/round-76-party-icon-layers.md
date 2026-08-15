# 第七十六輪：CHEAD／CBODY 玩家戰鬥小人圖層

狀態：`READY`（限目前 party icon 的合成與素材管線）

## 已確認

- 參考 `LoadPlayerCombatIcon` 以 `CHEAD` 載入頭部、以 `CBODY` 載入身體／武器，再呼叫 `MergeIcon`。
- 本作原始映像的 `CHEAD.DAX` block 0–0x0D 是 24×10 的 masked picture；`CBODY.DAX` block 0–0x1F 是 24×24 的 masked picture。
- 原始合成是左上對齊：透明 index `16` 顯示另一層；兩層都不透明時使用 bitwise OR。
- renderer-neutral `gfx.MergePictures` 已以相同規則合成小圖，並以測試覆蓋透明與重疊像素。

## 本輪成果

- `scripts/render_previews/` 會從本地原始 ZIP 抽出 CHEAD／CBODY，並產生六張 `assets/sprites/party-block-XX.png` 合成圖。
- Ebiten party fighter 優先顯示這些合成小人；敵人仍依 CPIC／SPRIT block 顯示。
- 原始 DAX 不放入 repository；可重新產生流程、manifest 與 PNG 證據放在 repository。

## 可重用知識

這套 Gold Box 圖像管線可抽成共用規則：DAX block → indexed picture → mask index 16 → layer composition → RGBA／Ebiten。後續遊戲只需提供 DAX member 名稱與 icon block mapping。

## 邊界

目前 party slot 使用 0–5 的 deterministic composite，尚未解碼玩家存檔中的 `head_icon`／`weapon_icon`／`icon_size` 全部來源，也尚未接入攻擊 icon、方向 flip、SPRIT position offset 與八方向戰場座標。

## 驗證

```sh
go test ./...
go run ./scripts
```
