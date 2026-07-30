# 394：Burial Glen 螳螂人防線與條件增援

狀態：`READY`

## 原始證據

| 成員 | SHA-256 |
|---|---|
| `ECL6.DAX` | `faca339db267cc88fe6f8dc6e42d7e52d362f790b6f2d49672f9758aa26175fb` |
| `GEO6.DAX` | `c2729f8b6d13ec6d497bf185841e5fb7d964dd797bd8c7c822f48053514b886c` |
| `MON6CHA.DAX` | `e739ed3dd2ccbfc6fa87d4c6d230723dafcd44ccba6f1f1f393f9a2b9f05c78b` |

ECL6 block `0x40` 的 SearchLocation 將 terrain `8Eh／8Fh／90h`
分派至 selector `0Eh／0Fh／10h`。三段都使用
`MON6CHA 40h` THRI-KREEN，但不是彼此獨立的固定戰鬥：

### Terrain `8Eh`：入口守軍

- payload `+1687h` 檢查 `4CC8h`，已完成便 EXIT。
- `+168Fh`：
  - `SETUP MONSTER 40h,2,40h`；
  - `A PARTY OF THRI-KREEN BAR YOUR ENTRANCE.`；
  - `4C02h=12`、`7F82h=1`；
  - 共用 combat routine 建立十二名 THRI-KREEN；
  - 勝利後 `4CC8h=1`。

GEO 中 `8Eh` 同時出現在 `(10,7)` 與 `(9,8)`；兩格共用
`4CC8h`，因此打過第一格後，第二格不可重播。

### Terrain `8Fh`：內側守衛

- payload `+16CEh` 檢查 `4CC9h`。
- `+16D6h`：
  - 顯示 `GUARDS HERE PREPARE FOR COMBAT.` 並 pause；
  - `SETUP MONSTER 40h,0,40h`；
  - `4C02h=6`、`7F82h=2`；
  - 建立六名 THRI-KREEN；
  - 勝利後 `4CC9h=1`。

### Terrain `90h`：營地與條件增援

- payload `+1713h` 檢查 `4CCAh`。
- `+171Bh` 顯示營地據守文字，建立十二名 THRI-KREEN，
  `4C02h=12／7F82h=3`；勝利後先寫 `4CCAh=1`。
- 接著讀 `4CC9h`：
  - 若已為一，跳至最終財寶；
  - 否則顯示其他螳螂人回應聲響，建立六名，
    `4C02h=6／7F82h=5`，勝利後寫 `4CC9h=1`。
- 再讀 `4CC8h`：
  - 若已為一，跳至最終財寶；
  - 否則顯示又有幾名落後者趕到，建立六名，
    `4C02h=6／7F82h=6`，勝利後寫 `4CC8h=1`。

所以從乾淨狀態直接進入 `90h` 是 `12→6→6` 三波；正常路徑若先清掉
`8Eh／8Fh`，則只打十二名營地守軍。remake 不能把其中任一種觀察硬編碼成
唯一固定波次。

## 財寶

所有需要的戰鬥結束後，payload `+1801h` 顯示
`YOU GATHER UP SOME VALUABLES.`，接著：

```text
TREASURE 0,0,0,2000,1500,4,6,81h
COMBAT
```

依既有 evidence-backed coin projection：

- `2000×200 + 1500×1000` copper，除以 `200` 得 `9500` gold；
- 4 gems；
- 6 jewelry；
- `ItemBlock=81h` 是一件 deterministic random item。

最後的無 monster `COMBAT` 是 treasure service boundary，不是第四場戰鬥。
原 ECL valuables 文字必須保留在財寶選單，不能被通用「發現財寶」覆蓋。

## GEO 正常玩家路徑

從 terrain `95h` `(14,10)`：

1. `(14,9)→(13,9)→(12,9)→(12,8)→(12,7)→(11,7)→(10,7)`
   抵達第一個 `8Eh`。
2. `(10,8)→(9,8)` 穿過已完成的第二個 `8Eh`，再到 `(9,9)` 的 `8Fh`。
3. 向西一步到 `(8,9)` 的 `90h`。

每一步都由 `CanMoveDungeonWrapped` 驗證。路上原版隨機遭遇可能先顯示
「看見一群……」的 pause，再進 encounter menu；玩家路徑測試需經正常
pause→FLEE continuation 回地城，不能跳過該中間 boundary。

## 實作與驗證

- 六段英文／繁中文字以 CoAB game-pack stable ID 驅動。
- raw ECL regression 驗證：
  - `8Eh` 十二名、`4CC8` 與重訪；
  - `8Fh` 六名、`4CC9`；
  - `90h` 乾淨狀態三波；
  - 預先設定 `4CC8／4CC9` 時抑制兩波增援；
  - 原始 `TREASURE` 八 operands 與 `4CCA`。
- 正常 State 玩家路徑從 Standing Stone 一路走到三道防線，驗證第二個
  `8Eh` 不重播、`90h` 只剩第一波、9500 gold／4 gems／6 jewelry／一件
  random item、繁中 valuables 文字、treasure UI 與地城 continuation。
- 通用 State／combat continuation 現在會在 treasure menu 保留當次
  `result.Text` 的資料包翻譯；沒有為 CoAB 標題寫特殊判斷。

## 明確邊界

- THRI-KREEN 多臂攻擊、特殊武器、AI、毒素及原版戰鬥動畫／音效時間碼仍未
  完成。
- 後續 terrain `91h／92h` 的蜘蛛陵墓與蛛卵巢穴仍是下一個玩家路徑缺口。
- 本輪不代表 Burial Glen、Myth Drannor 或整款遊戲已完整可通關。
