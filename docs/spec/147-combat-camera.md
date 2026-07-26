# 第一百四十七輪：combat camera

## RuleBook 證據

Combat Fighting 說明指出，戰鬥畫面會以當前 active character 為中心。這是 renderer 的 camera 行為，不應把 CombatMap 的絕對 tile 座標直接當成螢幕座標。

## 實作結果

- `combat.NewCombatCamera` 將 active CombatMap tile 對齊到 viewport tile `(4,2)`，`Apply` 再轉換所有 party／enemy sprite 與文字標籤。
- `game.State.CombatActiveFighter` 只讀暴露目前 turn 的 fighter，camera 不會修改回合或戰鬥資料。
- 沒有 active position 時維持原本的 deterministic formation／絕對座標 fallback，不虛構 camera origin。
- Ebiten combat renderer 已使用同一 camera transform，因此 reference placement 的較大座標不會直接被畫出畫面之外。

## 明確 boundary

本輪只實作 active-character camera transform；尚未解出原版 viewport 尺寸、地圖 tile 遮擋、scroll animation、戰場邊界、真實 Area camera state 或方向／facing 規則。

