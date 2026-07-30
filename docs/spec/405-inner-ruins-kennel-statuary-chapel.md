# 405：內部遺跡犬舍、活動雕像與私人禮拜堂

狀態：`READY`

## 範圍

本輪解碼 ECL／GEO `43h` 的三個戰鬥房間：

- terrain `87h`：地獄犬犬舍；
- terrain `88h`：活動雕像；
- terrain `89h`：私人禮拜堂。

私人禮拜堂由正常 Standing Stone 起始玩家路徑完成。犬舍與活動雕像的
raw ECL 分支、怪物與一次性行為已完成，但正常路徑必須先通過 terrain
`82h–85h` 的提朗瑟克斯／無名者最終儀式；本輪沒有用 direct entry 越過
這個主線 gate。

## 原始資料

- `ECL6.DAX`
  - SHA-256：
    `faca339db267cc88fe6f8dc6e42d7e52d362f790b6f2d49672f9758aa26175fb`
  - decoded block `43h`：5,363 bytes。
- `GEO6.DAX`
  - SHA-256：
    `c2729f8b6d13ec6d497bf185841e5fb7d964dd797bd8c7c822f48053514b886c`
  - decoded block `43h`：1,026 bytes。
- `MON6CHA.DAX`
  - SHA-256：
    `e739ed3dd2ccbfc6fa87d4c6d230723dafcd44ccba6f1f1f393f9a2b9f05c78b`
  - 相關 records：`44h／45h／46h／48h`。

所有命令、測試與地圖路線分析均在 `--network none` Docker 容器內執行。

## GEO 座標

| Terrain | 座標 |
|---|---|
| `87h` | `(1,9)` |
| `88h` | `(4,9)` |
| `89h` | `(8,13)`、`(9,13)` |
| 儀式區 | `82h (7,13)`、`83h (7,11)`、`84h (6,11)`、`85h (5,11)` |

由第 404 輪臥房 `(10,12)` 到私人禮拜堂：

```text
(10,12) → (9,12) → (9,13)
```

戰勝後回到東側走廊，再接近儀式區：

```text
(9,13) → (9,12) → (10,12) → (10,11) → (9,11)
       → (8,11) → (8,10) → (7,10)
```

`(7,10)` 南側 `(7,11)` 是 terrain `83h`，確定觸發最終儀式。西翼的
合法最短路線會穿過 `83h／84h／85h` 之一；因此不能在未處理儀式前聲稱
犬舍與活動雕像已由正常玩家路徑完成。

## Raw ECL 行為

### terrain 87h：犬舍

- `+0AEFh` 寫 `4C01h=1`。
- 顯示硫磺味、房間被改成犬舍，怪物起身。
- 原始流程保留一個文字為空的額外 PRESS pause，之後才顯示
  「提朗瑟克斯的爪牙衝上前攻擊」。
- 建立 MON6CHA `44h` HELL HOUND ×10。
- 戰後離開；重訪靜默。

### terrain 88h：活動雕像

- `+0B74h` 寫 `4C02h=1`。
- 顯示白骨與猙獰雕像開始活動。
- 顯示同一段爪牙攻擊文字。
- 建立 MON6CHA `45h` MARGOYLE ×10。
- 戰後離開；重訪靜默。

### terrain 89h：私人禮拜堂

- `+0BD9h` 寫 `4C03h=1`。
- 顯示雅致私人禮拜堂與衣著考究的老者。
- 老者稱隊伍為提朗瑟克斯的工具，並說班恩最近太偏愛他，所以讓隊伍死去
  反而更能傷害提朗瑟克斯；其他祭司隨即包圍。
- 建立：

  ```text
  MON6CHA 48h HIGH PRIEST    ×1
  MON6CHA 46h PRIEST OF BANE ×4
  ```

- 戰後顯示 remake 共用勝利提示，Continue 回到地城。

## 怪物記錄錨點

| Block | 名稱 | HP | 等級／HD | AC raw | 傷害 |
|---|---|---:|---:|---:|---|
| `44h` | HELL HOUND | 35 | Fighter 7 | 56 | `2d4` |
| `45h` | MARGOYLE | 30 | Fighter 6 | 58 | `1d6` |
| `46h` | PRIEST OF BANE | 48 | Cleric 8 | 50 | `1d6` |
| `48h` | HIGH PRIEST | 60 | Cleric 10 | 50 | `1d6` |

表中欄位是本輪 parser 可觀測值；祭司完整法術表、AI、特殊能力、動畫與音效
尚未因這張表而宣稱完成。

## 實作與驗證

- game-pack 新增五個繁中 stable IDs：
  - `myth-drannor.inner.kennel`
  - `myth-drannor.inner.statuary`
  - `myth-drannor.inner.minions-attack`
  - `myth-drannor.inner.chapel`
  - `myth-drannor.inner.chapel-priest`
- `TestRealInnerRuinsKennelStatuaryAndChapel` 直接執行 raw ECL，逐 pause 驗證
  三個旗標、四種 MON6 records、三場戰鬥與重訪。
- `TestRealPlayerPathStandingStoneToBurialGlen` 正常走到私人禮拜堂，完成
  HIGH PRIEST ×1／PRIEST OF BANE ×4 戰鬥，再走到 terrain `83h` 前一格。

## 完成邊界

犬舍與活動雕像目前是 raw branch 完成、正常玩家路徑尚未完成；其真正
handoff 依賴下一輪 terrain `82h–85h` 最終儀式。祭司法術、怪物特殊能力、
原版戰鬥場地、動畫、音效、戰利品，以及二樓與完整結局仍未完成。
