# 第一百二十輪：CAMP SAVE request adapter

狀態：`READY`（限 remake versioned party save 的 request／adapter boundary）

## 實作 contract

`CAMP → SAVE` 在 `game.State` 只產生一次性的 `ConsumeSaveRequest()` signal；當 roster 存在時，Ebiten adapter 以既有 `SavePartyFile` 寫入 configured party path，並把成功或錯誤訊息顯示在事件畫面。state core 不選擇 filesystem path，也不直接執行 I/O。無 party 時不產生 request。

這個 boundary 重用目前 versioned remake party JSON；它不宣稱已解析原版 `SAVGAM?.DAT` slot／area container，也不取代 F5 快捷鍵。

## 驗證

`TestCampMenuSaveEmitsRequest` 驗證 request 只可 consume 一次，並確認事件完成後返回 CAMP Menu；全套 `go test ./...` 已通過。
