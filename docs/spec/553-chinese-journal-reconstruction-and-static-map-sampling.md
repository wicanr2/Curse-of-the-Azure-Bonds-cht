# 第五百五十三輪：中文冒險手札重建與靜態地圖抽樣準則

狀態：`READY`（限手札 1–59 的可追溯摘要與地圖研究方法；不是全量遊戲內接線）

## 輸入與方法

- `珍020-青色枷的詛咒.rar`：中文掃描本；原檔與掃描頁不納入 Git。
- `Curse-of-the-Azure-Bonds_Misc_DOS_EN_Adventurers-Journal.pdf`：英文交叉校訂。
- `workplace/journal-ocr-406.txt` 與本輪 `/tmp` OCR：只作定位，不作文字權威。
- 手冊抽取與 OCR 均在 Docker、原始資料唯讀掛載下執行。

版本化成果是 `docs/manual/adventurers-journal-zh-TW.md` 與三份分冊。文字採摘要／
重述，保留條目編號、來源頁、提示與 confidence，不逐字散布掃描本。

## 手札 59 勘誤

中文 `084.jpg`／印刷頁 52 確實保存完整地圖與圖例。先前「只有英文 OCR 圖例、
沒有幾何圖像」只適用於 repo 當時已版控的 OCR，不代表原始中文掃描缺圖，現由本
規格 supersede。

地圖標籤與圖例是 `exact`；房間相對拓撲為 `layout-only`；GEO 座標、牆 byte、
`SEARCH／LOOK` 與出口副作用仍是 `unknown`，不可由圖面單獨升級。

## 眼魔洞穴的靜態證據閉環

既有原始資料已分別證明：

- 手札取得點：正常 session 到 `GEO4/0x25 (13,1,W)`。
- 德克薩姆固定遭遇：terrain `90h`／`8Fh`，局部座標 `(15,1)`／`(15,2)`。
- 洞穴出口事件：terrain `93h`，局部座標 `(6,3)`。

缺口不是事件內容未解，而是 `(13,1)` 到德克薩姆區、再到 `(6,3)` 的普通玩家
移動／互動 route 尚未閉合。下一輪先做手札房間圖與 GEO cell graph 的約束對位，
再查 ECL producer／consumer。只有仍有多個候選解時才啟動原版受控抽樣。

## 受控原版抽樣

DOSBox 不要求從新遊戲玩到深處。以來源存檔的唯讀副本、已驗證的 save parser／
非破壞性 patch，將隊伍放在候選邊前一格；保存 before／after 雜湊、座標、朝向、
按鍵、畫面、旗標與重載結果。此方法只驗證候選互動，不可冒充正常完整通關。

## 驗收與限制

- 1–59 均有索引；手札 1 明確標為未找到，不虛構。
- 手札 59 的標籤、圖例與來源頁可回查。
- 文件明確區分手札拓撲、GEO bytes、ECL 行為與 DOS runtime。
- 本輪未更改 engine、ECL runtime、遊戲規則或原始掃描。
