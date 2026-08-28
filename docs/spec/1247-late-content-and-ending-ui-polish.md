# 第一千二百四十七輪：後期內容與結局 UI polish

狀態：`READY`

## 內容抽樣與修正

本輪從機械稽核不易發現的「內容被流暢縮寫」反向抽樣商店、神殿、訓練場、
Journal 41／42／44／45／46／48、全部酒館傳聞與結局五頁，並以英中 locale、
DOS embedded strings 與 PC-98 結局字串交叉核對。

- 英文共有 13 則傳聞以 `As you consume the local excuse for food and drink` 開頭，
  嘲諷當地食物與飲料；繁中原本只有第 59 則保留，另外 12 則只剩事件事實。
  本輪全部補回相同敘事語氣，並加入資料驅動測試：只要英文仍有該前綴，繁中就
  必須保留「難稱為食物與飲料」，且候選數必須是 13。
- `Bastard Sword` 在敘事已訂為「混種劍」，但物品詞表仍輸出「闊刃大劍」。
  `item_type_22`／`item_name_22` 已統一，並由原版六章 ITEM 資料重生
  `docs/audit/item-names.md`：253 件、115 種名稱、126 個使用中名稱編號、缺譯 0。
- DOS 結局 `gauntlet contacts the pool, it contracts and shatters it` 搭配 PC-98
  明確句子，證明光芒之池收縮後震碎的是摩安德護手。繁中第 2 頁原寫成池子本身
  碎裂，已改為「池子便收縮，將護手震碎」。

## 結局 UI

spec 1154 已證實結局第 4 頁在原版有 8 行，其餘頁各 4 行。前端先前把所有結局頁
送進荒野共用的 4 行 message box，繁中第 4 頁會被安靜裁掉後半。

- `State.EndingScenePage` 明確公開目前結局頁狀態，frontend 不用比對翻譯文字猜頁面。
- 結局改用 36 個 CJK 字元寬、8 行、28px 行距的獨立安全矩形；繼續選項位於其下，
  不與最長頁重疊。
- `-ending-page 1..5` 由正式結局 transaction 開始，再以正常 `Select(0)` 推頁；它是
  renderer checkpoint，不冒稱完整通關路徑。
- 人工檢視 [`ending-page-4-remake.png`](../screenshots/ending-page-4-remake.png)：
  640×480，SHA-256
  `78b51fb39f9d6b6b5c540fe9f01471634668117345c3117bf3418317571ad5f7`；全文完整，
  無截斷、越界或與繼續選項重疊。

## 驗證

- `go test ./cmd/azure-bonds-game ./internal/game ./internal/localeaudit ./cmd/item-name-audit ./gamepack ./internal/screenshotmanifest`
  通過。
- `go run ./cmd/item-name-audit -output docs/audit/item-names.md`：
  `items=253 names=115 numbers=126 missing=0`。
- 結局 preview 測試驗證 1-based 邊界與經三次正常選擇抵達 zero-based page 3。

本輪證據支持內容語意與文字安全矩形，不宣稱結局畫面逐像素等同 DOS；其等級為
`layout-only`（文字內容本身另有 exact 字串證據）。
