# 第五百一十一輪：提爾佛頓設施正常移動路徑與事件群組邊界

狀態：`READY`（有界正常路徑規格，不代表完整遊戲完成）
日期：2026-08-09

## 目的

延續第 510 輪的 `State.MoveDungeon` 交易，將正式新遊戲在提爾佛頓的設施之間
移動改成同一個正常輸入入口，並記錄原始 ECL 的 one-shot 事件群組。這輪同時
修正測試解讀：不能因為座標抵達了高階祭司，就忽略先前招牌事件已消耗同一個
原始事件群組。

## 證據與位址空間

| 證據 | 觀察 | 等級 |
|---|---|---|
| 原始 `GEO2.DAX` block 1 | 開場中斷休息後的 `(4,13)` 可沿原始牆／門資料步行至 Filani `(6,5)`、Weaponers `(2,12)`、Gond altar `(0,7)`、Training Hall `(5,2)`、Tavern `(6,10)` 與高階祭司 `(1,10)` | `exact`（GEO decoded bytes／geometry 位址空間） |
| `State.MoveDungeon` | 每一步驗證 cardinal delta、GEO 可行性、座標投影、wall／roof、`7F81h` 清除與同一 `BlockSession` continuation | `strong inference`（依 ECL register／前端順序；尚非 DOS 逐幀 exact） |
| 原始 `ECL2.DAX` block 1 `+0286h` | SearchLocation 以 `C04F & 0x7F`、表格查詢後得到 `7F7Ah`，再依 selector 分派 | `exact`（ECL bytes／payload offset 位址空間） |
| 原始 `ECL2.DAX` block 1 `+1104h..+110Fh` | 高階祭司入口先執行 `AND 4C03h, 7F7Ah, 7F79h`；若結果非零即 `EXIT` | `exact`（ECL bytes／work address 位址空間） |
| 正常路徑 runtime | 先經 terrain `0x0D` 招牌後，`4C03h` 變成 `0x80`；抵達 terrain `0x8F` 時同一群組使高階祭司重訪保持靜默 | `exact`（remake ECL runtime；原版 DOS 逐幀仍未宣稱） |
| 高階祭司 fresh session | 不先消耗招牌群組時，terrain `0x8F` 仍產生 PICTURE 6、HEAD6／BODY6、YES／NO、Remove Curse、Journal 19 與返回原格 | `exact`（remake 原始 ECL integration；不是 pixel／instruction exact） |

`ECL work address`、GEO 座標、ECL payload offset 與檔案 offset 是不同位址空間，
本規格不把相同數字合併成同一地址。`4C03h` 的「共用 0x80 one-shot 群組」是由
AND／EXIT 控制流與 runtime 前後狀態支持的 `strong inference`；原始 flag 的跨存檔
格式與所有其他作品的通用語意仍不能由本輪擴大推論。

## 正常移動路徑

`TestRealNewGameBeginsAtGlobalBlockOne` 現在沿同一個 state-owned movement
transaction 驗證下列段落：

```text
開場中斷休息後 (4,13)
  → (6,5) Filani：途中顯示招牌／下水道傳聞 pause，抵達後正常續跑
  → (2,12) Weaponers of Cormyr
  → (0,7) Gond altar
  → (5,2) Training Hall
  → (6,10) Tavern
  → (1,10) high-priest cell，轉向北方後重訪保持靜默
```

招牌與傳聞新增的繁中顯示只存在 `assets/locale/zh-TW.json`，State 以穩定 locale
ID 解析，不在 movement transaction 或測試中塞入劇情中文。高階祭司的獨立分支
使用 fresh ECL session 驗證，避免測試為了強行同時看到兩個互斥 one-shot 事件而
清除 `4C03h` 或竄改 ECL 記憶體；這不是玩家流程 shortcut 被包裝成正常路徑。

## renderer／輸入契約

- `TurnDungeonWithGrid` 只改變面向並重新投影目前 geometry cell 的 wall／roof；
  轉向不自行執行 ECL，後續 movement／SEARCH 才消費該狀態。
- Ebiten 的 movement 與 turn frontend 會呼叫 State 交易；renderer 只在成功後
  refresh，不再另行改寫 `DungeonX`／`DungeonY` 或重複 lifecycle。
- known pause 只接受 game-pack／locale 定義的招牌與傳聞訊息；未知事件會使測試
  失敗，不得用「按一下繼續」吞掉新的 ECL boundary。

## 驗證

Focused Docker regression：

```text
GOWORK=/tmp/go.work GOMODCACHE=/cache/mod GOCACHE=/cache/build GOPROXY=off \
go test -buildvcs=false -count=1 ./internal/game \
  -run TestRealNewGameBeginsAtGlobalBlockOne -v
```

本輪完成後仍須跑 `AGENTS.md` 規定的 Xvfb formal gate、`coab-audit`、
`git diff --check` 與 Docker 清理檢查。上述路徑只提升提爾佛頓這段的證據等級，
不能宣稱完整 ECL、全地圖、全戰鬥、完整中文化或開場到結局已完成。

## 未完成邊界

- 高階祭司與招牌共享群組的完整 DOS runtime／畫面逐幀對照尚未保存。
- 提爾佛頓後續馬車、盜賊公會、下水道與火刀路徑仍有部分 coordinate-assisted
  測試；它們必須逐段改成正常 movement／event continuation。
- 完整 ECL external routine、戰鬥動畫／音效、save、全翻譯與通關驗收仍未完成。
