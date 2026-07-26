# 第七十七輪：玩家戰鬥 icon state 與方向／攻擊選擇

狀態：`READY`（限 renderer state boundary，尚未完成原始 player save decoder）

## 已確認

- reference `CombatIcon.LoadIcons` 會以 normal block 與 `block + 0x80` attack block 建立兩組圖。
- reference 同時快取 normal／attack 的水平翻轉圖，`GetIcon` 在 direction `> 3` 時選 flipped variant。
- reference Player record 的 icon 欄位為 `head_icon @ 0x141`、`weapon_icon @ 0x142`、`icon_id @ 0x143`、`icon_size @ 0x144`。

## 本輪成果

- `combat.Fighter` 增加 `HasPartyIcon`、`PartyHeadBlock`、`PartyBodyBlock`、`IconDirection`、`IconAttack`，保留未來 save／角色資料 decoder 的輸入邊界。
- `State.SetParty` 對尚未有原始 icon 資料的 remake roster 使用 deterministic six-slot fallback；已有欄位不被覆蓋。
- generator 產出 normal 與 attack 的 CHEAD＋CBODY 合成 PNG。
- `gfx.Picture.FlipHorizontal` 以 indexed pixels 實作原作方向 flip；Ebiten renderer 依 direction 進行水平繪製。

## 邊界

目前尚未把 DOS player record 的 `head_icon`／`weapon_icon`／`icon_size` 從原始 save 或角色建立資料解碼進來；完整八方向戰場座標、攻擊動畫時機與 icon recolor 仍待後續反組譯。

## 驗證

```sh
go test ./...
go run ./scripts
```
