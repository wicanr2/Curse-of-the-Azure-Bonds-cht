# 1251 — 營地選單是底部橫排指令列

- 狀態：`READY`（`CAMP`、`MAGIC`、`REST` 三層的版面）
- 證據等級：`exact`（原版 DOS 擷圖，見下節）；
  `unknown`（`ALTER` 那一層與選角色／選物品那幾層還沒擷）

## 問題

世界地圖畫面把選單畫成縱排，起點 `y=366`、每項 `+30`，而文字框的內部是
`y=264..448`。第四項落在 `y=456`——那是底框（`drawPanelFrame(8, 448, 624, 28)`）
的位置；第五項落在 `y=486`，畫布只有 480，所以**畫到外面去了**。

營地選單有七項（`SAVE VIEW MAGIC REST ALTER FIX EXIT`），法術那一層有六項。
於是原版路徑上有四項是玩家**看不到也點不到**的。沒有錯誤、沒有日誌，
只有畫面下緣一條被切掉的字。

同一個形狀在 `main.go` 另外兩處縱排也成立（沒有底圖時的野外／地點畫面自
`y=350` 起每項 `+40`；另一處自 `y=382` 起每項 `+28`），差別只在觸發它需要
幾個選項。

## 原版證據

`tools/dos-oracle-session.sh` 走正常路徑取得（載入 `savgama.dat` →
`BEGIN ADVENTURING` → `ENCAMP`）。原版把**每一層都畫成畫面最下面一行**：

| 層 | 畫面最下面那一行 | 擷圖 |
|---|---|---|
| 探索 | `AREA CAST VIEW ENCAMP SEARCH LOOK` | — |
| 營地 | `CAMP:SAVE VIEW MAGIC REST ALTER FIX EXIT` | [`camp-menu.png`](../reference/original-dos/camp/camp-menu.png) |
| 法術 | `CAST MEMORIZE SCRIBE DISPLAY REST EXIT` | [`camp-magic-menu.png`](../reference/original-dos/camp/camp-magic-menu.png) |
| 休息 | `REST DAYS HOURS MINS ADD SUBTRACT EXIT` | [`camp-rest-menu.png`](../reference/original-dos/camp/camp-rest-menu.png) |

放大那一行看得到三件事：`CAMP:` 是**洋紅**、選項是**綠色**、
**目前選項整個反白**（白底黑字），選項之間一個空白。只有營地主選單帶前綴。
七項加前綴剛好填滿 40 欄——原版是照這個寬度排的。

## 實作契約

1. `State.CampCommandRow()` 回報這一層是不是指令列，以及前綴。判斷必須
   **從最內層往外**：`enterCampMagicMenu` 會同時把 `campMenu` 設成 true，
   只看 `campMenu` 會替法術選單多畫一個前綴。
2. 選角色、選物品、選法術那幾層在原版是**清單**不是指令列，回 false，照舊逐行畫。
3. `drawCommandRow` 保證整行畫在給定寬度之內：間隔先壓縮到最少 1px，
   壓到底仍太寬才整行縮放。排版計算抽成 `commandRowLayout`（純函式），
   因為 Ebiten 的 image 在 game loop 之外讀不了像素，量不到實際畫出來的結果。
4. 目前選項用反白（白底黑字），不是加游標符號——原版就是反白。
5. 位置走版面自己的底框：世界地圖畫面用 `overlandCommandBaseline`（468，
   落在 `y=448..476` 之內），不要用 `adventureCommandBaseline`（478，那是地城
   畫面 `y=464..479` 的框，在這裡會貼上畫布底邊）。`modern-a6` 主題下
   `safeBottomBaseline` 會再把它收到 `454 - descent - 3`，避開下方雕帶。
6. 輸入不用改：`Update` 早就把左右鍵與上下鍵接成同一組游標移動。

## 驗證

- `TestCampCommandRowMatchesTheOriginalMenus`：三層都回指令列、只有主選單帶
  前綴、選角色那一層不是指令列、營地選單是七項。
- `TestCommandRowLayoutNeverExceedsItsWidth`：七項在 40～1200px 的可用寬度下
  都不超出（或自己標了要縮放）。
- `TestCommandRowLayoutFitsTheOriginalCampMenuWithoutScaling` 是正對照：
  remake 底部的 592px 放得下原版七項且**不必縮放、也不必壓縮間隔**。
  少了它，「不超出寬度」可以靠「一律縮放」蒙過去。
- remake 實機擷圖（`-camp-row camp|magic|rest -screenshot`）：
  [`coab-camp-command-row.png`](../screenshots/coab-camp-command-row.png)、
  [`coab-camp-magic-command-row.png`](../screenshots/coab-camp-magic-command-row.png)。
  七項與六項都完整可見。

## 還沒做

- **營地的畫面本身還是世界地圖。** 原版紮營時左上換成夜營圖（帳篷與營火）、
  右上是隊伍的 NAME／AC／HP 欄、狀態列多一個 `CAMPING`，訊息框寫
  `THE PARTY MAKES CAMP...`。目前只換了選單的排法。
- `ALTER` 那一層與選角色／選物品／選法術那幾層的原版版面還沒擷，
  所以那幾層維持縱排。它們的項數目前都在三項以內，不會溢出，但沒有原版證據。
- 另外兩處縱排（野外／地點畫面、`y=382` 那處）仍沒有溢出保護。
  營地選單走不到它們，但選項一多就會重演同一個形狀。
- `en.json` 與 `ja.json` 的 `camp_menu_prompt` 是機翻壞掉的
  （`Ban the menu.`／`メニューを禁止します。`），與本規格無關但同一個畫面看得到。
