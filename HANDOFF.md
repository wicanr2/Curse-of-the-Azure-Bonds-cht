# CoAB remake 最短交接

更新：2026-08-30。這是 compact／接手後的第一份狀態摘要；規則仍以
[`AGENTS.md`](AGENTS.md) 為準，深層現況才讀 [`CONTEXT.md`](CONTEXT.md)。

## 現況

- 目標：Go／Ebiten 繁中 remake，玩家可見體驗近似原版 99%；採風險導向代表性抽樣。
- 主線：規則層與連續鍵盤輸入皆可由開場到結局；落回原文 0。
- 一般強度：pending Bless 死鎖已修；2,455 幀後真實全滅是合法停止，不要求該樣本通關。
- 第一人稱：585 種牆面配置、1,498 張原版對照全部逐格相同。
- 音訊：runtime 正式資產為 12 首音樂＋9 個音效 OGG；仍待人耳 loop 與散布權確認。
- 發行：四個目標平台包可重生；Linux 與 Wine smoke 已做，Windows／macOS 真機尚待。
- 圖像：1,741 張 PNG 與 manifest 已涵蓋 sprite、TILES、戰場地形、AREA／共用
  符號、SKY 與牆片；runtime 不再從原版 DAX 解碼圖像。移除原版 ZIP 中 51 個
  圖像 member、保留 43 個非圖像 member 的 fixture，已抽樣通過開場 PIC、AREA、
  第一人稱、戰鬥與 BIGPIC，其中三個可直接比對區域為 0 像素差。
- 建隊／建角：正常入口依原版走種族→性別→職業→陣營→能力值／HP→姓名→
  READY／ACTION 圖示→存檔；圖示可調頭、武器、體型與六部位雙色。原版及手繪
  theme 均套用同一組持久色值，手繪素材可由 CHEAD／CBODY 任意組合後調色。
- Polish：已依「機械全量＋風險導向人工抽樣」完成；1,598 條中譯的三項缺口為 0，
  138 個詞彙問題 0。三批人工樣本涵蓋後期 Journal、全酒館傳聞、物品、商店／神殿／
  訓練、法術與結局，並修正系統性省句、混種劍跨畫面異名及護手結局指涉。UI manifest
  現有 12 張、planned 0；AREA／世界地圖／Fireball／建隊／角色 VIEW／最長結局頁均已
  人工檢視，日期與結局長文截斷已修。這不等於可選的 1,598 條出版級逐句編校。

## 下一個最小動作

依 [`docs/knowledge/remake-completeness-assessment.md`](docs/knowledge/remake-completeness-assessment.md)：

1. 把無圖像原版資料 fixture 納入發行封包抽樣，確認各平台包只需原版非圖像資料。
2. 抽樣主線、主要替代入口、前中後期高風險支線與曾出錯分支。
3. 逐首做人耳 OGG loop 驗收，並完成可散布權利清單。
4. 完成 Windows／macOS 真機、macOS 簽署／公證。

## 不可重開的舊 gate

- 不追求所有狀態逐像素、逐分支、逐週期全量閉合。
- 不把一般強度真實全滅改造成 forced win，也不暗中強化隊伍。
- `WORKLIST.md`、`docs/knowledge/coab-remake-todo.md`、`PLAN.md` 都是歷史台帳。
- `docs/project-status.md`、`GOLDEN_BOX_RE.md` 與舊根目錄知識路由已移除。

## 跨作品邊界

- CoAB 劇情、座標、旗標、翻譯、資產索引與存檔版本只能留在本作品 game pack／證據。
- 可重用候選與升格條件見
  [`gold-box-series-reuse.md`](docs/knowledge/gold-box-series-reuse.md)。
- 第二款合法持有的 Gold Box 作品通過格式、consumer 與最小 adapter 前，任何
  「Gold Box 共通」都只能標為候選或 strong inference。

## 驗證入口

- 可重生狀態：`go run ./cmd/remake-status -output docs/audit/remake-status.md`
- 函式完成／無直接 caller：`python scripts/function_completion_audit.py`；現行待解讀 0、
  無解釋 caller=0 為 0。
- 代表性測試：`go test ./gamepack ./internal/game ./cmd/azure-bonds-game -count=1`
- 所有工作負載一律經專案 Docker 工具鏈；結束時確認沒有殘留容器。
