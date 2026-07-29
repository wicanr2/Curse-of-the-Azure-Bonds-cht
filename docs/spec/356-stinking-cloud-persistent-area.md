# 第三百五十六輪：Stinking Cloud 持續區域

狀態：`READY`（限 `0x22` 建立四格雲、初次 Poison save、持續遮罩與消散）

## 問題

既有 remake 只有一次性的 projectile／impact 時間軸，無法表達原版
Stinking Cloud 在戰場上保留數回合、妨礙後續行動，或多片雲重疊後分別
消散。若只在 projectile 終點畫一張貼圖，範圍、地板還原、角色狀態與回合
規則都會錯。

## 證據

### DOS/EGA runtime

公開 PC／DOS 影片 `wwYsij1wDC4` `00:42:25.20–27.00`：

| 時間 | 可見行為 | 證據 |
|---|---|---|
| `25.20–25.40` | `05/85` 青色 projectile 飛向目標格 | 既有三張逐 100ms frame |
| `25.60` | projectile 消失，顯示 `CREATES A NOXIOUS CLOUD` | [`combat-stinking-cloud-created-004225600.png`](../reference/original-dos/combat-stinking-cloud-created-004225600.png) |
| `26.20` | 四個綠白雲格出現，角色仍畫在雲層之上 | [`combat-stinking-cloud-persistent-004226200.png`](../reference/original-dos/combat-stinking-cloud-persistent-004226200.png) |
| `27.00` | 雲格跨到下一段文字仍留在戰場 | [`combat-stinking-cloud-next-action-004227000.png`](../reference/original-dos/combat-stinking-cloud-next-action-004227000.png) |

影片逐像素證明建立順序、四格外觀、角色／雲的圖層與跨訊息持續；不能單靠
這一小段推算精確 DOS delay 或完整持續回合數。

### Reference consumer

公開 reference `ovr023.SpellStinkingCloud` 與 `Classes/Gbl.cs`：

- spell ID `0x22`，Magic User level 2，range 3，Poison save；
- `SmallCloudDirections = {8,2,3,4}`，`MapDirectionDelta` 對應
  `(0,0)、(1,0)、(1,1)、(0,1)`，所以是以 target 為西北角的 2×2，
  不是以中心為圓心；
- 每格先要求原地板 `groundTile > 0` 且 `move_cost < 0xFF`；
- 原地板保存在 `GasCloud.groundTile[]`，戰場 tile 改為
  `Tile_StinkingCloud = 0x1E`；
- 同一角色 footprint 若碰到兩個雲格，只處理一次；
- 建立後立刻對格內角色呼叫 `in_poison_cloud(1, player)`。

`ovr024.in_poison_cloud`：

- 通過 Poison save：顯示 `starts to cough`，套用一回合
  `stinking_cloud`，該 affect 禁止 use／cast；
- 失敗：顯示 `chokes and gags from nausea`，套用 `d4+1` 時間的
  helpless；
- helpless、animate dead 與兩個尚未命名免疫 affect 不再重複處理
  Stinking Cloud。

`ovr013.StinkingCloudAffect`：

- caster-level timeout 到期時顯示 `The air clears a little...`；
- 先還原這一片雲保存的原地板或 downed-player tile；
- 移除該 instance 後，重新把所有仍存在的 Stinking Cloud 畫回去，
  因此重疊雲不能用單一 boolean tile 表示。

Clue Book 另明說 Stinking Cloud 會持續數回合，可用來保護側翼、導引敵人，
並使怪物 helpless。這只支持用途與持續性，不取代上述 consumer 細節。

## Remake contract

規則層保存 ordered `PersistentArea` instances，而非改寫 renderer 的地形：

```text
kind, caster, center, created round, expires round
cells[] = position
```

原地形持續留在 title terrain adapter，不被 cloud core 覆寫，因此等價保留
原版 `groundTile[]` 的還原能力。每一格可同時被多個 instance 覆蓋；查詢時
只要仍有雲覆蓋就顯示 Stinking Cloud，移除一個 instance 不得清掉其他
instance。

`CastStinkingCloud` 必須原子化：

1. 驗證 caster、level、target 與四個候選格；
2. 只保留可通行格；
3. 建立持續到 `currentRound + casterLevel` 的 instance；
4. 依穩定 fighter order 對 footprint 相交角色各處理一次 Poison save；
5. 回傳 cells、saved／helpless duration，供 UI 與測試使用。

呈現層沿用 `VisualAreaSpell`：

```text
windup -> generic travel -> persistent-area reveal -> handoff
```

`combat_visuals` 新增作品中立 `area` phase，CoAB JSON 才指定
Stinking Cloud 使用全域 background graphic `0x1E`，其 raster 位於
`BackgroundTiles[0x1E].TileIndex = 0x26`，再依全域 graphic namespace
解析為 RANDCOM item 4。規則層不知道 RANDCOM，renderer
也不得重擲 saving throw。

## 驗收

- 固定 RNG 驗證成功 cough 一回合、失敗 helpless `d4+1`。
- 2×2 footprint 命中同一 large combatant 只處理一次。
- 牆格不建立雲，零個有效格時不消耗 spell slot。
- 兩片雲重疊，第一片到期後重疊格仍存在。
- caster level 回合到期後雲格清除。
- 正常玩家必須持有 memorized `0x22`，按鍵、tile cursor、Enter 後才消耗。
- Docker／Xvfb 畫面證明 travel 後四格綠白原版 raster 出現，人物畫在雲上。

## 尚未包含

- 移動進入／離開雲時每一步的完整 `in_poison_cloud` 呼叫時機。
- coughing 對 AC 的原內部值轉換。
- 免疫 affect `0x6F／0x7D` 的名稱與來源。
- exact DOS wall-clock delay。
- Cloudkill `0x5B` 的 3×3、HD 4–6 即死規則與獨立消散 instance；它必須
  另以相同 persistent-area contract 實作，但不得把兩種毒雲效果混成一條。
