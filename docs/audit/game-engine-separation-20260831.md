# Game pack／共用 engine 分離複查（2026-08-31）

## 結論

現行分離是**依賴方向正確，但尚未完全收旂**。

- CoAB 單向依賴獨立 module
  `github.com/wicanr2/golden-box-remake-engine`，鎖定 commit
  `0f569f283f8e`；engine 可在不掛載 CoAB 工作樹的 Docker 容器中
  `go test ./...` 全數通過。
- engine 的非測試 Go 程式碼搜不到 CoAB 標題、地名、NPC 或
  `coab` 條件分支；共用戰鬥、幾何、圖形、區域地圖、亂數與音訊套件
  在 engine，CoAB 劇情、事件和翻譯仍由 game repo 消費。
- 兩個 repository 工作樹獨立；engine 不是 CoAB gitlink，CoAB `go.mod`
  也沒有提交本機 `replace`。

但仍有三項邊界債務，不能宣稱已完全作品中立：

1. **DAX codec 留在 game repo**：`internal/dax/dax.go` 是通用 SSI Gold Box
   block container parser，含 DOS RLE 與 PC-98 block decoder，卻仍屬於 CoAB module。
   engine 現無等價 DAX package。這與「DAX codec 屬共用 engine」的架構契約不符。
2. **engine 含 CoAB 劇情 example**：
   `examples/curse-of-the-azure-bonds/events/pit-of-moander.json` 含 Alias、
   Dragonbait、Zhentil、Journal 32 與作品記憶體位址。它雖不進 Go runtime，
   仍是放錯 repository 的作品 payload；應移回 CoAB 或改成 synthetic fixture。
3. **ECL VM 還沒有清楚切開共用核心與作品 adapter**：CoAB `internal/ecl`
   同時含 opcode／控制流／亂數核心，也含 CoAB 記憶體位址、block 閘門、
   世界地圖與劇情 consumer。目前無法讓下一款 Gold Box 只提供 game pack
   就重用完整 ECL runtime。

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

1. 用 CoAB 與 Pool 真實 DAX 檔案同時定義 engine codec contract，保留 malformed
   input 的 fail-closed 測試。
2. 把 ECL decode／opcode／VM 控制流與作品 memory map／text rule／事件 consumer
   分成單向 interface；不允許 engine 分支比對遊戲 ID。
3. 移除或中立化 engine 的 CoAB example，再加入 Pool 作為第二個真實
   consumer。只有兩作都不需要 title branch 才可宣稱 API 真正可重用。

## 本次可重現收據

- engine HEAD：`0f569f283f8e`。
- 容器：`coab-go-test:20260729`，`--network none`，不掛載 CoAB。
- 命令：`go test ./...`，engine 全套件通過。
- 靜態搜尋：engine 的 `*.go` 對 CoAB 標題／地名／NPC／`coab` 命中 0；
  全 repository 命中只在文件、knowledge 與上述 example。
