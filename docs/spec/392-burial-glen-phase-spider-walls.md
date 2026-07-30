# 392：Burial Glen 相位蜘蛛牆防線

狀態：`READY`

## 原始證據

| 成員 | SHA-256 |
|---|---|
| `ECL6.DAX` | `faca339db267cc88fe6f8dc6e42d7e52d362f790b6f2d49672f9758aa26175fb` |
| `GEO6.DAX` | `c2729f8b6d13ec6d497bf185841e5fb7d964dd797bd8c7c822f48053514b886c` |
| `MON6CHA.DAX` | `e739ed3dd2ccbfc6fa87d4c6d230723dafcd44ccba6f1f1f393f9a2b9f05c78b` |

ECL6 block `0x40` 的 SearchLocation entry：

1. 讀 `C04Fh` terrain，與 `7Fh` 做 AND 後寫入 `7F79h`。
2. `ON GOTO` selector `0x13／0x14` 分別跳至 payload
   `+195Eh／+19B1h`。
3. `+195Eh`：
   - `4CCDh == 1` 時直接 EXIT；
   - `SETUP MONSTER 41h,1,41h`；
   - 原文 `AS YOU ENTER, SPIDERS COME OUT OF THE SOLID WALLS.`；
   - pause 後保存 `7F82h=8`、`4C01h=10`，由共用 combat routine
     產生十隻 `MON6CHA 41h` PHASE SPIDER；
   - 勝利 continuation 寫 `4CCDh=1`。
4. `+19B1h`：
   - `4CCEh == 1` 時直接 EXIT；
   - 同一 PHASE SPIDER template；
   - 原文 `GLOWING SPIDERS SKITTER FORWARD AT YOUR APPROACH.`；
   - 保存 `7F82h=9`、`4C01h=8`，產生八隻；
   - 勝利 continuation 寫 `4CCEh=1`。

GEO6 block `0x40` 的正常可行路徑：

- 黛米爾 `(13,14)` → `(13,13)` → `(13,12)` → `(12,12)` →
  `(12,11)` → terrain `93h` `(12,10)`，共五步。
- `(12,10)` → `(12,9)` → `(13,9)` → `(14,9)` →
  terrain `94h` `(14,8)`，再四步。

兩條路徑均由 `CanMoveDungeonWrapped` 驗證原始牆／門；途中 random
encounter 依正常 FLEE continuation 恢復同一座標與 ECL session。

## 實作與驗證

- 英文及繁中敘事以 stable message ID
  `myth-drannor.phase-spider-wall`／
  `myth-drannor.phase-spider-glowing` 放入 CoAB game pack。
- raw ECL session regression 驗證兩組文字、monster spawn、combat work、
  completion flags 及重訪 EXIT。
- 正常玩家 regression 從 Standing Stone、世界旅行、Burial Glen、
  紅網、墳墓、黛米爾一路步行到兩道蜘蛛防線，分別完成十隻與八隻
  PHASE SPIDER 戰鬥，再驗證 `4CCD／4CCE` 與重踏不重播。
- 沒有為這兩個事件新增 Go 劇情分支；既有 ECL runtime、MON6 adapter、
  combat continuation 與 JSON localization 已足以執行原始流程。

## 明確邊界

- terrain `95h` 的六隻 PHASE SPIDER 後另有骨堆
  `LOOT／REPLACE IN CRYPTS` 選單、treasure 與好感副作用，留待下一個
  milestone，不與本輪兩個單純戰鬥事件混稱完成。
- PHASE SPIDER 的位移相位、毒素與完整 AD&D 特殊能力尚未證明；本輪只完成
  原作 encounter composition、ECL continuation 與目前通用戰鬥投影。
- 尚未保存這兩場的 DOS 動態影片時間碼，不能宣稱攻擊動畫／音效 fidelity
  已完成。
