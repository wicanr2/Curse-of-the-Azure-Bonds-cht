# 第五百三十一輪：角色建立原版石框與繁中字級

狀態：`READY`（角色建立畫面的版面契約；不代表完整角色建立或整作通關）

## 本輪結論

角色建立畫面現在遵守與原版冒險選單相同的 640×480 畫面契約：先繪製本機抽出的
原版裂紋石框，再在內框中使用倚天粗體 16×15 字級。畫面文字仍由 locale／game
pack 解析，沒有把角色、選項或繁中文字串重新寫回 Go renderer。

本輪修正位於 `cmd/azure-bonds-game/main.go` 的 `drawCreation`：

- 使用 `drawOriginalAdventureFrame(screen)`，不再以沒有原版石框的純色畫布呈現。
- 角色建立的標題、提示、選項、能力值與操作說明統一使用 `compactFace`。
- 保留原版選項密度；24px 顯示字型仍可供短標題或其他畫面使用，但不套用到
  角色建立的密集清單。

## 可重現證據

### Remake runtime

在 Docker 內以 `rich2-go-ebiten:latest` 編譯後，使用 `-opening` 建立開場 State，
待遊戲視窗出現後把 `C` 鍵送到實際遊戲視窗。這會走正常的前端輸入分支
`State.OpenCharacterCreation()`，不是直接注入 `ModeCharacterCreation` 或座標。

輸出截圖：
[`docs/screenshots/character-creation-remake-640.png`](../screenshots/character-creation-remake-640.png)

| 欄位 | 值 |
| --- | --- |
| 畫面尺寸 | `640×480` |
| 證據分類 | `layout-reconstructed` |
| 執行環境 | Docker、`Xvfb :99`、`rich2-go-ebiten:latest`、倚天 `stdfont.15` |
| 啟動流程 | `-opening` → 實際 `C` 鍵 → 角色建立 |
| 截圖 SHA-256 | `f6ef2f54cf485610a8d62f64b007d06a4e87935e3d76991f22fbfe385abcb6a3` |

### 原始素材與版面邊界

- DOS 低解析素材與石框仍以原始映像為來源；本輪沒有改寫原始映像。
- `ExtendedAdventureFrame` 的原版裂紋石框是素材證據；640×480 延伸畫布的下方
  命令區與中文字行距屬於重製版版面重建，不能宣稱逐像素 DOS exact。
- 截圖只證明角色建立的 renderer 已使用正確舞台、框線與字級；不證明完整
  copy-protection、所有角色建立輸入、存檔相容性或正常路徑已完成。

## 驗證

Docker 內通過：

```text
go test ./gamepack ./internal/game ./internal/party
go build ./cmd/azure-bonds-game
```

編譯使用同 workspace 的暫時 local `replace` 指向
`golden-box-remake-engine/`，只存在於 ignored `workplace/coab-test.mod`，正式
`go.mod` 與兩個 repository 的邊界沒有變更。

## 尚未完成

角色建立的刪除／改名／多職業完整規則、正常 copy-protection 後的完整玩家路徑、
全畫面逐張原版比對，以及整作的 ECL、地圖、戰鬥、音效、存檔、翻譯與三平台
發行仍依 `WORKLIST.md` 與 `docs/knowledge/coab-re-coverage-matrix.md` 的現行閘門進行。
