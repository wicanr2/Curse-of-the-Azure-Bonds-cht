# Game pack／共用 engine 分離複查（2026-08-31）

## 結論

現行 game／engine 已完成實體、版本庫與程式邊界分離。

- CoAB 單向依賴獨立 module
  `github.com/wicanr2/golden-box-remake-engine`，鎖定 commit
  `0f569f283f8e`；engine 可在不掛載 CoAB 工作樹的 Docker 容器中
  `go test ./...` 全數通過。
- engine 的非測試 Go 程式碼搜不到 CoAB 標題、地名、NPC 或
  `coab` 條件分支；共用戰鬥、幾何、圖形、區域地圖、亂數與音訊套件
  在 engine，CoAB 劇情、事件和翻譯仍由 game repo 消費。
- 兩個 repository 工作樹獨立；engine 不是 CoAB gitlink，CoAB `go.mod`
  也沒有提交本機 `replace`。

本輪已關閉先前的三項邊界債務：

1. DAX DOS／PC-98 codec 與 malformed-input 測試已移至 engine `dax` package；
   CoAB 全部 consumer 直接匯入 engine module。
2. CoAB 的 Moander 劇情 example 已移回 game repo
   `examples/engine-pack/events/`；engine 只留 synthetic fixture。
3. ECL operand、instruction framing、變長 record、branch table 與控制流圖已移至
   engine `ecl` package；CoAB `internal/ecl` 只保留作品 memory、文字、事件效果與
   UI adapter。
4. engine 工作副本固定在 `../golden-box-remake-engine/`；CoAB 不再忽略或容納
   nested clone，bootstrap／proxy 工具也只使用同層 repo。

## 不應誤判為邊界漏洞的項目

- `cmd/azure-bonds-game` 的 Tilverton、Dracandros、Tyranthraxus 預覽旗標是
  CoAB 專屬前端與擷圖入口，留在 game repo 位置正確。
- `internal/game` 的 CoAB 地點 enum、起點與 ECL 狀態 adapter 是作品層；
  其中的通用機制若要提取，應先由第二款真實 game pack 證明 API 形狀，
  不以改名就冒充 engine。
- engine `docs/knowledge/` 可記錄「這個契約首先由 CoAB bytes 證實」；
  證據來源不等於 runtime 硬編碼。但 archived 文件中的舊地名／座標不可當成
  共用規格。

## 下一款 Gold Box 的使用門槛

Pool of Radiance 不得複製 CoAB `internal/dax`或整包 `internal/ecl`。
開工順序應為：

1. 直接消費 engine `dax`，再用 Pool 真實 DAX 樣本補第二作品的 count、offset、
   stride、malformed-input 與 consumer 驗證；不得另寫一份 parser。
2. 直接消費 engine `ecl` 的 instruction／control-flow API；Pool 自己提供 memory
   map、text rule 與 event consumer，不允許 engine 分支比對遊戲 ID。
3. Pool 成為第二個真實 consumer 後，重新審查目前仍標為跨作品候選的 API；只有
   兩作都不需要 title branch 才可升格為已證明可重用契約。

## 本次可重現收據

- engine HEAD：`86ac57e498c9`。
- 容器：`coab-go-test:20260729`，`--network none`，不掛載 CoAB。
- 命令：`go test ./...`，engine 全套件通過。
- 靜態搜尋：engine 的 `*.go` 對 CoAB 標題／地名／NPC／`coab` 命中 0；
  全 repository 命中只在文件、knowledge 與上述 example。
