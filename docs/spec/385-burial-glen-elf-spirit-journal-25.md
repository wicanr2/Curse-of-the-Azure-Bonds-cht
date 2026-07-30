# 第 385 輪：Burial Glen 精靈幽魂與手札 25

狀態：`READY`

## 問題

第 384 輪只證明玩家能由世界旅行進入 ECL6／GEO6 block `0x40`，但文字
畫面結束後仍沿用舊的 `(7,13,N)`，沒有接受新 ECL 寫入的出生點；因此尚不能
由正常地城移動觸發 Burial Glen 事件。

## 原始證據

- ECL6 block `0x40` initial entry `+0x0014`：
  - 只在 previous ECL `4BF2h == 50h` 時寫入 `C04B=2`、`C04C=15`、
    `C04D=1`，即 `(2,15,E)`；
  - `LOAD FILES 40,02,FF`、`LOAD PIECES 17,18,16`；
  - 顯示龍盔指出 Tyranthraxus 在北方。
- GEO6 block `0x40`：
  - `(2,15)` 向東至 `(3,15)` 可通行；
  - `(3,15)` 向北至 `(3,14)` 可通行；
  - `(3,14)` terrain 是 `01h`。
- SearchLocation entry 1、terrain `01h`：
  - PICTURE block `72`；
  - 原始選項 `GREET／FLEE／ATTACK`；
  - `GREET` 記錄 Journal Entry 25；
  - `FLEE` 顯示「羊／食物」威嚇後幽魂淡去；
  - `ATTACK` 使幽魂消失。

以上 ECL 結果由原始 `curseoftheazurebonds.zip` 的 ECL6／GEO6 bytes 與
三分支 bounded session 測試交叉驗證。公開 Burial Glen 攻略亦把入口後
第一個事件標成 elfish spirit，並記載 `GREET` 取得 Journal 25 與關鍵詞
`Krrkik`：

- <https://www.gamebanshee.com/curseoftheazurebonds/walkthrough/burialglen.php>

使用者提供的 Adventurer’s Journal PDF 是 17 頁純掃描影像，沒有文字層；
本輪 OCR 無法可靠辨識第 25 條。手札英文另以公開 C64 文件轉錄交叉核對：

- <https://www.lemon64.com/doc/curse-of-the-azure-bonds/166>

## Remake 對應

- 有文字的 dungeon `NEWECL` 入口按 Continue 後，若
  `pendingDungeonEntry` 尚未完成，State 會同步 ECL map registers 並清除
  pending；不再沿用前一張地圖座標。
- 事件摘要、三個結果及 Journal 25 全文以穩定 `message_id` 放在 CoAB
  game pack。產品測試由受測 pack 查期望值，不複製繁中字串。
- `GREET` 是可重用 UI 動詞，只在 locale option adapter 加入「致意」；
  Burial Glen 劇情與 Journal 內容沒有寫入共用 engine。

## 驗證

- 原始 ECL oracle 分別跑 `GREET／FLEE／ATTACK`，驗證 PICTURE 72、原始
  選項與各分支固定英文。
- 正常 State 玩家路徑沿用第 384 輪全部操作，進入 block `0x40` 後確認
  `(2,15,E)`，依 GEO 可通行方向移動兩步到 `(3,14)`，選 `GREET`，最後
  驗證 JSON 訊息與 Journal 25 已加入遊戲內手札。

## 尚未完成

- Journal 25 所指紅網 terrain `82h`、`Krrkik`、蜘蛛與 rakshasa 後續。
- Burial Glen 其餘 terrain、隨機遭遇、休息安全條件、女王、東側遺跡出口。
- 本輪沒有原版同狀態 runtime 截圖；PICTURE 72 的版面仍須 DOSBox
  exact-state capture。
