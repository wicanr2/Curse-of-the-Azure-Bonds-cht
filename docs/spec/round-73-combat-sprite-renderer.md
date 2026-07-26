# 第七十三輪：Ebiten combat sprite renderer

狀態：`READY`（限 CPIC asset mapping 與目前 playable combat slice）

## 本輪成果

- `combat.Fighter` 保存 `SpriteSet`／`SpriteBlock`；ECL `LOAD MONSTER` 的 `IconBlock` 會透過 monster adapter 傳遞到戰鬥 fighter。
- Ebiten game 從 `assets/sprites/cpic*-block-*-item-00.png` 載入透明 PNG。
- 戰鬥場景現在會在 party／enemy 名稱與 HP 旁顯示 CPIC 小人；精確 block 不存在時使用排序後的 deterministic fallback，避免畫面空白。
- monster regression 驗證 `LOAD MONSTER` 的 CPIC block 沒有在資料橋中遺失。

## 邊界

這是目前可操作戰鬥 slice 的 sprite renderer，不代表已完成原版 `CHEAD/CBODY` 角色換裝、`SPRIT*.DAX` 動畫、八方向 combat placement、攻擊動畫或完整戰鬥 UI。

## 驗證

```sh
go test ./...
go run ./cmd/azure-bonds-game -font /path/to/chinese-font.ttf -encounter
```
