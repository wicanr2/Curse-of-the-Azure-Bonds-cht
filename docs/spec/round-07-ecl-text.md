# 第七輪：ECL packed text extraction

狀態：`DRAFT`

`internal/ecl.DecodePackedText` 與 `FindPackedTextCandidates` 已將 ECL 內觀察到的 `0x80 + length + 6-bit payload` 轉為原文候選。此輸出可作為翻譯資源的初始輸入，但不直接當成完整字串表：同樣的 marker 在 GEO／圖片資料中會造成誤判。

## 驗收

```sh
go test ./...
go run ./cmd/azure-bonds -image curseoftheazurebonds.zip -member ECL1.DAX -strings
```

下一步是以 ECL command context、文字重複位置與英文手冊對照去重，再建立 `en`／`zh-TW` stable keys。
