# 第 539 輪：繁中 GUI 寬度與溢框修正

狀態：`READY`（renderer 版面契約；不是所有畫面逐像素 exact）

## 問題

README 舊截圖顯示中文訊息、地點名與選項以英文固定字元數換行，長字串會突出
右側石框或覆蓋下一個欄位。這不是把字體縮小就能可靠解決的問題：倚天粗體的
實際 glyph advance、標點寬度與不同字形 fallback 都可能不同。

## 修正契約

- logical canvas 固定為 `640×480`，正文與 status 使用倚天粗體 `16×15` 基線。
- `drawWrappedText` 改依目前 `font.Face` 的實際 glyph advance 測量換行，並以
  rune 為單位處理文字，不切斷 UTF-8／Big5 多位元組字元。
- 單行欄位使用 `drawFittedText`，超出明確區域時以 rune-safe 截斷並顯示省略號。
- 冒險標題／地點／事件訊息／選項、地城第一人稱、手札、角色建立與戰鬥 HUD
  都傳入各自的最大寬度；不再以固定英文字元數假設中文也能放下。
- 石框、人物 HEAD／BODY 舞台與左上場景的 renderer contract 不改成 generic
  灰框；本輪只修文字佈局，原版忠實 theme 與日後美化 theme 仍分離。

## 畫面證據

使用 Docker／Xvfb 與本機倚天字型重新產生並檢查 640×480 畫面：

- [`gold-box-layout-adventure.png`](../screenshots/gold-box-layout-adventure.png)
  （目前冒險／地城文字版面）
- [`tilverton-inn.png`](../screenshots/tilverton-inn.png)
  （提爾佛頓設施與中文訊息）
- [`tilverton-first-person-remake.png`](../screenshots/tilverton-first-person-remake.png)
  （第一人稱場景與右側 party/status）

目前證據等級是 `layout-reconstructed`：可確認文字留在指定區域、石框與欄位
關係沒有被中文撐破；尚未宣稱與同一原版狀態逐像素相等，也尚未完成所有戰鬥、
地圖、對話、頭像與手札頁面的逐張 fidelity 稽核。

## 測試

`cmd/azure-bonds-game` 以 basic font face 驗證實際字寬換行與單行裁切；Docker／
Xvfb 代表性套件測試通過：

```text
go test -modfile=workplace/coab-test.mod ./cmd/azure-bonds-game ./gamepack
```

產品文字仍由 locale JSON 與 stable message ID 提供；本修正沒有把中文內容搬回
Go 測試或 renderer 常數。
