# 第三百一十五輪：法師塔入口與德拉坎德羅斯龍群騙局

狀態：`READY`

## 反組譯證據

- ECL5 block `0x32` per-turn 在 GEO5 `(6,15)`、DOS half-facing `3`
  （renderer 西向 `6`）成立時，於 `+0x0513` 顯示正走入法師塔，寫入
  `C04B/C04C/C04D = 7/15/3`，再執行 `NEWECL 0x33`。
- 同格非西向不得切換 block；SearchLocation 的 terrain `0x8F` 另有不同 handler。
- block `0x33` initial entry：
  - `+0x0051 LOAD FILES 0x33,2,0xFF`；
  - `+0x0058 LOAD PIECES 0x0E,0x0F,0xFF`；
  - `+0x0065 PICTURE 0x33`；
  - 描述受魔法保護、石工無瑕、四周高山環繞的五層塔樓庭院。
- 第一個 PRINT RETURN 後以 monster setup／PICTURE `0x3A`、兩次
  `APPROACH` 讓德拉坎德羅斯走近。他自稱一直等待隊伍，接著提供原始
  `COMBAT / WAIT / FLEE / PARLAY`。四項最終都匯流；WAIT 不先播放
  「枷印令隊伍定身」文字，其餘三項會。
- 匯流後隊伍突然出現在塔頂黑龍群中。一隻龍離群，德拉坎德羅斯假稱
  伊爾明斯特命令隊伍屠龍；枷印強迫攻擊，但目標只是煙霧消散的幻象。
- 他再次定住隊伍，向龍群發表記錄為手札條目 15 的演說。使用者提供的
  Adventurer's Journal 文字證實：他誣稱隊伍是伊爾明斯特與泰蘭索斯的
  屠龍棋子；一隻龍指出發光枷印才是控制證據，逼他解除枷印。
- 解除時 ECL 在 `+0x0415` 再寫入 `4CFF=1`；同一長流程中此位址先前已由
  火刀事件設定，因此本輪只保存 raw write，不把它另猜成「第二枚枷印計數器」。
  畫面依序使用 PICTURE
  `0x36/0x35/0x3A/0x37` 呈現龍群、幻象、德拉坎德羅斯與解除枷印。
  隨後停在四項 vertical menu：
  `ATTACK DRAGONS / ATTACK WIZARD / FLEE / PARLAY WITH THE DRAGONS`。
  本輪以此真實選擇邊界為穩定終點，不預先猜測四條後續結果。

## 引擎與中文契約

- 轉場必須由原 per-turn direction gate 與 NEWECL 產生，不得由 renderer
  看到座標後硬指定 block。
- source block 寫入的 registers、shared memory、APPROACH count、PICTURE、
  PRINT RETURN cursor 與 target block must survive the same session.
- 中文敘事按每個原始 pause 分段；手札 15 只在 ECL 真正輸出
  `JOURNAL ENTRY 15` 後解鎖，不可由進塔或先讀 README 提前加入。
- 缺少 optional WALLDEF piece 的 frontend 診斷只能寫入 renderer label，
  不得覆蓋已在 State 保存的 ECL 中文敘事。
- 640×480 畫面沿用 24px 正文；PICTURE 51／53／54／55／58 與原始人物素材採
  nearest-neighbor 整數放大。

## 驗收

- 正式長流程在 block `0x32` 原 GEO `(6,15,W)` 執行 lifecycle。
- 驗證 block `0x33`、GEO5 `0x33`、位置 `(7,15,W)`、pieces `14/15/FF`、
  PICTURE 51 與庭院繁中。
- 逐一通過德拉坎德羅斯、塔頂龍群、幻象、手札 15、解除枷印等 pause，
  最後保留四項原始 vertical-menu key。
- 驗證手札 15 在事件前不存在、事件後出現。
