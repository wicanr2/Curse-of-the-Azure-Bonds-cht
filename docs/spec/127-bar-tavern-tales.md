# 第一百二十七輪：城市 BAR Tavern Tales

狀態：`READY`（限 BAR menu、已翻譯 Tavern Tale sequence 與場所返回 boundary）

## 證據

原始 ECL1 的 packed text 已讀到 `HAVE A DRINK`、`YOU OVERHEAR TAVERN TALE`、`WHAT WILL YOU DRINK?` 與 `WILL YOU DRINK?` 等 BAR 分支；RuleBook 說明酒館提供 gossip、stories、information，並要求買酒後聽故事。Adventure Journal 的 `TAVERN TALES` 區段提供編號傳聞，且提醒傳聞可能真假不同、應依遊戲觸發順序閱讀。

本輪先接入不猜價格與觸發條件的窄 boundary：場所進入 BAR 後可閱讀由資料層提供的順序傳聞，讀完回到 BAR menu，再可返回城市場所選單。預設前六則是 Adventure Journal 的繁中整理；後續 ECL decoder 可用 `SetBarTales` 替換每座城市／劇情的正確 sequence。

## 實作 contract

- `ModePlace` 選擇 `BAR` 會進入繁中 `聽酒館傳聞／離開酒館` menu。
- `BAR_LISTEN` 每次只消費下一則 injected tale，建立 `ModeEvent`，按 Enter 回到 BAR menu；沒有 tale 時只顯示明確訊息，不修改 party。
- `BAR_EXIT` 返回城市 `INN／STORE／BAR／LEAVE` 場所選單；荒野不能誤用此返回路徑。
- `SetBarTales` 只保存 script/data layer 的順序；買酒價格、金幣扣除、城市條件、重複傳聞與原版 random trigger 仍待 ECL／資料反組譯。

## 驗證

`TestBarMenuReadsTavernTalesAndReturnsToPlaces` 覆蓋 BAR 進入、逐則閱讀、tale index、回到 BAR menu、離開與城市場所返回。Docker 內 `go test ./...` 與 locale JSON parse 已通過。
