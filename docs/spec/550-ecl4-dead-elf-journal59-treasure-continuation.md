# 第五百五十輪：眼魔洞穴死精靈、手札 59 與戰利品續跑

狀態：`READY`（限 ECL4 block `0x22` 的死精靈皮袋分支與同一 session 的戰利品
服務；不是 Dexam、洞穴出口或完整主線）
日期：2026-08-12

## 目的與勘誤

本輪把第 548 輪 A2 事件後的正常玩家路徑往前推進，但不再以 `(15,1)` 的搜尋邊
假裝已找到洞穴出口。該邊目前只有 `strong inference`，不能替代原始 GEO 移動或
runtime trace。

正常 session 現在會從洞穴 E1 `(5,7,W)` 經 A2、三次 `PRESS` 與原始
`C04B/C04C/C04D=13/1/3` 位置交易，抵達死精靈 `(13,1,W)`；玩家可選擇
`EXAMINE REMAINS → PICK UP POUCH`，承受陷阱、取得手札 59，結束原始無生怪
`COMBAT`／戰利品服務後留在同一格。所有故事文字、手札與選項均從 CoAB game-pack
解析；engine 不知道 Dexam、座標、地圖圖例或手札號碼。

## 原始證據

輸入：

- `curseoftheazurebonds.zip` SHA-256：
  `c98698a6271c17177dfdb27f34b0389b7d34f58ef206e92575393f4655f5b26d`
- 其中 `ECL4.DAX` block `0x22`、A2 continuation 的起點 `+050A`。
- `Curse-of-the-Azure-Bonds_Misc_DOS_EN_Adventurers-Journal.pdf` SHA-256：
  `33471b8200d4aab93d2f52065601bebe548cb535a29064a3a0a47c2e833d88c9`。

在 A2 的三次 `PRESS` 與死精靈 `EXAMINE REMAINS` 選擇之後，bounded raw session
產生以下可重現序列：

| 步驟 | 原始 ECL 結果 | 證據等級 |
|---|---|---|
| 檢查遺骸 | 「皮袋埋在腐朽衣物下」文字，再等待按鍵 | `exact` |
| 第一次按鍵 | `PICK UP POUCH`／`POKE AT POUCH`／`CAST FIND TRAP`／`LEAVE` 選單 | `exact` |
| 拿起皮袋 | 「A GAS TRAP GOES OFF!」並在 `+08A6` 寫入 `7F79h=0` | `exact` |
| 下一次按鍵 | 地圖文字、`1d12` 傷害 request（flags `C0h`、save flags `3`）並在 `+09AC` 寫入 `4C07h=80h` | `exact` |
| 再次按鍵 | 原始 `COMBAT` request，PC `+089C`；此段沒有 monster spawn | `exact`（ECL request） |

地圖原文明說「標出 Dexam 的祭壇與似乎通往外面的路徑」，並把它放入 Journal
Entry 59。原始 Adventurer's Journal 的條目 59 地圖圖例還列出門、拱門、動物聲響、
牛頭人聲響、Dexam 的神殿與祭壇。這些圖例文字以 PDF 與既有 OCR 交叉讀取；文字
手札的描述標為 `strong inference`，因為原始地圖 bitmap 尚未接入遊戲內 Journal
renderer。地圖 bitmap／實際出口 route 仍是後續 P0 工作，不能由描述反推坐標或牆值。

`4C03h=1` 在這條 trace 中保持不變；本規格只記錄該路徑的 raw state，未替其命名。
同樣地，`4C07h=80h` 只記錄事件寫入，尚未宣稱它是全局地圖或戰鬥規則欄位。

## Remake 與資料邊界

- `dexam.dead-elf.gas-trap` 是由原始英文 fragments 對應的 `text_rule`；前端沒有
  hardcode 中文陷阱句子。
- `dexam.dead-elf.map` 以 `journal_message_ids: ["journal.59"]` 解鎖可在遊戲內讀取
  的手札；這是通用 game-pack／Journal resolver 行為。
- State 將無 monster spawn 的 ECL `COMBAT` request 轉為既有戰利品服務邊界。正常
  session 在 `TREASURE_EXIT` 後執行地城 lifecycle，清空暫存物品與選項並回到
  `(13,1,W)`；這是 remake runtime contract，不把它誤稱為原始戰鬥 AI 證據。
- 手札 59 現先提供可讀的繁中圖例摘要，並明示原始圖像保存於《冒險者手札》。待
  原圖抽取、版面驗證後再將 bitmap 放入遊戲內，不以合成圖替代原作證據。

## 驗證

Docker 內以本機 Go toolchain、唯讀原始映像與 nested engine replace 執行：

```text
go test -v ./internal/ecl \
  -run '^TestRealECL4CaveDeadElfPouchUnlocksJournal59AndRequestsCombat$' -count=1
go test -v ./gamepack \
  -run '^Test.*Journal.*|TestBeholderCaveMapHandoffContinuesSameECLResult$' -count=1
go test -v ./internal/game \
  -run '^TestRealNewGameContinuesFromHapToBeholderCaveEntrance$' -count=1
```

最後一項從真正新遊戲開始，以普通移動與正常 ECL continuation 走到手札 59，並
驗證 `4C03h=1`、`4C07h=80h`、兩個 pending treasure item、`TREASURE_EXIT` 後的
同座標地城 lifecycle。它不設定洞穴座標、不注入 Dexam 戰鬥，也不呼叫推測性搜尋邊。

## 範圍限制與下一步

- 本輪沒有證明手札地圖上的「似乎通往外面」是哪個 GEO route；下一步先從原始
  GEO4、ECL4、手札圖與 DOS runtime 建立候選路徑，再以正常輸入驗證。
- `TestRealBeholderCaveDexamAndZhentilBattles` 仍是局部 raw-event regression，不能
  作為本輪 normal-path 的替代證據。
- Dexam、梅杜莎、牛頭人、護符、洞穴出口、奧莉芙／迪姆斯沃特離隊與回世界圖均
  尚未接入同一新遊戲 session。
- 原版圖像、戰鬥演出、音效與完整中文校對仍屬 P1，不能因手札文字已可讀就宣稱
  完成。
