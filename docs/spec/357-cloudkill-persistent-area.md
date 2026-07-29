# 第三百五十七輪：Cloudkill 3×3 毒雲與 Hit Dice 判定

狀態：`READY`（限建立、持續、施法落點效果與低 HD 移動限制）

## 問題

Cloudkill 不是綠色 Stinking Cloud 的換色版本。原版使用另一個 3×3 持續
區域、另一張 RANDCOM raster，並依 Hit Dice 決定自動死亡、帶修正的
Poison save 或完全不受影響。若共用 Stinking Cloud 的 2×2 與
cough／helpless 規則，戰術與畫面都會錯。

## 證據

### 原始資料與 consumer

`Classes/Spells.cs`、`Classes/Gbl.cs` 與 reference consumer：

- spell ID `0x5B`，Magic User level 5，range 9，Poison save；
- `CloudDirections = {8,0,1,2,3,4,5,6,7}`，即 target 本身加八方向，
  形成以 target 為中心的 3×3；
- `ovr023.SpellCloudKill` 只保留 `groundTile > 0` 且
  `move_cost < 0xFF` 的格，建立後立即對格內角色呼叫
  `in_poison_cloud(1, player)`；
- `Tile_CloudKill = 0x1C`，`BackgroundTiles[0x1C].TileIndex = 0x24`，
  全域圖號 `0x22..0x27` 對應 RANDCOM item `0..5`，故本效果使用
  RANDCOM item 2 的藍白毒雲；PC-98 公開畫面另可確認九格藍色雲的視覺語意；
- caster-level timeout 經 `CloudKillAffect` 還原這個 instance 的地板，
  移除後重畫其餘 Cloudkill，故重疊 instance 必須獨立保存。

`ovr024.in_poison_cloud(1, player)` 的 Cloudkill 分支：

| Hit Dice | 原版判定 |
|---|---|
| 0–4 | 不擲骰，直接 poisoned／killed |
| 5 | Poison save `-4`，失敗死亡 |
| 6 | Poison save `+0`，失敗死亡 |
| 7+ | 此分支不產生效果 |

`SaveVerseType.Poison = 0` 交叉確認上述 save table index。`SpellCloudKill`
反編譯尾端在九次迴圈後仍存取 `groundTile[var_16]`，此時 `var_16 == 9`；
這是明顯越界的 reference 翻譯瑕疵，不納入 remake contract。

### 移動與回合

`ovr010.CanMove` 在目的格為 poisonous cloud 時，對 Hit Dice `< 7` 且未受
相關保護、未逃跑角色把移動成本提高到剩餘移動力以上，直接阻止踏入。
`ovr009.BattleRoundChecks` 則每回合對全體呼叫 `in_poison_cloud(0, player)`。
因此忠實行為不是「成功踏入後才擲一次」：

1. 普通低 HD 角色原則上不能主動踏入 Cloudkill；
2. 施法落點覆蓋、回合開始或其他強制位移仍可讓角色處於雲內並觸發效果。

目前 Fighter 尚未保存原版魔法保護 affects，第一階段只實作可證明的 Hit Dice
門檻；保護例外維持未完成，不假裝不存在。

## Remake contract

- `PersistentAreaCloudkill` 與 Stinking Cloud 共用 ordered instance
  容器，但效果函式與格形分離。
- `Fighter.HitDice` 必須由 party offset `0xE5` 與 MON*CHA offset `0xE5`
  實際投影進戰鬥模型；只在 parser 保存而不接入 runtime 不算完成。
- `CastCloudkill` 先建立最多九個有效格，再依穩定 fighter ID 順序各處理一次：
  0–4 自動死亡、5 用 `-4` Poison save、6 用 `0`、7+ 無 impact。
- 死亡使用既有戰鬥死亡狀態與 `VisualImpactTarget.Killed`，不可只顯示文字。
- 呈現為 `windup → generic travel → persistent area reveal → per-target
  death → handoff`；CoAB JSON 指定 travel 與 RANDCOM item 2，renderer
  不寫死作品 block。
- 移動檢查在改變座標前檢查目的 footprint。碰到 Cloudkill 且
  `HitDice < 7` 時拒絕移動；攻擊相鄰敵人格的既有流程不受影響。

## 驗收

- 驗證 3×3、牆格過濾、重疊與 caster-level 消散。
- 固定 RNG 驗證 HD 4 自動死亡、HD 5 的 `-4`、HD 6 的 `0`、HD 7 無效果。
- party 與 monster 兩條 adapter 都保留 `HitDice`。
- 低 HD 目的 footprint 碰到任一 Cloudkill 格即拒絕移動，HD 7 可通過。
- 正常玩家持有 memorized `0x5B`，由快捷鍵、tile cursor、Enter 施放後才
  消耗 slot。
- Docker／Xvfb 截圖顯示 3×3 藍白原版 raster、人物仍在雲層之上。

## 尚未包含

- protect magic 與尚未命名 affect `0x6F／0x7D／0x85` 的免疫。
- 每回合重複 `in_poison_cloud(0)` 的完整 affect／死亡文字時序。
- DOS Cloudkill 動態影片的絕對時間碼與 wall-clock delay；目前 PC-98
  screenshot 只標 `layout-only`，不得據此宣稱完整動畫 fidelity。
