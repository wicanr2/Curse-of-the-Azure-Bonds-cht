# 第一百五十九輪：State LOAD PIECES request

## 實作結果

第 158 輪的 ECL `LoadPiecesRequested` signal 現在由 `game.State.Select` 保存為 `State.LoadPieces`，並透過 `ConsumeLoadPiecesRequest()` 一次性轉交未來的 map-piece loader。這與既有 `ConsumeGeoMapRequest()` 的 renderer-neutral request boundary 一致；重複消費不會再次觸發載入。

## 明確 boundary

本輪不解讀三個 selector 的檔案名稱、順序或 floor／wall／tile 寫入規則，也不改變目前 GEO renderer；DUNGCOM／8X8D／WALLDEF／TILES 的實際 adapter 仍需各章節反組譯證據。
