# 《青色枷的詛咒》Remake 發行包

本程式是繁體中文 remake，不包含原版遊戲的公開散布授權。

由本專案權利人擁有的程式碼、翻譯、文件與原創內容採 PolyForm Noncommercial
License 1.0.0；允許非商業使用、修改及散布，商業使用須另行取得授權。完整條款
與第三方素材排除以發行包內的 `LICENSE`／`NOTICE.md` 為準。

## 補丁包（patch）

把自己合法持有的 DOS 遊戲檔打包成 `curseoftheazurebonds.zip`，放在執行檔或
AppImage 同一層；ZIP 內需保留原始檔名。補丁包不含該 ZIP，也不含尚未確認
公開散布權的 PC-98 音樂 OGG。

## 本機完整包（full-local）

建置工具可為開發者在本機產生含原版 ZIP 與本機音樂的完整包，僅供自己驗收；
請勿上傳、提交 Git 或再散布。

存檔預設寫到作業系統的使用者設定目錄
`curse-of-the-azure-bonds-remake/party.json`，不會寫進唯讀的 AppImage 或 `.app`。
可用 `-party-save <路徑>` 明確改成其他位置。

## 發行驗收

Linux full-local 可用 `tools/linux-release-smoke.sh <版本>` 在 Docker／Xvfb
中啟動並截圖。Windows full-local 可用
`tools/windows-release-smoke.sh <版本>` 在專案的 Wine／Xvfb image 驗收。
Wine 是交叉平台啟動證據，不取代 Windows 真機驗收。
