# 1188 — 玩家路徑：把戰役檢查點交給真的前端畫一張

- 證據等級：`exact`（`cmd/azure-bonds-game` 實跑產生的 PNG）
- 上游：`TestRealNewGameRunsToTheEnding`（`internal/game/campaign_normal_test.go`）
- 產物：`tools/campaign-frames.sh`、`cmd/campaign-frame-audit`、
  `docs/audit/campaign-frames.md`
- 狀態：`READY`

## 問題

`remake-status` 的「開場到結局的同一 session」那一列寫著：

> 端到端測試跑得到結局，但那是**測試路徑**；真人從開場玩到結局沒有自動化的證明。

那句話是準確的。`TestRealNewGameRunsToTheEnding` 從角色建立一路打到提朗瑟克斯，
24 個劇情段全綠，還附帶語系、隊伍連續性、攜帶物、存檔往返等檢查——但它驅動的是
`*State`，**畫面那一層一次都沒被跑到**。玩家執行的是 `cmd/azure-bonds-game`。

## 做法

戰役測試每一段結束時本來就會存一份快照（段界往返驗證用），只是寫到
`t.TempDir()`。加一個環境變數把它們導出來，再讓真的前端逐一載入並畫一張：

```
COAB_CAMPAIGN_SNAPSHOT_DIR=... go test -run TestRealNewGameRunsToTheEnding
cmd/azure-bonds-game -eten-font ... -party-load <快照> -screenshot <png>
```

⚠ 字型在 repo 外（`/home/anr2/cht/etan_font`），要另外唯讀掛進容器。
**少了它每一個字都是豆腐框，而畫面其他部分完全正常**——只看「有沒有產生 PNG」
驗不出來，第一次跑就踩到了。

## 結果

23 個檢查點全部畫得出來，沒有一張空白（`docs/audit/campaign-frames.md`）。
模式分佈：地城 13、荒野 8、事件 1、戰鬥 1。

⚠ **「畫得出來」不等於「畫對了」。** 像素統計驗得到「不是空白」，驗不到字型
有沒有載入，也驗不到畫的是不是那一段該有的畫面。

### 一個**不是**缺陷的觀察

三對檢查點畫出來**位元組完全相同**。其中兩對的存檔在 `MapX`／`MapY`、朝向、
`InDungeon` 上都不一樣，看起來像「畫面沒跟著位置變」。

實際上世界地圖那一張只由 `Area.CurrentCity` 決定（`drawOverlandMap` 標的是
目前城市，不是座標），而那兩對的 `CurrentCity` 相同 ⇒ **畫面本來就該一樣**。
`cmd/campaign-frame-audit` 因此把每組的 `CurrentCity` 一起印出來，
免得下一個人看到重複雜湊就當成缺陷。

## 找到的兩個缺陷（都已修）

兩個都在 `LoadPartyFile`，都是同一類：**存檔保存了狀態，但沒有保存畫面上的字，
而讀檔路徑用建檔當下的值把它們填掉。**

### 一、讀回來顯示的是開場選單

`LoadPartyFile` 最後**無條件**設世界地圖 hub：

```go
s.Prompt = s.catalog.Text("party_ready", ...)          // 「隊伍已建立．準備開始冒險．」
s.Choices = []string{"進入城市", "繼續旅程", "紮營"}
```

於是存在事件畫面上的檔讀回來，玩家人在密斯卓諾墓園，畫面卻是剛建好隊伍的
世界地圖選單。F5 快速存檔**沒有限制模式**，所以這是玩家走得到的。

修法：`ModeEvent` 改成給「請按 Enter 繼續」並清掉選單；其餘模式不動。

### 二、地名永遠是建檔當下的那一個

`Location` 有還原，**`LocationName`（畫面標題那一行）沒有**。原本的還原路徑走
`setWorldLocation(Area.CurrentCity)`，而那一支只在世界地圖模式跑——理由寫在
原本的註解裡：地城與戰鬥存檔的 `CurrentCity` 可能是舊值或 0，拿它去覆寫會把
有意義的 `Location` 蓋掉。

修法：加一份 `Location` → 語系鍵的對照，由**已經還原好的 `Location`** 推出
顯示名稱。⚠ 這一支**只寫 `LocationName`、不碰 `Location`**，所以原註解講的
hazard 在結構上不成立。

### 為什麼 `*State` 層的測試看不到這兩個

讀完檔就直接做下一個動作，`Prompt`／`Choices`／`LocationName` 在被看到之前
就被覆蓋掉了。**這兩格只有「畫出來」才會被觀察到**——而在這一輪之前，
從來沒有人把戰役的狀態畫出來過。

★ 同一個函式裡上面幾行就有一段 `eventReturnMode` 的註解，講的是同一類問題
（存檔沒有保存畫面上的過場狀態），當時只補了「按下一步會不會卡住」，
沒有補「畫面對不對」。**同一類缺口補了一半會看起來像補完了。**

## 這一份不宣稱什麼

- **沒有**宣稱真人玩得完。這裡證明的是「戰役路線上的每個檢查點，真的前端載入
  得起來並畫得出非空白畫面」，不是「操作流程全部走得通」。輸入那一層
  （鍵盤 → `Update`）仍然沒有被自動化驗過。
- **沒有**宣稱畫面內容正確。像素分佈只驗得到「不是空白」。
- 檢查點是**段界**，不是玩家會存檔的地方；兩者不一定重合。
