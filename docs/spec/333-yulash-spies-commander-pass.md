# 第三百三十三輪：尤拉什間諜戰與指揮官通行證

狀態：`READY`

## 原版資料證據

GEO3 block `0x10` 的指揮官等候室位於 `(0,3,E)`。向東踏入 `(1,3)` 時，
raw terrain 為 `0x9A`；ECL3 block `0x10` 的 SearchLocation 以
`0x9A & 0x3F = 0x1A` dispatch 到 `+0x0349`：

- `4C41=0` 時把旗標設為 `1`，建立事件；重訪不再播放。
- `SETUP MONSTER 0x13,0,0x12` 後顯示散塔林間諜衝出指揮官辦公室。
- `FIGHT THE MEN` 建立 MON3 `0x13 ×1`、`0x14 ×8`、`0x15 ×2`，
  對應牧師、戰士與法師，icon blocks 同為 `0x13/0x14/0x15`。
- 勝利後隊伍獲得正面評價並被帶去晉見紅羽衛指揮官。

指揮官詢問來意時使用共用五態度 parlay menu。Clue Book 明載
HAUGHTY／NICE 有利、SLY／MEEK 不利；本輪驗證 NICE。成功分支顯示
Journal Entry 22，指揮官准許自由通行，並提供 Journal Entry 52 的尤拉什地圖，
最後由側門送回 `(1,3,E)`。

## 引擎與中文契約

- dungeon lifecycle 的順序必須是 per-turn 後接 SearchLocation。
- ECL 的空字串輸出不是可見事件，不能把 State 切成 `ModeEvent` 或吞掉後續
  terrain dispatch；只有至少一段非空白文字才可阻斷 SearchLocation。
- 保留 MON3 三種原始記錄、數量與 icon，不得以十一名泛用戰士取代。
- 顯示名稱為「散塔林牧師／散塔林戰士／散塔林法師」。
- 手札 22 與 52 必須在指揮官滿意後才解鎖；手札 52 以中文描述地圖並指回
  隨附 Adventure Journal 的原始圖像，不提前洩漏。
- 事件完成後保留 Area 3／GEO3 block `0x10`，返回 `(1,3,E)`。

## 可沿用的 Gold Box 知識

五入口 lifecycle 中，per-turn 可能產生 CALL 或空 packed text，而沒有玩家可見
事件。作品 adapter 應按實際 signal 判斷是否中止，不可只以 `len(Text)>0`
判定；空字串仍會讓 slice 長度非零。這項規則適用其他 Gold Box 作品的
per-turn → SearchLocation pipeline。

## 驗收

- real-session 從希爾斯法一路走到尤拉什等候室，再踏入 `(1,3)`。
- 驗證 terrain `0x9A`、繁中兩選項與 `4C41` exactly-once 行為。
- 驗證 1/8/2 三種 MON3 敵軍及繁中名稱。
- 戰鬥勝利後選 NICE，解鎖手札 22／52，經側門回到 `(1,3,E)`。
- 重跑 lifecycle 不得再次觸發間諜事件。
