# 395：Burial Glen 蜘蛛陵墓與幽魂警告

狀態：`READY`

## 原始證據

| 成員 | SHA-256 |
|---|---|
| `ECL6.DAX` | `faca339db267cc88fe6f8dc6e42d7e52d362f790b6f2d49672f9758aa26175fb` |
| `GEO6.DAX` | `c2729f8b6d13ec6d497bf185841e5fb7d964dd797bd8c7c822f48053514b886c` |
| `MON6CHA.DAX` | `e739ed3dd2ccbfc6fa87d4c6d230723dafcd44ccba6f1f1f393f9a2b9f05c78b` |

ECL6 block `0x40` 將 terrain `91h／92h` 分派至 selector
`11h／12h`。兩段都使用蜘蛛素材 `41h`，實際 monster record 是
`MON6CHA 42h` GIANT SPIDER。

## Terrain `91h`：掛滿蛛網的陵墓

payload `+1830h`：

1. `4CCBh == 1` 時 EXIT。
2. `SETUP MONSTER 41h,0,41h`。
3. 顯示：
   `WEBS FESTOON THIS MAUSOLEUM. THE WEBS ARE INHABITED.`
4. pause 後寫 `7F82h=7`、`4C00h=8`。
5. 共用 combat routine 建立八隻 GIANT SPIDER。
6. 勝利 continuation 寫 `4CCBh=1`，重訪直接 EXIT。

GEO6 block `0x40` 中 terrain `91h` 位於 `(9,2)`。

## Terrain `92h`：漏斗狀蛛網

payload `+1884h` 先以 `4CCCh` 做一次性 gate；未完成時從 `+188Ch`
開始：

1. 顯示 `YOU SEE A FUNNEL OF WEBS.` 並 pause。
2. 寫 `7ECBh=1`，比較黛米爾好感 `4CBAh` 與 `80h`。
3. 若 `4CBAh < 80h`：
   - 直接跳到攻擊分支；
   - 不顯示幽魂警告；
   - 沒有 YES／NO 選單。
4. 若 `4CBAh >= 80h`：
   - 顯示幽魂警告蜘蛛會拚死保護巢穴；
   - 詢問 `DO YOU CONTINUE?`，選項為 `YES／NO`；
   - `NO` 直接 EXIT，不寫 `4CCC`，所以重訪仍會發生；
   - `YES` 寫 `7ECBh=2` 後進攻擊分支。
5. 攻擊分支 `+1908h`：
   - 顯示蜘蛛躍出，後方可見蛛卵；
   - 在戰鬥前先寫 `4CCCh=1`；
   - `SETUP MONSTER 41h,0,41h`；
   - pause 後寫敵方 attack-roll work `7F70h=2`；
   - `7F82h=0`、`4C00h=4`；
   - 建立四隻 GIANT SPIDER。

`4CCC` 是進攻後即消耗，而不是勝利後才寫。若未來實作撤退／戰敗重返，
必須維持這個原始時序，不能把它改成「打贏才完成」。

GEO6 block `0x40` 中 terrain `92h` 位於 `(10,1)`。

## 正常玩家路徑

從第 394 輪 terrain `90h` `(8,9)`：

```text
(9,9) → (9,8) → (9,7) → (10,7) → (11,7)
→ (11,6) → (11,5) → (11,4) → (11,3)
→ (10,3) → (9,3) → (9,2 / terrain 91h)
```

擊敗八隻巨蛛後，再走：

```text
(9,1) → (10,1 / terrain 92h)
```

所有步驟均由 `CanMoveDungeonWrapped` 驗證。路徑會再經過已完成的
`8Eh／8Fh`，必須依 `4CC8／4CC9` 保持安靜；隨機遭遇仍需走正常
pause→encounter menu→FLEE continuation。

目前正常玩家路徑的 `4CBAh` 高於 `80h`，因此驗證：

1. 幽魂警告與繁中 YES／NO。
2. 先選 NO，回地城且 `4CCC` 未設定。
3. 重踏後再選 YES。
4. 蛛卵文字、戰前 `4CCC=1`、四隻 GIANT SPIDER 與敵方命中 `+2`。
5. 勝利後回地城，重踏不重播。

## 資料化與測試

- 四段敘事及 YES／NO 使用 CoAB game-pack stable ID；繁中文字沒有寫死在
  Go 測試或 frontend。
- raw ECL regression 分開覆蓋：
  - terrain `91h` 八隻與 `4CCB`；
  - 高好感警告＋NO 可重訪；
  - 高好感 YES＋四隻＋`7F70=2`；
  - 低好感直接略過警告。
- 正常 State 玩家路徑由 Standing Stone 一路走到兩座陵墓，驗證資料包文字、
  選項、旗標時序、敵方 attack-roll modifier 與重訪。
- 現有通用 ECL、option localization、combat work projection 已足夠承載；
  沒有新增 CoAB 專用 Go 劇情分支。

## 明確邊界

- GIANT SPIDER 毒素、網、AI、原版戰鬥動畫與音效時間碼仍未完成。
- 原文提到蛛卵，但這段 ECL 沒有額外的摧毀／搜刮蛛卵選單或財寶；不得從
  畫面敘述自行發明玩法。
- Burial Glen 更西側 terrain `96h`、`89h／8Ah／8Bh` 與區域出口仍待後續
  玩家路徑稽核。
