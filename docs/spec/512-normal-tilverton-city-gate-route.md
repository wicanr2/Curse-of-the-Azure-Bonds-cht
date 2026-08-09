# 第五百一十二輪：提爾佛頓城門與皇家馬車正常 GEO 路徑

狀態：`READY`（有界正常玩家路徑規格，不代表完整遊戲完成）
日期：2026-08-09

## 目的

延續第 511 輪的提爾佛頓設施路徑，移除高階祭司之後前往城門的座標注入。這輪
要求從現有地城狀態沿 decoded `GEO2.DAX` 真正逐格移動，在原始 ECL 觸發城門
封路、返回城內，再以正確面向進入皇家馬車事件與後續戰鬥。

## 證據與位址空間

| 證據 | 觀察 | 等級 |
|---|---|---|
| decoded `GEO2.DAX` block 1 | `(1,10)` 到 `(1,0)` 的 16 步路徑逐格通過雙側牆／門檢查，路徑為 `E,E,N,N,N,E,N,N,N,N,N,N,N,W,W,W` | `exact`（GEO decoded bytes／geometry 位址空間） |
| `TestRealNewGameBeginsAtGlobalBlockOne` | 高階祭司後使用 16 次 `State.MoveDungeon`，沒有寫入目標座標或直接呼叫目標事件 | `exact`（remake normal-input regression） |
| terrain／ECL continuation | 最後一步抵達 `(1,0,W)` 時顯示 game-pack stable ID `tilverton.carriage-gate-closed`；選擇後回到同一地城格 | `exact`（remake ECL runtime；尚未宣稱 DOS 逐幀 exact） |
| `TurnDungeonWithGrid` + lifecycle | 由城門西向轉為北向後，取得 `PICTURE 11`、皇家馬車訊息，接續青色枷強制攻擊與皇家衛兵戰鬥 | `exact`（已重播的 remake boundary；畫面逐像素與 timing 仍另行驗收） |
| game-pack `text_rules` | 綠袍女人傳聞以 `tilverton.green-robes-rumor` 的英文來源、繁中翻譯與片段 matcher 保存；State 不新增劇情中文 | `exact`（資料分層／locale resolver） |

`GEO` 座標、ECL work address、ECL payload offset 與檔案 offset 是不同位址空間，
本規格不把相同數字合併成同一地址。`State.MoveDungeon` 與原 DOS movement loop
的逐指令順序仍是 `strong inference`；本輪只宣稱正常輸入、事件 boundary 與續跑
結果已閉合。

## 正常玩家路徑

測試由第 511 輪正常提爾佛頓狀態繼續，使用原始 GEO 可行路徑：

```text
(1,10,N)
  → (2,10,E) → (3,10,E)
  → (3,9,N) → (3,8,N) → (3,7,N)
  → (4,7,E)
  → (4,6,N) → (4,5,N) → (4,4,N) → (4,3,N) → (4,2,N)
  → (4,1,N) → (4,0,N)
  → (3,0,W) → (2,0,W) → (1,0,W)
```

最後一步的正常 ECL boundary 是城門封路訊息。玩家按下唯一的繼續選項後，State
回到 `(1,0,W)` 的 `ModeDungeon`；接著轉向北方而不重複移動，才執行下一個
lifecycle，取得皇家馬車 `PICTURE 11`。之後同一個 ECL session 依序驗證：

1. 皇家衛兵喊話與青色枷強制攻擊。
2. `6` 名戰鬥角色、`5` 名敵方目標的皇家衛兵戰鬥。
3. 戰後紅袍人綁走假國王、投降選項與入獄。
4. 盜賊救援 `PICTURE 2`，再銜接提爾佛頓盜賊公會地城 block 2。

這段路徑的測試期待值以 `tilverton.*` stable message ID／game-pack resolver 取得，
沒有複製目前繁中顯示字串作為產品規則。

## 實作契約

- 地城移動一律經 `State.MoveDungeon`；前端不得直接改寫
  `DungeonX`／`DungeonY` 來模擬城門抵達。
- `TurnDungeonWithGrid` 只改面向與目前 geometry cell 的 wall／roof projection，
  不自行執行 ECL；轉向後的 lifecycle 才消費城門朝向。
- 城門封路是移動途中產生的 boundary，不能在路徑結束後再重播一次來製造畫面。
- 原始 ECL 的英文片段與繁中顯示由 game-pack `text_rules` 管理；State 的通用
  `MatchText` bridge 不應新增本作專用中文 fallback。

## 驗證

聚焦回歸：

```text
GOWORK=/tmp/go.work GOMODCACHE=/cache/mod GOCACHE=/cache/build GOPROXY=off \
go test -buildvcs=false -count=1 ./gamepack ./internal/game \
  -run 'Test(EmbeddedPackValidatesAndOwnsZhentilText|RealNewGameBeginsAtGlobalBlockOne)$' -v
```

提交前仍須依 `AGENTS.md` 在 Docker／Xvfb 執行 CoAB 正式套件 gate、
`coab-audit`、`git diff --check`，並檢查沒有留下專案相關執行中容器。

## 未完成邊界

- 皇家馬車後的公會內部與下水道仍有 coordinate-assisted 測試段落，尚未全部改成
  正常地圖移動。
- 城門／馬車的 DOSBox 逐幀畫面、音效、戰鬥動畫與原始 wall-clock timing 尚未
  做 pixel／instruction exact 對照。
- 全部 ECL、所有地圖、完整 AD&D 規則、完整中文化、原版 save／音樂音效與開場
  到結局的單一路徑仍未完成。
