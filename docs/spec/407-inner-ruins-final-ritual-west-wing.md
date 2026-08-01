# 407：提朗瑟克斯／無名者最終儀式與西翼正常玩家路徑

狀態：`READY`

## 範圍

本輪完成 ECL／GEO block `43h` terrain `82h–85h` 共用的最終儀式，並由
Standing Stone 起始長玩家路徑穿越這個主線 gate，接續完成第 405 輪尚只有
raw branch 證據的活動雕像 `88h` 與犬舍 `87h`。

這仍不是迷斯卓諾章節或全遊戲結局：儀式後提朗瑟克斯逃走，玩家只擊敗留下
的十四名爪牙；後續追擊、真正最終戰與結局仍待解碼及正常路徑驗收。

## 原始資料

- `ECL6.DAX` SHA-256：
  `faca339db267cc88fe6f8dc6e42d7e52d362f790b6f2d49672f9758aa26175fb`
  - decoded block `43h`：5,363 bytes。
- `GEO6.DAX` SHA-256：
  `c2729f8b6d13ec6d497bf185841e5fb7d964dd797bd8c7c822f48053514b886c`
  - decoded block `43h`：1,026 bytes。
- `MON6CHA.DAX` SHA-256：
  `e739ed3dd2ccbfc6fa87d4c6d230723dafcd44ccba6f1f1f393f9a2b9f05c78b`
  - 本輪敵人 records：`44h／45h／48h`。
- 使用者提供的 Adventurer's Journal PDF；OCR 定位到原書第 20 頁的
  Journal Entry 48，翻譯仍以 PDF 頁面人工核對，不把 OCR 錯字當原文。

所有執行、解析、測試、抓圖與轉檔均在 Docker 內完成；原始 ZIP／PDF 唯讀。

## Raw ECL 儀式序列

terrain `82h (7,13)`、`83h (7,11)`、`84h (6,11)`、`85h (5,11)` 分派
同一段事件。首次執行於 payload `+0537h` 寫 `4C00h=1`，因此同區其餘三格
在儀式完成後靜默，不會重播。

原始互動順序如下；括號是 PICTURE block，沒有括號者維持第一人稱／文字框：

1. 一個聲音命令隊伍向前，枷印迫使眾人走到房間中央。
2. 提朗瑟克斯宣告可透過枷印控制隊伍。
3. `PICTURE 47h`：演說並要求記錄 Journal Entry 48。
4. `PICTURE 46h`：隊伍被迫把三件神器交給祭司。
5. `PICTURE 4Ch`：提朗瑟克斯命令祭司把神器丟入光芒之池。
6. 祭司逐一拋入神器，池水旋動後物品消失。
7. `PICTURE 47h`：提朗瑟克斯交出寫有解除枷印密語的羊皮紙。
8. 他命令受枷者同行，準備完成最後法術。
9. `PICTURE 46h`：祭司掀開兜帽，原來是無名者；他把神器拋回隊伍。
10. `PICTURE 47h`：提朗瑟克斯擊倒無名者；無名者臨終默念密語。
11. 枷印的控制力消退，但提朗瑟克斯印記仍在。
12. 隊伍取回神器；提朗瑟克斯下令爪牙攻擊後逃走。
13. 顯示共用的爪牙進攻文字，進入戰鬥。

ECL 在人物切換處寫 `7EE1h=46h→FFh`、稍後寫 `43h→FFh`，對應祭司／
無名者人物素材生命週期。這段沒有發出 `FIND ITEM`、`DESTROY ITEM` 或
`TREASURE`；三神器的交出與取回是同一事件內的敘事，不得由 frontend
自行刪除再重建 inventory。

戰鬥建立順序與 exact 數量：

| MON6CHA | 敵人 | 數量 |
|---|---|---:|
| `48h` | HIGH PRIEST | 2 |
| `44h` | HELL HOUND | 6 |
| `45h` | MARGOYLE | 6 |

以 `4CBDh=0／1` 分別執行同一 raw session，直到戰鬥 boundary 的文字、
PICTURE、`4C00h` 與三組 spawn 完全相同。這只證明該旗標不改變這段已觀察
輸出；不能擴大解釋成羅剎妖協議在整個章節都沒有其他效果。

## Journal Entry 48

game-pack 以 stable ID `journal.48` 保存完整繁中內容；只有 ECL 顯示
「記入 Journal Entry 48」後才加入遊戲內手札。內容說明提朗瑟克斯已取得
三神器，並打算利用青色枷印像光芒之池一樣，在隊伍成員身體、光芒之池及
其他受枷者之間轉移附身。

儀式的十二段敘事與手札均由 game-pack text rules 配對原始 token；State、
frontend 與測試不複製繁中顯示字串。

## 正常 GEO 玩家路徑

第 405 輪停在 `(7,10)`。本輪依雙側牆面驗證向南進入 `(7,11)` terrain
`83h`，完成十四人戰後，`4C00h=1` 使相鄰儀式格靜默，路線得以向西：

```text
(7,11) → (6,11) 84h → (5,11) 85h
       → (5,10) → (4,10) → (4,9) 88h 活動雕像
       → (4,10) → (3,10) → (2,10) → (1,10) → (1,9) 87h 犬舍
```

活動雕像的十隻 MARGOYLE 與犬舍的十隻 HELL HOUND 均由正式 combat
scheduler 戰勝並續跑回原地城。犬舍原始空文字 PRESS pause 仍保留。
途中龍盔另正常觸發「二樓東北角」提示，已新增 stable ID
`myth-drannor.inner.helm-northeast`，不再顯示未翻譯英文。

## 驗證

- `TestRealInnerRuinsTyranthraxusAndNamelessFinalRitual`：兩種 `4CBDh`
  初始值、十三段 raw sequence、六個 PICTURE、十四人 spawn、戰後
  continuation、`4C00h` 與重訪靜默。
- `TestRealPlayerPathStandingStoneToBurialGlen`：從 Standing Stone 的完整
  已累積路徑進入儀式、逐段繁中與手札 48、戰勝十四人，再沿合法 GEO
  完成活動雕像與犬舍兩戰，驗證 `4C02h／4C01h=1`。
- `-inner-ritual` 是 deterministic 畫面 checkpoint，從正式 ECL session
  顯示 `PICTURE 47h` 與繁中手札提示；它只作畫面稽核，正常玩家路徑證據
  仍以上述長回歸為準。

## 尚未完成

- HIGH PRIEST、HELL HOUND、MARGOYLE 的完整法術／特殊能力與 DOS 動態演出。
- 無名者死亡、神器飛行與枷印解除是否有額外原版音效／動畫的 runtime oracle。
- terrain `86h` 之後的東北角追擊、提朗瑟克斯真正最終戰、結局與章節離開。
