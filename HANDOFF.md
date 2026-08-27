# CoAB remake 最短交接

更新：2026-08-27。這是 compact／接手後的第一份狀態摘要；規則仍以
[`AGENTS.md`](AGENTS.md) 為準，深層現況才讀 [`CONTEXT.md`](CONTEXT.md)。

## 現況

- 目標：Go／Ebiten 繁中 remake，玩家可見體驗近似原版 99%；採風險導向代表性抽樣。
- 主線：規則層與連續鍵盤輸入皆可由開場到結局；落回原文 0。
- 一般強度：pending Bless 死鎖已修；2,455 幀後真實全滅是合法停止，不要求該樣本通關。
- 第一人稱：585 種牆面配置、1,498 張原版對照全部逐格相同。
- 音訊：runtime 正式資產為 12 首音樂＋9 個音效 OGG；仍待人耳 loop 與散布權確認。
- 發行：四個目標平台包可重生；Linux 與 Wine smoke 已做，Windows／macOS 真機尚待。
- 圖像：780 張 sprite PNG 已由 runtime 使用；TILES、戰場地形、AREA／共用符號、
  SKY、牆片六類仍從原版 DAX 載入，是目前最大 release 整合項。

## 下一個最小動作

依 [`docs/knowledge/remake-completeness-assessment.md`](docs/knowledge/remake-completeness-assessment.md)：

1. 把六類 runtime 圖像依 [`PNG 獨立化盤點`](docs/audit/png-asset-independence-2026-08-27.md)
   匯成 PNG＋必要 manifest；先做缺少圖像 DAX member 仍能啟動的測試。
2. 抽樣主線、主要替代入口、前中後期高風險支線與曾出錯分支。
3. 抽樣選單、戰鬥、PIC、AREA、結局的原版／remake 同狀態畫面。
4. 完成 Windows／macOS 真機、macOS 簽署／公證及對外授權清單。

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
