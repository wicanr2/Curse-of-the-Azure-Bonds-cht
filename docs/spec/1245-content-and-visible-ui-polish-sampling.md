# 1245 — 內容與玩家可見 UI polish：第一批代表性抽樣

狀態：`READY`（本批修正與樣本）；整體抽樣 polish 已由後續 spec 1247 收尾。

## 決策與範圍

使用者決定同時推進內容 polish 與玩家可見 UI polish。原版忠實 theme、流程與
素材比例維持不變；本批不以拉伸、重畫或簡化流程掩蓋問題。驗證採風險導向代表性
抽樣，發現系統性差異才擴大同類樣本。

## 內容抽樣

機械閘門結果：

- `locale-drift-audit`：玩家文字缺鍵與違規 0。
- `glossary-audit`：138 個有效詞條，問題 0。
- `cht-proofread-audit`：1,598 條中譯；未翻、同句多譯、半形標點皆為 0；
  1,274 條全形標點正對照成立。

人工抽樣涵蓋長篇手札、結局相關手札、世界事件、酒館傳聞、迷斯卓諾、巫師塔、
提爾佛頓、散提爾堡、戰鬥員名稱與建角職業。確認並修正：

1. `journal.21` 的 `bastard sword` 不應譯成「雙手劍」，改為「混種劍」。
2. 同句 `Sis` 是普通稱呼，不是名為「西絲」的人；改為「姊妹」，並刪除錯誤
   glossary 詞條及中文手冊中的同源斷言。
3. `world.tavern-tale-59` 補回原文對當地飲食的挖苦，`crown jewels` 改為
   「王室珠寶」。

## UI 抽樣

以正式 `cmd/azure-bonds-game`、Docker／Xvfb、640×480 擷取並人工檢視：

- AREA 地圖；
- 世界地圖；
- Fireball 命中關鍵幀；
- 原版順序建角後的隊伍組裝頁。

四張均已收入 `docs/screenshots/manifest.json`。世界地圖日期原使用 156px 安全框，
繁中只顯示到「第0…」；改成右側 `(x=336, width=288)`，可容納四位數年份且仍留
8px 右邊界。`TestOverlandDateTextSafeRectangleFitsTraditionalChineseDate` 以 16px
CJK cell 明訂這項契約。

## 尚未宣稱

- 尚未完成 1,598 條逐句人工校對；本批是機械全量＋高風險人工抽樣。
- 本批當時尚未驗收 CAMP→VIEW；該缺口已由後續 spec 1246 以正常 DOS oracle
  勘誤並關閉，現行 screenshot manifest 不再有 planned 項目。
- 法術樣本只證明 Fireball 一個固定關鍵幀，不代表所有動畫相位。
- 四張畫面是 `material-exact/layout-reconstructed`，不宣稱完整畫面逐像素 exact。
