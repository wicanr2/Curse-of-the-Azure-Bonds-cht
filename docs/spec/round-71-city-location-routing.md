# 第七十一輪：三城市 location routing

狀態：`READY`（限 opening city menu 與共同 map/place contract）

## 本輪成果

- 以已驗證的 ECL opening sequence `0,0,1,<city>` 將城市選單映射為：
  - `0`：`SHADOWDALE`／暗影谷
  - `1`：`ASHABENFORD`／阿沙本福德
  - `2`：`DAGGER FALLS`／匕首瀑布
- `Area.CurrentCity` 保存選擇，wilderness floor 生成改用對應的 reference `CityInfo` flags，而非固定 index 0。
- `WILDERNESS`／`EXIT` map-entry 與共用 INN／STORE／BAR／LEAVE place contract 可由三個 location 共用。
- INN／STORE／BAR 已各自顯示繁中 place-event screen，按 Enter 可返回場所選單；商店交易、客棧休息與酒館情報 routine 仍分開列為待實作。
- 保留原始英文 location 名稱與繁中顯示名稱。

## 邊界

本輪證明的是 city routing 與 map/place state contract，不宣稱三城市各自的原始 tile、ECL 場所 routine、商店／酒館／客棧內容已完成。

## 驗證

`go test ./internal/game` 會測試三個城市的 location、locale、原始名稱與 `Area.CurrentCity` mapping；完整驗證使用 `go test ./...`。
