# Gold Box 系列重用邊界

目的：讓下一款 Gold Box remake 能沿用方法與已證實 contract，同時不把 CoAB
常數、格式版本或劇情假裝成系列事實。這是路由文件，不複製逐輪數字。

## 可直接重用

| 類別 | 可重用內容 | 每款仍要驗證 |
|---|---|---|
| 工作法 | 原始 bytes／runtime → READY spec → typed 實作 → 玩家路徑 | 輸入雜湊、平台、位址空間 |
| 證據 | `exact`／`strong inference`／`hypothesis`／`unknown` 與非破壞性位址索引 | 每個語意的 producer、consumer、原始定位 |
| 驗證 | 決定性 RNG、parser invariant、save round-trip、截圖 metadata、抽樣 gate | 各作品高風險狀態與正常玩家路徑 |
| 低階能力 | storage、grid、bitmap font、locale overlay、標準像素／音訊介面 | 畫布、字型、色盤、資產授權 |

## 只能當候選

下列名稱看似共通，第二款作品驗證前不得宣稱同格式或直接搬用：

- ECL opcode 語意、operand 數量、work-memory 位址與 external routine。
- DAX／GEO block 編號、record stride、tile／wall selector 與座標方向。
- AD&D 戰鬥 quirk、法術表、怪物／物品記錄與 RNG 消耗順序。
- 角色、SAVGAM、效果 sidecar 與跨遊戲角色轉移格式。
- DOS／PC-98 音訊 driver、曲目表、播放時序與硬體近似。

升格條件：第二款合法持有的遊戲需完成 executable／toolchain、原始 bytes、consumer、
視覺或 runtime oracle、save round-trip，以及不含作品名稱分支的最小 adapter。

## 永遠留在作品層

- 劇情、人物、地名、選項、Journal、翻譯與攻略。
- 章節／block、旗標、座標、入口、遭遇編成、寶物與 NPC continuation。
- 原版資產索引、作品專屬 fallback、測試路線與發行授權清單。
- 未達第二作品驗證的欄位名、函式語意與格式推論。

## 新作品啟動清單

1. 建立原始檔 inventory、SHA-256、平台與合法來源，不先套 CoAB schema。
2. 選一條 source bytes → typed parser → rule → UI → save 的垂直切片。
3. 對同狀態原版畫面／行為，標出 exact、nearby 或 layout-only。
4. 逐項比對 engine contract；不合者先留作品 adapter，不為了共用而扭曲資料。
5. 第二 consumer 通過後才抽取，並在 engine 測試同時保留兩款正對照。

深層依據：[`engine-title-split.md`](engine-title-split.md)、
[`ssi-rpg-cross-project-lessons.md`](ssi-rpg-cross-project-lessons.md) 與獨立 engine repo。
