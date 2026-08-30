# 1248：六部位雙色戰鬥圖示與手繪 theme

狀態：`CONFORMED`
日期：2026-08-30

## 證據與勘誤

DOS `overlay-17:03EE3h`／PC-98 `overlay-17:045D0h` 的 `SETACTIVEICON`
已由 spec 1037 逐條閉合。角色記錄 `+145h..+14Ah` 是六個部位色 byte；每格低
nibble 是第一色、高 nibble 是第二色，兩色均在 `0..15` 環繞。部位順序是
Body、Arm、Leg、Hair／Face、Shield、Weapon。原版存檔樣本
`CHRDATB1.SAV`、`CHRDATC1.SAV`、`CHRDATA1.sav` 的固定值均為
`91 A2 B3 C4 E6 F7`，也就是模板 palette 對：`1/9、2/10、3/11、4/12、6/14、7/15`。

spec 1247 只接回頭、武器、體型，並明列六色為後續邊界。本規格補完該邊界；
不改寫 spec 1037 的原始定位與推論等級。

## Typed 契約

- `Character` 保存 `[6]uint8` 圖示色；全零的舊 JSON 存檔視為缺欄位，讀取時採
  原版預設 `91 A2 B3 C4 E6 F7`。
- DOS record codec 必須逐 byte round-trip `+145h..+14Ah`，不得碰未知 `+14Bh`。
- `Fighter` 攜帶同一組色值；建角預覽、加入隊伍、戰鬥、角色庫與正式存讀檔
  使用同一來源，不另設 UI-only 色表。
- 建角 icon 頁在 Head／Weapon／Size 後提供六部位各兩色，共十二個可編輯欄位；
  左右依原版在 `0..15` 環繞。Keep 保存，Exit 連同頭、武器、體型及六 bytes 全部還原。
- original theme 以原版模板 palette 對為遮罩，把命中的 palette index 精確換成
  玩家色值；透明與非模板色保持不變。
- modern-a6 手繪 theme 必須先以手繪 CHEAD／CBODY layer 組合所有可用頭×武器，
  再用相對應原版 layer 作語意遮罩。手繪像素保留明暗與 alpha，只把部位色相映射
  到玩家選色；不得用整張 sprite 單色 tint，也不得只支援少數預合成組合。
- theme 切換只改素材來源；角色六色值與目前建角游標不得改變。

## 驗收

1. 六 byte 預設、十二 nibble 正反環繞、Keep／Exit 還原均有 state 測試。
2. DOS codec 對 `+145h..+14Ah` 做非預設值 round-trip，`+14Bh` 保持原值。
3. Character → Fighter → renderer 色值不漂移，remake save round-trip 保留。
4. original sprite 的十二個模板 palette index 各自只受對應欄位控制。
5. modern-a6 任一非預合成頭×武器組合仍使用手繪 layers；改一個部位色只改該
   語意遮罩，透明、輪廓與其他部位保持不變。
6. 正常標題→建角→icon 頁修改顏色→F2 雙向切換→保存→加入→出發抽樣通過。

## 停止線

手繪遮罩是 `material-exact/layout-reconstructed`：部位身分取自 exact 原版 indexed
layer，手繪邊緣以最近的原版實心像素延伸；不宣稱手繪像素與 DOS 逐像素相同。

## 實作與驗證結果

- `Character.IconColors`、`Fighter.PartyIconColors` 與 DOS `+145h..+14Ah` 已接通；
  非預設六 byte round-trip 通過，`+14Bh` 正對照保持不變。
- 建角頁顯示六部位各兩色；十二 nibble 使用 `0..F` 環繞，Exit 整組還原。
- original renderer 逐一辨識十二個模板 palette index；modern-a6 以原版 layer 作
  語意 guide，保留手繪 alpha／明暗。像素測試證明改 Body 不污染 Arm 或輪廓。
- 刻意使用沒有預合成檔的 `head 01 × body 02`，成功由手繪 CHEAD／CBODY 組合並
  套色，證明新素材不受六張舊 composite 限制。
- 正常按鍵路徑在 icon 頁修改 Body 第一色為 `2`、切換 F2、保存、加入並出發；
  角色庫讀回仍為 `0x92`。正式抓圖
  `workplace/game-tester-20260830/creation-icon-six-colors-modern.png` 已人工檢查，
  十五欄、雙預覽與提示均未被框線遮擋。
- 完整回歸另撞到合法邊界：矮人基礎體質 18 經種族 `+1` 會成為 19，舊的最終
  驗證卻仍用擲點上限 18 而隨機擋住出發；現已把基礎擲點上限與種族調整後上限
  分開，19 正對照通過、20 負對照仍拒絕。
